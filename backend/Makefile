# Makefile - Simplified development commands

.PHONY: help build run test sync clean up down logs rebuild restart swagger

## Show help
help:
	@echo "Available commands:"
	@echo "  make build     - Build binary"
	@echo "  make run       - Run API locally (port 8080)"
	@echo "  make test      - Run tests"
	@echo "  make sync      - Sync providers and calculate scores"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make up        - Start all services (docker-compose)"
	@echo "  make down      - Stop all services"
	@echo "  make logs      - View logs (all services)"
	@echo "  make rebuild   - Rebuild and restart services"
	@echo "  make restart   - Restart backend service"
	@echo "  make swagger   - Generate API documentation"

## Build binary
build:
	go build -o bin/api ./cmd/api

## Run API locally
run:
	go run ./cmd/api

## Run tests
test:
	go test ./...

## Sync providers and calculate scores
sync:
	go run ./cmd/sync

## Seed test data for pagination testing
seed:
	go run ./cmd/seed

## Clean build artifacts
clean:
	rm -rf bin/ coverage.out api tmp/

## Start all services (docker-compose)
up:
	cd .. && docker-compose up -d

## Stop all services
down:
	cd .. && docker-compose down

## View logs
logs:
	cd .. && docker-compose logs -f

## Rebuild and restart services
rebuild:
	cd .. && docker-compose up -d --build

## Restart backend service only
restart:
	cd .. && docker-compose restart backend

## Generate Swagger API documentation
swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs
