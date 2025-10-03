# Simple Makefile for a Go project

# Load environment variables from .env file
include .env
export

# Database connection URL for migrations
DB_URL := postgresql://$(BLUEPRINT_DB_USERNAME):$(BLUEPRINT_DB_PASSWORD)@$(BLUEPRINT_DB_HOST):$(BLUEPRINT_DB_PORT)/$(BLUEPRINT_DB_DATABASE)?sslmode=disable&search_path=$(BLUEPRINT_DB_SCHEMA)

# Build the application
all: build test
templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi
tailwind-install:
	@if [ ! -f tailwindcss ]; then curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss; fi
	
	@chmod +x tailwindcss

build: tailwind-install templ-install
	@echo "Building..."
	@templ generate
	@./tailwindcss -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Migration commands
migrate-install:
	@if ! command -v migrate > /dev/null; then \
		read -p "golang-migrate is not installed. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
			if [ ! -x "$$(command -v migrate)" ]; then \
				echo "migrate installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install migrate. Exiting..."; \
			exit 1; \
		fi; \
	fi

migrate-create: migrate-install
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir internal/database/migrations -seq $$name

migrate-up: migrate-install
	@migrate -path internal/database/migrations -database "$(DB_URL)" up

migrate-down: migrate-install
	@migrate -path internal/database/migrations -database "$(DB_URL)" down 1

migrate-force: migrate-install
	@read -p "Enter version: " version; \
	migrate -path internal/database/migrations -database "$(DB_URL)" force $$version

# sqlc commands
sqlc-install:
	@if ! command -v sqlc > /dev/null; then \
		read -p "sqlc is not installed. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest; \
			if [ ! -x "$$(command -v sqlc)" ]; then \
				echo "sqlc installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install sqlc. Exiting..."; \
			exit 1; \
		fi; \
	fi

sqlc-generate: sqlc-install
	@sqlc generate

.PHONY: all build run test clean watch tailwind-install docker-run docker-down itest templ-install migrate-install migrate-create migrate-up migrate-down migrate-force sqlc-install sqlc-generate
