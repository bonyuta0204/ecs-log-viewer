DIST     := bin
BINARIES := $(DIST)/ecs-log-viewer
SOURCES  := $(shell find . -name '*.go')

.PHONY: all build clean run lint help test test-coverage

help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -h "##" $(MAKEFILE_LIST) | grep -v grep | sed -e 's/\(.*\):.*##\(.*\)/  \1: \2/'

$(DIST):
	mkdir -p $(DIST)

$(BINARIES): $(DIST) $(SOURCES)
	go build -o $(BINARIES) ./cmd/ecs-log-viewer

build: $(BINARIES) ## Build the application

run: build ## Run the application
	$(BINARIES)

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf $(DIST) coverage.out coverage.html