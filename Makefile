.DEFAULT_GOAL := help

# go source files, ignore vendor directory
PKGS = $(shell go list ./... | grep -v /vendor)
COVERAGE_FILE = coverage.txt
COVERAGE_HTML_FILE = coverage.html

# Get first path on multiple GOPATH environments
GOPATH := $(shell echo ${GOPATH} | cut -d: -f1)

.PHONY: test
test: ## Run all the tests
	@echo "--> Running tests..."
	@go test -covermode=atomic -coverprofile=$(COVERAGE_FILE) -race -failfast -timeout=30s $(PKGS)

.PHONY: cover
cover: test ## Run all the tests and opens the coverage report
	@echo "--> Creating HTML coverage report at $(COVERAGE_HTML_FILE)..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML_FILE)

build: clean ## Build the app
	@echo "--> Building..."
	@go build ./...

.PHONY: clean
clean: ## Clean all built artifacts
	@echo "--> Cleaning all built artifacts..."
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML_FILE)
	@go clean
	@go mod tidy

BIN_DIR := $(GOPATH)/bin

GOLANGCI := $(BIN_DIR)/golangci-lint
GOLANGCI_VERSION := 1.30.0

$(GOLANGCI):
	@echo "--> Installing golangci v$(GOLANGCI_VERSION)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) v$(GOLANGCI_VERSION)

.PHONY: lint
lint: $(GOLANGCI) ## Run linter
	@echo "--> Running linter golangci v$(GOLANGCI_VERSION)..."
	@$(GOLANGCI) run

.PHONY: ci
ci: test lint ## Run all the tests and code checks

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
