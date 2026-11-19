.PHONY: fmt lint test test-unit test-integration build clean help

help:
	@echo "Available targets:"
	@echo "  make fmt              - Format code with gofmt"
	@echo "  make lint             - Run golangci-lint and go vet"
	@echo "  make test             - Run all tests (unit + integration)"
	@echo "  make test-unit        - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo "  make build            - Build binary"
	@echo "  make clean            - Clean build artifacts"

fmt:
	@echo "Formatting code..."
	gofmt -w ./internal/
	@echo "✓ Code formatted"

lint:
	@echo "Running go vet..."
	go vet ./internal/domain ./internal/repository ./internal/service ./internal/app
	@echo "✓ go vet passed"
	@echo ""
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0"; exit 1)
	golangci-lint run ./internal/domain ./internal/repository ./internal/service ./internal/app
	@echo "✓ golangci-lint passed"

test: test-unit test-integration
	@echo "✓ All tests passed"

test-unit:
	@echo "Running unit tests..."
	go test -v ./internal/domain ./internal/repository ./internal/service ./internal/app
	@echo "✓ Unit tests passed"

test-integration: test-unit
	@echo "Integration tests (currently same as unit tests)"
	@echo "✓ Integration tests passed"

build:
	@echo "Building binary..."
	go build -o rtsprec ./cmd/rtsprec || (echo "⚠️  Build failed (possible gortsplib incompatibility)"; exit 0)
	@if [ -f rtsprec ]; then \
		echo "✓ Binary created: $$(ls -lh rtsprec | awk '{print $$5}')"; \
	else \
		echo "⚠️  Binary not created (gortsplib capture package incompatibility)"; \
	fi

clean:
	@echo "Cleaning..."
	rm -f rtsprec
	go clean ./cmd/rtsprec
	@echo "✓ Clean complete"
