package database

import (
	"context"
	"errors"
	"fmt"

	"anek-bot/internal/config"
	"anek-bot/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoJokesFound = errors.New("no jokes found in database")
)

type ConnectionError struct {
	Host string
	Port int
	Err  error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("failed to connect to database at %s:%d: %v", e.Host, e.Err, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, &ConnectionError{
			Host: cfg.Host,
			Port: cfg.Port,
			Err:  err,
		}
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, &ConnectionError{
			Host: cfg.Host,
			Port: cfg.Port,
			Err:  err,
		}
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

type JokeRepository struct {
	db *DB
}

func NewJokeRepository(db *DB) *JokeRepository {
	return &JokeRepository{db: db}
}

func (r *JokeRepository) Create(ctx context.Context, joke *models.Joke) error {
	query := `
		INSERT INTO jokes (content, source, source_url, hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (hash) DO NOTHING
		RETURNING id, created_at
	`
	return r.db.Pool.QueryRow(ctx, query, joke.Content, joke.Source, joke.SourceURL, joke.Hash).Scan(&joke.ID, &joke.CreatedAt)
}

func (r *JokeRepository) GetRandom(ctx context.Context) (*models.Joke, error) {
	query := `
		UPDATE jokes
		SET used_count = used_count + 1
		WHERE id = (
			SELECT id FROM jokes
			ORDER BY RANDOM()
			LIMIT 1
		)
		RETURNING id, content, source, source_url, hash, created_at, used_count
	`
	var joke models.Joke
	err := r.db.Pool.QueryRow(ctx, query).Scan(
		&joke.ID, &joke.Content, &joke.Source,
		&joke.SourceURL, &joke.Hash, &joke.CreatedAt, &joke.UsedCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoJokesFound
		}
		return nil, err
	}
	return &joke, nil
}

func (r *JokeRepository) GetRandomBySource(ctx context.Context, source models.JokeSource) (*models.Joke, error) {
	query := `
		UPDATE jokes
		SET used_count = used_count + 1
		WHERE id = (
			SELECT id FROM jokes
			WHERE source = $1
			ORDER BY RANDOM()
			LIMIT 1
		)
		RETURNING id, content, source, source_url, hash, created_at, used_count
	`
	var joke models.Joke
	err := r.db.Pool.QueryRow(ctx, query, source).Scan(
		&joke.ID, &joke.Content, &joke.Source,
		&joke.SourceURL, &joke.Hash, &joke.CreatedAt, &joke.UsedCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoJokesFound
		}
		return nil, err
	}
	return &joke, nil
}

func (r *JokeRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM jokes").Scan(&count)
	return count, err
}

func (r *JokeRepository) CountBySource(ctx context.Context, source models.JokeSource) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM jokes WHERE source = $1", source).Scan(&count)
	return count, err
}

func (r *JokeRepository) HashExists(ctx context.Context, hash string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM jokes WHERE hash = $1)", hash).Scan(&exists)
	return exists, err
}

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Upsert(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			last_interaction = CURRENT_TIMESTAMP
		RETURNING id, created_at
	`
	return r.db.Pool.QueryRow(ctx, query,
		user.TelegramID, user.Username, user.FirstName, user.LastName,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
