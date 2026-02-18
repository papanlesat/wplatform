.PHONY: help dev build run test clean install deps

help:
	@echo "Available commands:"
	@echo "  make dev       - Start development environment"
	@echo "  make build     - Build the backend"
	@echo "  make run       - Run the backend"
	@echo "  make test      - Run tests"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make install   - Install dependencies"
	@echo "  make deps      - Download dependencies"
	@echo "  make db-migrate- Apply database migrations"

dev:
	docker-compose up -d
	@echo "Development environment started. Traefik dashboard at http://localhost:8080"

build:
	cd backend && go build -o ../bin/api ./cmd/api

run:
	cd backend && go run ./cmd/api

test:
	cd backend && go test -v ./...

clean:
	rm -rf bin/
	rm -rf node_modules/
	rm -rf backend/vendor/

install:
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

deps:
	cd backend && go mod tidy
	cd backend && go mod vendor
	cd frontend && npm ci

db-migrate:
	docker exec wp-platform-db psql -U wpplatform -d wp_platform -f /docker-entrypoint-initdb.d/init.sql
