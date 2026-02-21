package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"anek-bot/internal/config"
	"anek-bot/internal/models"
	"anek-bot/internal/queue"
	"anek-bot/pkg/logger"
)

type Queue interface {
	PublishJoke(ctx context.Context, joke *queue.JokeMessage) error
}

type Parser struct {
	cfg    config.ParserConfig
	client *http.Client
	q      Queue
}

func New(cfg config.ParserConfig, q Queue, opts ...Option) *Parser {
	p := &Parser{
		cfg: cfg,
		q:   q,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

type Option func(*Parser)

func WithHTTPClient(client *http.Client) Option {
	return func(p *Parser) {
		p.client = client
	}
}

type RedditPost struct {
	Data struct {
		Children []struct {
			Data struct {
				Title     string `json:"title"`
				Selftext  string `json:"selftext"`
				Permalink string `json:"permalink"`
				URL       string `json:"url"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func (p *Parser) Start(ctx context.Context) error {
	if !p.cfg.Enabled {
		return nil
	}

	logger.Info("Running initial parse...")
	if err := p.ParseAll(ctx); err != nil {
		logger.Error("Initial parse failed", logger.Err(err))
		return fmt.Errorf("initial parse failed: %w", err)
	}
	logger.Info("Initial parse completed")

	ticker := time.NewTicker(p.cfg.IntervalMins)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.ParseAll(ctx); err != nil {
				return fmt.Errorf("parse failed: %w", err)
			}
		}
	}
}

func (p *Parser) ParseAll(ctx context.Context) error {
	if p.cfg.Sources.Reddit.Enabled {
		if err := p.parseReddit(ctx); err != nil {
			return fmt.Errorf("reddit parsing failed: %w", err)
		}
	}

	if p.cfg.Sources.Anekdot.Enabled {
		if err := p.parseAnekdot(ctx); err != nil {
			return fmt.Errorf("anekdot parsing failed: %w", err)
		}
	}

	return nil
}

func (p *Parser) parseReddit(ctx context.Context) error {
	limit := p.cfg.Sources.Reddit.Limit

	for _, subreddit := range p.cfg.Sources.Reddit.Subreddits {
		logger.Info("Parsing subreddit", logger.String("subreddit", subreddit))
		url := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=%d", subreddit, limit)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			logger.Error("Failed to create request", logger.Err(err))
			return err
		}
		req.Header.Set("User-Agent", "anek-bot/1.0")

		resp, err := p.client.Do(req)
		if err != nil {
			logger.Error("Failed to fetch subreddit", logger.String("subreddit", subreddit), logger.Err(err))
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warn("Non-OK status from Reddit", logger.String("subreddit", subreddit), logger.Int("status", resp.StatusCode))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var posts RedditPost
		if err := json.Unmarshal(body, &posts); err != nil {
			return err
		}

		for _, child := range posts.Data.Children {
			post := child.Data
			content := post.Selftext
			if content == "" {
				continue
			}
			if len(content) > 3000 {
				content = content[:3000]
			}

			joke := &queue.JokeMessage{
				Content:   content,
				Source:    models.SourceReddit,
				SourceURL: "https://reddit.com" + post.Permalink,
				Hash:      generateHash(content),
			}

			if err := p.q.PublishJoke(ctx, joke); err != nil {
				logger.Error("Failed to publish joke to queue",
					logger.Err(err),
					logger.String("source", string(joke.Source)),
				)
				continue
			}
			logger.Info("Published joke to queue", logger.String("source", string(joke.Source)), logger.String("hash", joke.Hash))
		}
	}

	return nil
}

func (p *Parser) parseAnekdot(ctx context.Context) error {
	url := "https://anekdot.ru/random/anekdot/"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("anekdot.ru returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	anekdotPattern := regexp.MustCompile(`<div class="text">([\s\S]*?)</div>`)
	matches := anekdotPattern.FindAllStringSubmatch(string(body), -1)

	limit := p.cfg.Sources.Anekdot.Limit
	count := 0

	for _, match := range matches {
		if count >= limit {
			break
		}

		content := cleanHTML(match[1])
		content = strings.TrimSpace(content)

		if len(content) < 10 || len(content) > 3000 {
			continue
		}

		joke := &queue.JokeMessage{
			Content:   content,
			Source:    models.SourceAnekdot,
			SourceURL: "https://anekdot.ru",
			Hash:      generateHash(content),
		}

		if err := p.q.PublishJoke(ctx, joke); err != nil {
			logger.Error("Failed to publish joke to queue",
				logger.Err(err),
				logger.String("source", string(joke.Source)),
			)
			continue
		}
		count++
	}

	return nil
}

func generateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func cleanHTML(text string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	return text
}
