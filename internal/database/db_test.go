package database

import (
	"errors"
	"testing"

	"anek-bot/internal/models"
)

func TestConnectionError(t *testing.T) {
	baseErr := errors.New("connection refused")
	err := &ConnectionError{
		Host: "localhost",
		Port: 5432,
		Err:  baseErr,
	}

	if err.Error() == "" {
		t.Error("Expected error message")
	}

	if !errors.Is(err, baseErr) {
		t.Error("Expected underlying error to be unwrapped")
	}
}

func TestConnectionErrorMessage(t *testing.T) {
	baseErr := errors.New("connection refused")
	err := &ConnectionError{
		Host: "postgres.example.com",
		Port: 5432,
		Err:  baseErr,
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected error message")
	}
	if len(errMsg) < 10 {
		t.Errorf("Error() too short: %v", errMsg)
	}
}

func TestErrNoJokesFound(t *testing.T) {
	if !errors.Is(ErrNoJokesFound, ErrNoJokesFound) {
		t.Error("ErrNoJokesFound should match itself")
	}
}

func TestJokeModel(t *testing.T) {
	joke := models.Joke{
		ID:        1,
		Content:   "Why did the chicken cross the road?",
		Source:    "reddit",
		SourceURL: "https://reddit.com/r/Jokes/comments/abc123",
		Hash:      "def456",
		UsedCount: 5,
	}

	if joke.ID != 1 {
		t.Errorf("ID = %v, want 1", joke.ID)
	}
	if joke.Content == "" {
		t.Error("Content should not be empty")
	}
	if joke.Source != "reddit" {
		t.Errorf("Source = %v, want reddit", joke.Source)
	}
	if joke.UsedCount != 5 {
		t.Errorf("UsedCount = %v, want 5", joke.UsedCount)
	}
}

func TestUserModel(t *testing.T) {
	user := models.User{
		ID:         1,
		TelegramID: 123456789,
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
	}

	if user.ID != 1 {
		t.Errorf("ID = %v, want 1", user.ID)
	}
	if user.TelegramID != 123456789 {
		t.Errorf("TelegramID = %v, want 123456789", user.TelegramID)
	}
	if user.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", user.Username)
	}
}

func TestJokeSourceConstants(t *testing.T) {
	if models.SourceReddit != "reddit" {
		t.Errorf("SourceReddit = %v, want reddit", models.SourceReddit)
	}
	if models.SourceAnekdot != "anekdot" {
		t.Errorf("SourceAnekdot = %v, want anekdot", models.SourceAnekdot)
	}
}
