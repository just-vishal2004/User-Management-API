# ─── Variables ────────────────────────────────────────────────────────────────
APP_NAME    := ainyx-backend
BUILD_DIR   := bin
MAIN_PATH   := cmd/server/main.go
DB_URL      := postgres://vishalkumar@localhost:5432/ainyx_users?sslmode=disable
MIGRATE_DIR := db/migrations

# ─── Build ────────────────────────────────────────────────────────────────────

## build: compile the application binary into bin/
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Binary created at $(BUILD_DIR)/$(APP_NAME)"

## run: run the application directly with go run
run:
	go run $(MAIN_PATH)

## dev: run with air for live reload during development
dev:
	air

# ─── Database ─────────────────────────────────────────────────────────────────

## migrate-up: apply all pending migrations
migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(DB_URL)" up

## migrate-down: roll back the last migration
migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(DB_URL)" down 1

## migrate-create name=<migration_name>: create a new migration file
migrate-create:
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $(name)

# ─── SQLC ─────────────────────────────────────────────────────────────────────

## sqlc: regenerate database code from SQL queries
sqlc:
	cd db/sqlc && sqlc generate && cd ../..

# ─── Testing ──────────────────────────────────────────────────────────────────

## test: run all tests
test:
	go test ./... -v

## test-cover: run all tests with coverage report
test-cover:
	go test ./... -v -cover

# ─── Code Quality ─────────────────────────────────────────────────────────────

## lint: run go vet for static analysis
lint:
	go vet ./...

## tidy: tidy go modules
tidy:
	go mod tidy

# ─── Docker ───────────────────────────────────────────────────────────────────

## docker-build: build the Docker image
docker-build:
	docker build -t $(APP_NAME) .

## docker-run: run the application in Docker
docker-run:
	docker run --env-file .env -p 3000:3000 $(APP_NAME)

## docker-down: stop and remove containers
docker-down:
	docker-compose down

## docker-up: start all services with docker-compose
docker-up:
	docker-compose up --build

# ─── Cleanup ──────────────────────────────────────────────────────────────────

## clean: remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) tmp/
	@echo "Done."

# ─── Help ─────────────────────────────────────────────────────────────────────

## help: display all available make commands
help:
	@echo "Available commands:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

.PHONY: build run dev migrate-up migrate-down migrate-create sqlc test test-cover lint tidy docker-build docker-run docker-down docker-up clean help