package models

import "time"

type Joke struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	SourceURL string    `json:"source_url"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UsedCount int       `json:"used_count"`
}

type User struct {
	ID              int64     `json:"id"`
	TelegramID      int64     `json:"telegram_id"`
	Username        string    `json:"username"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	CreatedAt       time.Time `json:"created_at"`
	LastInteraction time.Time `json:"last_interaction"`
}

type JokeSource string

const (
	SourceReddit  JokeSource = "reddit"
	SourceAnekdot JokeSource = "anekdot"
)
