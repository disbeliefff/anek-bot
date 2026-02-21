.PHONY: help build test run clean migrate start stop logs

help:
	@echo "Anek Bot - Makefile Commands"
	@echo ""
	@echo "  make build      - Build Docker image"
	@echo "  make up         - Start postgres and nats"
	@echo "  make down       - Stop all services"
	@echo "  make migrate    - Run database migrations"
	@echo "  make start      - Start bot (runs up + migrate + bot)"
	@echo "  make logs       - View bot logs"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean up containers and volumes"

build:
	docker build -t anek-bot:latest .

up:
	docker compose up -d postgres nats

down:
	docker compose down

migrate:
	docker compose --profile migrate build --no-cache
	docker compose --profile migrate run migrate
	docker compose --profile migrate run nats-init

logs:
	docker compose logs -f bot

test:
	go test ./...

clean:
	docker compose down -v
	docker rmi anek-bot:latest 2>/dev/null || true

start: up migrate
	@echo "Waiting for services..."
	@sleep 5
	docker compose up -d bot
