package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"anek-bot/internal/bot"
	"anek-bot/internal/config"
	"anek-bot/internal/database"
	"anek-bot/internal/models"
	"anek-bot/internal/parser"
	"anek-bot/internal/queue"
	"anek-bot/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, config.ErrEmptyBotToken) {
			fmt.Fprintln(os.Stderr, "Error: TELEGRAM_BOT_TOKEN environment variable is required")
		} else if errors.Is(err, config.ErrEmptyDBPassword) {
			fmt.Fprintln(os.Stderr, "Error: DB_PASSWORD environment variable is required")
		} else {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		}
		os.Exit(1)
	}

	logger.Init(cfg.App.LogLevel, nil)
	logger.Info("Starting anek-bot",
		logger.String("app", cfg.App.Name),
		logger.String("environment", cfg.App.Environment),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		var dbErr *database.ConnectionError
		if errors.As(err, &dbErr) {
			logger.Error("Failed to connect to database",
				logger.Err(dbErr),
				logger.String("host", cfg.Database.Host),
				logger.Int("port", cfg.Database.Port),
			)
		} else {
			logger.Error("Failed to connect to database",
				logger.Err(err),
			)
		}
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Connected to database")

	q, err := queue.New(cfg.NATS)
	if err != nil {
		logger.Error("Failed to connect to NATS", logger.Err(err))
		os.Exit(1)
	}
	defer q.Close()
	logger.Info("Connected to NATS", logger.String("url", cfg.NATS.URL))

	jokeRepo := database.NewJokeRepository(db)
	userRepo := database.NewUserRepository(db)

	go func() {
		logger.Info("Starting joke consumer...")
		if err := q.ConsumeJokes(ctx, func(joke *queue.JokeMessage) error {
			m := &models.Joke{
				Content:   joke.Content,
				Source:    string(joke.Source),
				SourceURL: joke.SourceURL,
				Hash:      joke.Hash,
			}
			if err := jokeRepo.Create(ctx, m); err != nil {
				logger.Error("Failed to save joke to database",
					logger.Err(err),
					logger.String("hash", joke.Hash),
				)
				return err
			}
			logger.Debug("Joke saved to database", logger.String("hash", joke.Hash))
			return nil
		}); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("Joke consumer error", logger.Err(err))
		}
	}()

	telegramBot, err := bot.New(cfg.Bot, jokeRepo, userRepo, q)
	if err != nil {
		logger.Error("Failed to create bot", logger.Err(err))
		os.Exit(1)
	}

	tbot, err := telegramBot.Start()
	if err != nil {
		logger.Error("Failed to start bot", logger.Err(err))
		os.Exit(1)
	}
	logger.Info("Telegram bot started")

	go func() {
		logger.Info("Starting parser...")
		p := parser.New(cfg.Parser, q)
		if err := p.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("Parser error", logger.Err(err))
		}
	}()

	healthMux := http.NewServeMux()
	healthMux.HandleFunc(cfg.Health.Endpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	healthServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Health.Port),
		Handler: healthMux,
	}

	go func() {
		logger.Info("Health server starting",
			logger.Int("port", cfg.Health.Port),
		)
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Health server error", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	tbot.Stop()

	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error shutting down health server", logger.Err(err))
	}

	logger.Info("Bot stopped gracefully")
}
