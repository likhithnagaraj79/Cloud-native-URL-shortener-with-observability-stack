BINARY=url-shortener
MAIN=./cmd/server
DOCKER_COMPOSE=deployments/docker/docker-compose.yml

.PHONY: build run test lint tidy docker-up docker-down migrate-up migrate-down

build:
	go build -ldflags="-s -w" -o bin/$(BINARY) $(MAIN)

run:
	go run $(MAIN)

test:
	go test ./... -v -race -coverprofile=coverage.out

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-up:
	docker compose -f $(DOCKER_COMPOSE) up --build -d

docker-down:
	docker compose -f $(DOCKER_COMPOSE) down -v

migrate-up:
	migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSL_MODE" up

migrate-down:
	migrate -path migrations -database "postgres://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSL_MODE" down 1

logs:
	docker compose -f $(DOCKER_COMPOSE) logs -f app

ps:
	docker compose -f $(DOCKER_COMPOSE) ps
