package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"anek-bot/internal/config"
	"anek-bot/internal/database"
	"anek-bot/internal/models"
	"anek-bot/internal/queue"
	"anek-bot/pkg/logger"

	"gopkg.in/telebot.v4"
)

var ErrRateLimited = errors.New("telegram rate limited")

type Bot struct {
	settings telebot.Settings
	jokeDB   *database.JokeRepository
	userDB   *database.UserRepository
	q        *queue.NATS
	tbot     *telebot.Bot
	cfg      config.BotConfig
}

func New(cfg config.BotConfig, jokeDB *database.JokeRepository, userDB *database.UserRepository, q *queue.NATS) (*Bot, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	return &Bot{
		cfg:    cfg,
		jokeDB: jokeDB,
		userDB: userDB,
		q:      q,
		settings: telebot.Settings{
			Token:  cfg.Token,
			Poller: &telebot.LongPoller{Timeout: 10},
		},
	}, nil
}

func (b *Bot) Start() (*telebot.Bot, error) {
	tbot, err := telebot.NewBot(b.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	b.tbot = tbot
	b.setupHandlers(tbot)

	go b.startTelegramConsumer(context.Background())

	go tbot.Start()

	return tbot, nil
}

func (b *Bot) setupHandlers(bot *telebot.Bot) {
	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		logger.Info("Incoming text message",
			logger.Int64("user_id", c.Sender().ID),
			logger.String("username", c.Sender().Username),
			logger.String("text", c.Text()),
		)
		return b.handleText(c)
	})

	bot.Handle(telebot.OnEdited, func(c telebot.Context) error {
		logger.Info("Incoming edited message",
			logger.Int64("user_id", c.Sender().ID),
			logger.String("username", c.Sender().Username),
		)
		return nil
	})

	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		logger.Info("Incoming callback",
			logger.Int64("user_id", c.Sender().ID),
			logger.String("callback_data", c.Callback().Data),
		)
		return nil
	})

	bot.Handle(telebot.OnChatJoinRequest, func(c telebot.Context) error {
		logger.Info("Incoming chat join request",
			logger.Int64("user_id", c.Sender().ID),
			logger.String("username", c.Sender().Username),
		)
		return nil
	})

	bot.Handle("/start", b.handleStart)
	bot.Handle("/joke", b.handleJoke)
	bot.Handle("/stats", b.handleStats)
	bot.Handle("/help", b.handleHelp)
}

func (b *Bot) startTelegramConsumer(ctx context.Context) {
	if b.q == nil {
		return
	}

	go func() {
		err := b.q.ConsumeTelegramMessages(ctx, func(msg *queue.TelegramMessage) error {
			return b.sendMessageWithRetry(msg.ChatID, msg.Text)
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("Telegram consumer error", logger.Err(err))
		}
	}()
}

func (b *Bot) sendMessageWithRetry(chatID int64, text string) error {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		_, err := b.tbot.Send(&telebot.Chat{ID: chatID}, text, &telebot.SendOptions{
			ParseMode: telebot.ModeMarkdown,
		})

		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "Too Many Requests") || strings.Contains(errStr, "rate") {
				logger.Warn("Rate limited, retrying...",
					logger.Int("retry", i+1),
					logger.Int("max_retries", maxRetries),
				)
				time.Sleep(retryDelay)
				retryDelay *= 2
				continue
			}
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	return ErrRateLimited
}

func (b *Bot) handleStart(c telebot.Context) error {
	user := &models.User{
		TelegramID: c.Sender().ID,
		Username:   c.Sender().Username,
		FirstName:  c.Sender().FirstName,
		LastName:   c.Sender().LastName,
	}

	ctx := context.Background()
	if err := b.userDB.Upsert(ctx, user); err != nil {
		logger.Error("Failed to save user", logger.Err(err))
	}

	welcome := "*Welcome to Anek Bot!*\n\n" +
		"I'll send you random jokes from Reddit and anekdot.ru.\n\n" +
		"Commands:\n" +
		"- /joke - Get a random joke\n" +
		"- /joke reddit - Get a joke from Reddit\n" +
		"- /joke anekdot - Get a joke from anekdot.ru\n" +
		"- /stats - Bot statistics\n" +
		"- /help - Show this help message"

	return b.queueOrSend(c.Sender().ID, welcome)
}

func (b *Bot) handleJoke(c telebot.Context) error {
	args := c.Args()
	var joke *models.Joke
	var err error
	ctx := context.Background()

	if len(args) > 0 {
		source := strings.ToLower(args[0])
		switch source {
		case "reddit":
			joke, err = b.jokeDB.GetRandomBySource(ctx, models.SourceReddit)
		case "anekdot":
			joke, err = b.jokeDB.GetRandomBySource(ctx, models.SourceAnekdot)
		default:
			return b.queueOrSend(c.Sender().ID, "Unknown source. Use: /joke, /joke reddit, or /joke anekdot")
		}
	} else {
		joke, err = b.jokeDB.GetRandom(ctx)
	}

	if err != nil {
		logger.Error("Failed to get joke", logger.Err(err))
		return b.queueOrSend(c.Sender().ID, "Sorry, no jokes available right now. Try again later!")
	}

	sourceLabel := "[anekdot]"
	if joke.Source == string(models.SourceReddit) {
		sourceLabel = "[reddit]"
	}

	msg := fmt.Sprintf("*Joke*\n\n%s\n\n%s", joke.Content, sourceLabel)

	return b.queueOrSend(c.Sender().ID, msg)
}

func (b *Bot) queueOrSend(chatID int64, text string) error {
	if b.q != nil {
		msg := &queue.TelegramMessage{
			ChatID: chatID,
			Text:   text,
		}
		if err := b.q.PublishTelegramMessage(context.Background(), msg); err != nil {
			logger.Error("Failed to queue telegram message", logger.Err(err))
		}
		return nil
	}

	_, err := b.tbot.Send(&telebot.Chat{ID: chatID}, text, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
	return err
}

func (b *Bot) handleStats(c telebot.Context) error {
	ctx := context.Background()
	totalJokes, err := b.jokeDB.Count(ctx)
	if err != nil {
		return b.queueOrSend(c.Sender().ID, "Failed to get statistics")
	}

	redditJokes, _ := b.jokeDB.CountBySource(ctx, models.SourceReddit)
	anekdotJokes, _ := b.jokeDB.CountBySource(ctx, models.SourceAnekdot)
	totalUsers, _ := b.userDB.Count(ctx)

	stats := fmt.Sprintf(
		"*Bot Statistics*\n\n"+
			"Total jokes: %d\n"+
			"Reddit jokes: %d\n"+
			"Anekdot jokes: %d\n"+
			"Total users: %d",
		totalJokes, redditJokes, anekdotJokes, totalUsers,
	)

	return b.queueOrSend(c.Sender().ID, stats)
}

func (b *Bot) handleHelp(c telebot.Context) error {
	help := "*Help*\n\n" +
		"Commands:\n" +
		"- /start - Start the bot\n" +
		"- /joke - Get a random joke\n" +
		"- /joke reddit - Get a joke from Reddit\n" +
		"- /joke anekdot - Get a joke from anekdot.ru\n" +
		"- /stats - Show bot statistics\n" +
		"- /help - Show this help message"

	return b.queueOrSend(c.Sender().ID, help)
}

func (b *Bot) handleText(c telebot.Context) error {
	return b.queueOrSend(c.Sender().ID, "Use /joke to get a joke!")
}
