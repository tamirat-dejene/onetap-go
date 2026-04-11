.PHONY: run build test coverage lint tidy clean

## Run the development server
run:
	go run ./cmd/server/main.go

## Build production binary
build:
	go build -o bin/salary-advance ./cmd/server/main.go

## Run all unit tests
test:
	go test ./... -v

## Run tests with coverage report
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

## Tidy dependencies
tidy:
	go mod tidy

## Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...
