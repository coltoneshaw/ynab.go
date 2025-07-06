#!/usr/bin/env bash

.PHONY: lint test coverage help

GO_PACKAGES=$(shell go list ./...)
GO ?= $(shell command -v go 2> /dev/null)
export GOBIN ?= $(PWD)/bin
GOLANGCI_LINT_VERSION=v2.2.1

# Install go tools
install-go-tools:
	@mkdir -p $(GOBIN)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
	@GOBIN=$(GOBIN) $(GO) install mvdan.cc/gofumpt@latest

lint: install-go-tools
	$(GO) fmt ./...
	$(GO) vet ./...
	$(GOBIN)/golangci-lint run ./...

# Go validation (format, vet, lint, test, coverage)
go-check: install-go-tools
	$(GO) fmt ./...
	$(GO) vet ./...
	$(GOBIN)/golangci-lint run ./...
	$(GO) test ./...

test: ## Run unittests
	@go test -race -short ./...

coverage: ## Generate global code coverage report
	@./coverage.sh;

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
