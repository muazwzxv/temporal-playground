.PHONY: help up down restart logs clean ps build run migrate-up migrate-down migrate-status migrate-create docker-build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start all services
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@echo "Services started!"
	@echo "MySQL:     localhost:3306 (user: appuser, password: apppassword, db: user-management)"

	@echo "Redis:     localhost:6379"


	@echo "Kafka:     localhost:9092"
	@echo "Kafka UI:  http://localhost:8090"

	@echo "App:       http://localhost:8080"

down: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## Show logs from all services
	docker-compose logs -f

logs-app: ## Show application logs
	docker-compose logs -f app

logs-mysql: ## Show MySQL logs
	docker-compose logs -f mysql

clean: ## Stop services and remove volumes
	docker-compose down -v
	@echo "All services stopped and volumes removed"

ps: ## Show running services
	docker-compose ps

build: ## Build the Go application
	go build -o user-management ./cmd/server

run: ## Run the Go application locally
	go run ./cmd/server/main.go

migrate-up: ## Run all pending migrations
	goose -dir internal/database/migrations mysql "appuser:apppassword@tcp(localhost:3306)/user-management" up

migrate-down: ## Rollback last migration
	goose -dir internal/database/migrations mysql "appuser:apppassword@tcp(localhost:3306)/user-management" down

migrate-status: ## Show migration status
	goose -dir internal/database/migrations mysql "appuser:apppassword@tcp(localhost:3306)/user-management" status

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	goose -dir internal/database/migrations create $(NAME) sql

sqlc-generate: ## Generate Go code from SQL queries
	@echo "Generating sqlc code..."
	cd internal/database && sqlc generate
	@echo "sqlc code generation complete!"

docker-build: ## Build the Docker image
	docker build -t user-management:latest .

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	golangci-lint run

tidy: ## Tidy and verify dependencies
	go mod tidy
	go mod verify
