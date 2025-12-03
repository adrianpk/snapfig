# Snapfig Makefile

# Variables
BINARY_NAME := snapfig
BUILD_DIR := .
GO := go
INSTALL_PATH := /usr/local/bin

# Build flags
LDFLAGS := -s -w

.PHONY: all build run install uninstall clean test fmt vet

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

# Run tests
test:
	$(GO) test -v ./...

# Format code
fmt:
	$(GO) fmt ./...

# Vet code
vet:
	$(GO) vet ./...

# Check (fmt + vet)
check: fmt vet
