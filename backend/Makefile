.PHONY: build run test clean migrate-up migrate-down docker-up docker-down

# Build the application
build:
	go build -o bin/server cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -race -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go mod download

# Start docker services
docker-up:
	docker-compose up -d

# Stop docker services
docker-down:
	docker-compose down

# Run database migrations (requires golang-migrate)
migrate-up:
	migrate -path migrations -database "postgres://auction:auction123@localhost:5432/auction_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://auction:auction123@localhost:5432/auction_db?sslmode=disable" down

# Run migrations using psql
migrate-psql:
	psql -h localhost -U auction -d auction_db -f migrations/001_initial_schema.up.sql
	psql -h localhost -U auction -d auction_db -f migrations/002_seed_categories.up.sql

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate swagger docs (requires swag)
swagger:
	swag init -g cmd/server/main.go -o docs

# Development: run with hot reload (requires air)
dev:
	air

# All: format, lint, test, build
all: fmt lint test build
