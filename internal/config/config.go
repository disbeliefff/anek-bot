package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ErrEmptyBotToken   = errors.New("telegram bot token is required")
	ErrEmptyDBPassword = errors.New("database password is required")
)

type Config struct {
	App      AppConfig      `yaml:"app" env:"APP"`
	Database DatabaseConfig `yaml:"database" env:"DB"`
	Bot      BotConfig      `yaml:"bot" env:"BOT"`
	Parser   ParserConfig   `yaml:"parser" env:"PARSER"`
	NATS     NATSConfig     `yaml:"nats" env:"NATS"`
	Health   HealthConfig   `yaml:"health" env:"HEALTH"`
}

type AppConfig struct {
	Name        string `yaml:"name" env:"NAME" env-default:"anek-bot"`
	Environment string `yaml:"environment" env:"ENVIRONMENT" env-default:"production"`
	LogLevel    string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
}

type DatabaseConfig struct {
	Host           string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port           int    `yaml:"port" env:"PORT" env-default:"5432"`
	User           string `yaml:"user" env:"USER" env-default:"anekbot"`
	Password       string `yaml:"password" env:"PASSWORD"`
	Name           string `yaml:"name" env:"NAME" env-default:"anekbot"`
	MaxConnections int    `yaml:"max_connections" env:"MAX_CONNECTIONS" env-default:"25"`
	MinConnections int    `yaml:"min_connections" env:"MIN_CONNECTIONS" env-default:"5"`
}

func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

type BotConfig struct {
	Token     string `yaml:"token" env:"TOKEN"`
	ParseMode string `yaml:"parse_mode" env:"PARSE_MODE" env-default:"Markdown"`
}

type ParserConfig struct {
	Enabled      bool          `yaml:"enabled" env:"ENABLED" env-default:"true"`
	IntervalMins time.Duration `yaml:"interval_minutes" env:"INTERVAL_MINUTES" env-default:"30m"`
	Sources      SourcesConfig `yaml:"sources" env:"PARSER_SOURCES"`
}

type SourcesConfig struct {
	Reddit  RedditConfig  `yaml:"reddit" env:"PARSER_REDDIT"`
	Anekdot AnekdotConfig `yaml:"anekdot" env:"PARSER_ANEKDOT"`
}

type RedditConfig struct {
	Enabled    bool     `yaml:"enabled" env:"ENABLED" env-default:"true"`
	Subreddits []string `yaml:"subreddits" env:"SUBREDDITS" env-separator:","`
	Limit      int      `yaml:"limit" env:"LIMIT" env-default:"25"`
}

type AnekdotConfig struct {
	Enabled bool `yaml:"enabled" env:"ENABLED" env-default:"true"`
	Limit   int  `yaml:"limit" env:"LIMIT" env-default:"20"`
}

type HealthConfig struct {
	Port     int    `yaml:"port" env:"PORT" env-default:"8080"`
	Endpoint string `yaml:"endpoint" env:"ENDPOINT" env-default:"/healthz"`
}

type NATSConfig struct {
	URL        string `yaml:"url" env:"URL" env-default:"nats://localhost:4222"`
	StreamName string `yaml:"stream_name" env:"STREAM_NAME" env-default:"ANEK"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.prod.yaml"
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from %s: %w", configPath, err)
	}

	cleanenv.ReadEnv(&cfg)

	if cfg.Bot.Token == "" {
		return nil, ErrEmptyBotToken
	}

	if cfg.Database.Password == "" {
		return nil, ErrEmptyDBPassword
	}

	return &cfg, nil
}
