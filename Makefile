APP_NAME := app
BINARY   := ./$(APP_NAME)
MAIN     := main.go

.PHONY: run build test lint clean tidy

## Run the app in development (uses SQLite, no .env required)
run:
	go run $(MAIN)

## Build a production binary
build:
	go build -ldflags="-s -w" -o $(BINARY) $(MAIN)

## Run all tests
test:
	go test ./tests/... -v

## Run unit tests only
test-unit:
	go test ./tests/unit/... -v

## Run integration tests only
test-integration:
	go test ./tests/integration/... -v

## Tidy and vendor go modules
tidy:
	go mod tidy

## Run go vet + staticcheck
lint:
	go vet ./...

## Delete build artifacts
clean:
	rm -f $(BINARY)

## Show this help
help:
	@grep -E '^##' Makefile | sed 's/## //'
