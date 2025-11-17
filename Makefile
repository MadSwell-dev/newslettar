.PHONY: build clean test run docker-build docker-run install deb help

# Build variables
BINARY_NAME=newslettar
VERSION=$(shell cat version.json | grep version | cut -d'"' -f4)
BUILD_DIR=build
DIST_DIR=dist
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Build flags
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"
BUILDFLAGS=-trimpath

help: ## Show this help message
	@echo "Newslettar Build System"
	@echo "======================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go mod tidy
	CGO_ENABLED=0 go build $(BUILDFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=linux GOARCH=arm go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm .
	GOOS=darwin GOARCH=amd64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(BUILDFLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Built all platforms in $(DIST_DIR)/"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(BINARY_NAME)
	@echo "✓ Cleaned"

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests complete"

coverage: test ## Show test coverage
	go tool cover -html=coverage.out

run: build ## Build and run locally
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) -web

install: build ## Install binary to /usr/local/bin
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installed"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "✓ Docker image built: $(BINARY_NAME):$(VERSION)"

docker-run: docker-build ## Build and run Docker container
	@echo "Starting Docker container..."
	docker run --rm -it -p 8080:8080 --name $(BINARY_NAME) $(BINARY_NAME):latest

docker-push: docker-build ## Build and push Docker image (requires Docker Hub login)
	@echo "Pushing Docker image..."
	docker tag $(BINARY_NAME):$(VERSION) agencefanfare/$(BINARY_NAME):$(VERSION)
	docker tag $(BINARY_NAME):$(VERSION) agencefanfare/$(BINARY_NAME):latest
	docker push agencefanfare/$(BINARY_NAME):$(VERSION)
	docker push agencefanfare/$(BINARY_NAME):latest
	@echo "✓ Pushed to Docker Hub"

deb: ## Build Debian package
	@echo "Building Debian package..."
	@./scripts/build-deb.sh
	@echo "✓ Debian package built"

fmt: ## Format Go code
	@echo "Formatting code..."
	gofmt -s -w $(GO_FILES)
	@echo "✓ Formatted"

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin" && exit 1)
	golangci-lint run ./...
	@echo "✓ Linting complete"

dev: ## Run in development mode with auto-reload (requires air)
	@which air > /dev/null || (echo "air not installed. Run: go install github.com/cosmtrek/air@latest" && exit 1)
	air

.DEFAULT_GOAL := help
