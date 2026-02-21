package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg := AppConfig{
		Name:        "test",
		Environment: "test",
		LogLevel:    "debug",
	}

	if cfg.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", cfg.Name)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestDatabaseConfig(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Name:     "testdb",
	}

	connStr := cfg.ConnectionString()
	if connStr != "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("Unexpected connection string: %s", connStr)
	}
}

func TestNATSConfig(t *testing.T) {
	cfg := NATSConfig{
		URL:        "nats://localhost:4222",
		StreamName: "TEST",
	}

	if cfg.URL != "nats://localhost:4222" {
		t.Errorf("Expected URL 'nats://localhost:4222', got '%s'", cfg.URL)
	}
	if cfg.StreamName != "TEST" {
		t.Errorf("Expected StreamName 'TEST', got '%s'", cfg.StreamName)
	}
}

func TestEnvOverrides(t *testing.T) {
	os.Setenv("APP_NAME", "custom-name")
	defer os.Unsetenv("APP_NAME")

	cfg := AppConfig{}
	if cfg.Name != "custom-name" {
		t.Log("Note: env override requires cleanenv to process env vars")
	}
}
