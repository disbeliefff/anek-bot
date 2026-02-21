package queue

import (
	"encoding/json"
	"testing"

	"anek-bot/internal/models"
)

func TestJokeMessageJSON(t *testing.T) {
	joke := JokeMessage{
		Content:   "Why did the chicken cross the road?",
		Source:    models.SourceReddit,
		SourceURL: "https://reddit.com/r/Jokes/comments/123",
		Hash:      "abc123def456",
	}

	data, err := json.Marshal(joke)
	if err != nil {
		t.Fatalf("Failed to marshal JokeMessage: %v", err)
	}

	var parsed JokeMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JokeMessage: %v", err)
	}

	if parsed.Content != joke.Content {
		t.Errorf("Content = %v, want %v", parsed.Content, joke.Content)
	}
	if parsed.Source != joke.Source {
		t.Errorf("Source = %v, want %v", parsed.Source, joke.Source)
	}
	if parsed.SourceURL != joke.SourceURL {
		t.Errorf("SourceURL = %v, want %v", parsed.SourceURL, joke.SourceURL)
	}
	if parsed.Hash != joke.Hash {
		t.Errorf("Hash = %v, want %v", parsed.Hash, joke.Hash)
	}
}

func TestTelegramMessageJSON(t *testing.T) {
	msg := TelegramMessage{
		ChatID: 123456789,
		Text:   "Hello, world!",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal TelegramMessage: %v", err)
	}

	var parsed TelegramMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal TelegramMessage: %v", err)
	}

	if parsed.ChatID != msg.ChatID {
		t.Errorf("ChatID = %v, want %v", parsed.ChatID, msg.ChatID)
	}
	if parsed.Text != msg.Text {
		t.Errorf("Text = %v, want %v", parsed.Text, msg.Text)
	}
}

func TestJokeSourceValues(t *testing.T) {
	tests := []struct {
		source   models.JokeSource
		expected string
	}{
		{models.SourceReddit, "reddit"},
		{models.SourceAnekdot, "anekdot"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("Source = %v, want %v", tt.source, tt.expected)
			}
		})
	}
}
