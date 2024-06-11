# Variables
GO_CMD=go
GO_MOD=$(GO_CMD) mod
GO_BUILD=$(GO_CMD) build
GO_RUN=$(GO_CMD) run
GO_TEST=$(GO_CMD) test
GO_FMT=$(GO_CMD) fmt
BINARY_NAME=crawler

# Default target
.PHONY: all
all: build

# Install dependencies
.PHONY: deps
deps:
	$(GO_MOD) tidy

# Format code
.PHONY: fmt
fmt:
	$(GO_FMT) ./...

# Run tests
.PHONY: test
test:
	$(GO_TEST) -v ./...

# Build the application
.PHONY: build
build: fmt
	$(GO_BUILD) -o $(BINARY_NAME) ./cmd

# Run the application
.PHONY: run
run: build
	./$(BINARY_NAME)

# Clean up build artifacts
.PHONY: clean
clean:
	@rm -f $(BINARY_NAME)
	@echo "Cleaned up build artifacts."

# Build and run the application
.PHONY: start
start: build run
