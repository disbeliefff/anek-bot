-- Create jokes table
CREATE TABLE IF NOT EXISTS jokes (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    source VARCHAR(50) NOT NULL,
    source_url TEXT,
    hash VARCHAR(64) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    used_count INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_jokes_source ON jokes(source);
CREATE INDEX IF NOT EXISTS idx_jokes_hash ON jokes(hash);
CREATE INDEX IF NOT EXISTS idx_jokes_created_at ON jokes(created_at);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_interaction TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_last_interaction ON users(last_interaction);
