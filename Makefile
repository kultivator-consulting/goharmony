.PHONY: test test-verbose test-coverage bench lint fmt clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmarks"
	@echo "  lint          - Run golint and go vet"
	@echo "  fmt           - Format code with gofmt"
	@echo "  clean         - Clean build artifacts"
	@echo "  example-basic - Run basic example"
	@echo "  example-stream - Run streaming example"

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench:
	go test -bench=. -benchmem ./...

lint:
	@which golint > /dev/null || go install golang.org/x/lint/golint@latest
	golint ./...
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -f coverage.out coverage.html
	go clean

example-basic:
	go run examples/basic/main.go

example-stream:
	go run examples/streaming/main.go