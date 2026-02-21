package main

import (
	"context"
	"fmt"
	"time"

	"anek-bot/internal/config"
	"anek-bot/internal/parser"
	"anek-bot/internal/queue"
	"anek-bot/pkg/logger"
)

type testQueue struct {
	jokes []string
}

func (t *testQueue) PublishJoke(ctx context.Context, joke *queue.JokeMessage) error {
	length := len(joke.Content)
	if length > 50 {
		length = 50
	}
	t.jokes = append(t.jokes, fmt.Sprintf("[%s] %s...", joke.Source, joke.Content[:length]))
	return nil
}

func main() {
	logger.Init("debug", nil)

	fmt.Println("=== Testing Parser ===")
	fmt.Println()

	cfg := config.ParserConfig{
		Enabled:      true,
		IntervalMins: 30 * time.Minute,
		Sources: config.SourcesConfig{
			Reddit: config.RedditConfig{
				Enabled:    true,
				Subreddits: []string{"Jokes"},
				Limit:      5,
			},
			Anekdot: config.AnekdotConfig{
				Enabled: true,
				Limit:   5,
			},
		},
	}

	tq := &testQueue{}

	p := parser.New(cfg, tq)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("Testing Reddit parsing...")
	if err := p.ParseAll(ctx); err != nil {
		logger.Error("Parse error", logger.Err(err))
	} else {
		fmt.Printf("✓ Reddit: Parsed %d jokes\n", len(tq.jokes))
		for i, joke := range tq.jokes {
			fmt.Printf("  %d: %s\n", i+1, joke)
		}
	}

	fmt.Println()
	fmt.Println("=== Test Complete ===")
}
