package parser

import (
	"testing"

	"anek-bot/internal/models"
)

func TestGenerateHash(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantLen int
	}{
		{"simple joke", "Why did the chicken cross the road?", 64},
		{"empty string", "", 64},
		{"long content", string(make([]byte, 1000)), 64},
		{"unicode", "Почему программист ушел с работы? Потому что он не получил arrays (a raise)", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateHash(tt.content)
			if len(got) != tt.wantLen {
				t.Errorf("generateHash() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple tags",
			input:    "<p>Hello</p>",
			expected: "Hello",
		},
		{
			name:     "complex tags",
			input:    "<div class=\"text\"><b>Bold</b> and <i>italic</i></div>",
			expected: "Bold and italic",
		},
		{
			name:     "nbsp entity",
			input:    "Hello&nbsp;World",
			expected: "Hello World",
		},
		{
			name:     "quot entity",
			input:    "&quot;quoted&quot;",
			expected: "\"quoted\"",
		},
		{
			name:     "amp entity",
			input:    "A &amp; B",
			expected: "A & B",
		},
		{
			name:     "lt gt entities",
			input:    "&lt;div&gt;",
			expected: "<div>",
		},
		{
			name:     "no tags",
			input:    "Just plain text",
			expected: "Just plain text",
		},
		{
			name:     "mixed content",
			input:    "<div class=\"text\">Привет &amp; мир!</div>",
			expected: "Привет & мир!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanHTML(tt.input)
			if got != tt.expected {
				t.Errorf("cleanHTML() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestJokeSource(t *testing.T) {
	tests := []struct {
		name     string
		source   models.JokeSource
		expected string
	}{
		{"reddit source", models.SourceReddit, "reddit"},
		{"anekdot source", models.SourceAnekdot, "anekdot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("JokeSource = %v, want %v", tt.source, tt.expected)
			}
		})
	}
}
