# Snapfig Makefile

# Variables
BINARY_NAME := snapfig
BUILD_DIR := .
GO := go
INSTALL_PATH := /usr/local/bin
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Build flags
LDFLAGS := -s -w

.PHONY: all build run install uninstall clean test test-verbose test-race coverage coverage-html lint lint-fix fmt vet check ci

# Default target
all: build

# Build the binary
build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run without building binary (go run)
dev:
	$(GO) run .

# Build and run
run: build
	./$(BINARY_NAME)

# Install to system
install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

# Uninstall from system
uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Removed $(INSTALL_PATH)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

# Run tests with coverage summary
test:
	@$(GO) test -coverprofile=$(COVERAGE_FILE) ./... > /dev/null 2>&1; \
	echo ""; \
	echo "Coverage by package:"; \
	echo "┌──────────────────────────────┬──────────┐"; \
	echo "│ Package                      │ Coverage │"; \
	echo "├──────────────────────────────┼──────────┤"; \
	$(GO) test -cover ./... 2>&1 | grep "^ok" | \
		sed 's|github.com/adrianpk/snapfig/||' | \
		awk '{pkg=$$2; for(i=1;i<=NF;i++) if($$i ~ /%$$/) cov=$$i; if(cov) printf "│ %-28s │ %8s │\n", pkg, cov}'; \
	echo "├──────────────────────────────┼──────────┤"; \
	$(GO) tool cover -func=$(COVERAGE_FILE) 2>/dev/null | grep "^total:" | awk '{printf "│ %-28s │ %8s │\n", "Total", $$3}'; \
	echo "└──────────────────────────────┴──────────┘"

# Run tests with verbose output
test-verbose:
	$(GO) test -v ./...

# Run tests with race detector
test-race:
	$(GO) test -race ./...

# Run tests with coverage
coverage:
	$(GO) test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

# Generate HTML coverage report
coverage-html: coverage
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run golangci-lint (uses PATH or ~/go/bin)
lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || ~/go/bin/golangci-lint run

# Run golangci-lint with auto-fix
lint-fix:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run --fix || ~/go/bin/golangci-lint run --fix

# Format code
fmt:
	$(GO) fmt ./...
	goimports -w -local github.com/adrianpk/snapfig .

# Vet code
vet:
	$(GO) vet ./...

# Check (fmt + vet + lint)
check: fmt vet lint

# CI target (lint + test + coverage)
ci: lint test coverage
