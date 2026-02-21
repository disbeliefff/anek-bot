package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"anek-bot/internal/config"
	"anek-bot/internal/models"
	"anek-bot/pkg/logger"

	"github.com/nats-io/nats.go"
)

const (
	JokeSubject     = "jokes.new"
	TelegramSubject = "telegram.send"
	ConsumerGroup   = "anek-bot"
)

type NATS struct {
	conn      *nats.Conn
	jetstream nats.JetStream
	cfg       config.NATSConfig
}

func New(cfg config.NATSConfig) (*NATS, error) {
	conn, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get JetStream: %w", err)
	}

	n := &NATS{
		conn:      conn,
		jetstream: js,
		cfg:       cfg,
	}

	return n, nil
}

func (n *NATS) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

type JokeMessage struct {
	Content   string            `json:"content"`
	Source    models.JokeSource `json:"source"`
	SourceURL string            `json:"source_url"`
	Hash      string            `json:"hash"`
}

func (n *NATS) PublishJoke(ctx context.Context, joke *JokeMessage) error {
	data, err := json.Marshal(joke)
	if err != nil {
		return fmt.Errorf("failed to marshal joke: %w", err)
	}

	_, err = n.jetstream.Publish(JokeSubject, data, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to publish joke: %w", err)
	}

	logger.Debug("Joke published to queue",
		logger.String("source", string(joke.Source)),
		logger.String("hash", joke.Hash),
	)

	return nil
}

type TelegramMessage struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func (n *NATS) PublishTelegramMessage(ctx context.Context, msg *TelegramMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	_, err = n.jetstream.Publish(TelegramSubject, data, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to publish telegram message: %w", err)
	}

	logger.Debug("Telegram message published to queue",
		logger.Any("chat_id", msg.ChatID),
	)

	return nil
}

func (n *NATS) ConsumeJokes(ctx context.Context, handler func(*JokeMessage) error) error {
	sub, err := n.jetstream.PullSubscribe(
		JokeSubject,
		ConsumerGroup,
		nats.BindStream(n.cfg.StreamName),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to jokes: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(500))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				return fmt.Errorf("failed to fetch messages: %w", err)
			}

			for _, msg := range msgs {
				var joke JokeMessage
				if err := json.Unmarshal(msg.Data, &joke); err != nil {
					logger.Error("Failed to unmarshal joke message",
						logger.Err(err),
					)
					msg.Nak()
					continue
				}

				if err := handler(&joke); err != nil {
					logger.Error("Failed to process joke",
						logger.Err(err),
					)
					msg.Nak()
					continue
				}

				msg.Ack()
			}
		}
	}
}

func (n *NATS) ConsumeTelegramMessages(ctx context.Context, handler func(*TelegramMessage) error) error {
	sub, err := n.jetstream.PullSubscribe(
		TelegramSubject,
		ConsumerGroup,
		nats.BindStream(n.cfg.StreamName),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to telegram: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(500))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				return fmt.Errorf("failed to fetch messages: %w", err)
			}

			for _, msg := range msgs {
				var telegramMsg TelegramMessage
				if err := json.Unmarshal(msg.Data, &telegramMsg); err != nil {
					logger.Error("Failed to unmarshal telegram message",
						logger.Err(err),
					)
					msg.Nak()
					continue
				}

				if err := handler(&telegramMsg); err != nil {
					logger.Error("Failed to send telegram message",
						logger.Err(err),
					)
					msg.Nak()
					continue
				}

				msg.Ack()
			}
		}
	}
}
