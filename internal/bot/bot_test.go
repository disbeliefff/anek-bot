package bot

import (
	"testing"

	"anek-bot/internal/config"
	"anek-bot/internal/database"
	"anek-bot/internal/models"
	"anek-bot/internal/queue"
)

func TestNewBot(t *testing.T) {
	cfg := config.BotConfig{
		Token:     "test-token",
		ParseMode: "Markdown",
	}

	_, err := New(cfg, nil, nil, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestNewBotNoToken(t *testing.T) {
	cfg := config.BotConfig{
		Token:     "",
		ParseMode: "Markdown",
	}

	_, err := New(cfg, nil, nil, nil)
	if err == nil {
		t.Error("Expected error when token is empty")
	}
}

func TestJokeRepository(t *testing.T) {
	_ = models.SourceReddit
	_ = models.SourceAnekdot
	_ = queue.JokeMessage{}
	_ = database.ErrNoJokesFound
}
