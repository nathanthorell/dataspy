.PHONY: setup lint test coverage build clean all

# Default target
all: lint test build

# Install development tools
setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod tidy

# Run linting
lint:
	golangci-lint run ./...

# Run tests
test:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build the binary
build:
	go build -o build/dataspy

# Clean build artifacts
clean:
	rm -rf build/
	rm -f coverage.out coverage.html

# Run the application
run: build
	./build/dataspy
