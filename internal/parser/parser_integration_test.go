package parser

import (
	"encoding/json"
	"testing"

	"anek-bot/internal/models"
	"anek-bot/internal/queue"
)

func TestRedditPostParsing(t *testing.T) {
	jsonData := `{
		"data": {
			"children": [
				{
					"data": {
						"title": "Why did the chicken?",
						"selftext": "To get to the other side!",
						"permalink": "/r/Jokes/comments/abc123",
						"url": "https://reddit.com"
					}
				},
				{
					"data": {
						"title": "Funny joke",
						"selftext": "This is a test joke content",
						"permalink": "/r/Jokes/comments/def456",
						"url": "https://reddit.com"
					}
				}
			]
		}
	}`

	var posts RedditPost
	err := json.Unmarshal([]byte(jsonData), &posts)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(posts.Data.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(posts.Data.Children))
	}

	if posts.Data.Children[0].Data.Selftext != "To get to the other side!" {
		t.Errorf("Expected 'To get to the other side!', got '%s'", posts.Data.Children[0].Data.Selftext)
	}

	if posts.Data.Children[0].Data.Permalink != "/r/Jokes/comments/abc123" {
		t.Errorf("Expected '/r/Jokes/comments/abc123', got '%s'", posts.Data.Children[0].Data.Permalink)
	}
}

func TestJokeMessageContent(t *testing.T) {
	joke := &queue.JokeMessage{
		Content:   "Test joke content",
		Source:    models.SourceReddit,
		SourceURL: "https://reddit.com/r/Jokes/comments/abc",
		Hash:      generateHash("Test joke content"),
	}

	if joke.Content == "" {
		t.Error("Content should not be empty")
	}

	if joke.Source != models.SourceReddit {
		t.Errorf("Expected source %v, got %v", models.SourceReddit, joke.Source)
	}

	if len(joke.Hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(joke.Hash))
	}
}
