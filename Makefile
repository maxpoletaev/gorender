.DEFAULT_GOAL := help

TEST_PACKAGE = ./...
PWD = $(shell pwd)
GO_MODULE = github.com/maxpoletaev/goxgl
COMMIT_HASH = $(shell git rev-parse --short HEAD)

.PHONY: help
help: ## print help (this message)
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)## \(.*\)/\1;\3/p' \
	| column -t  -s ';'

.PHONY: build
build: ## build dendy
	@echo "--------- running: $@ ---------"
	CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -o=goxgl .

PHONY: test
test: ## run tests
	@echo "--------- running: $@ ---------"
	@go test -v $(TEST_PACKAGE)
