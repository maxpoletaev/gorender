.DEFAULT_GOAL := help

TEST_PACKAGE = ./...
PWD = $(shell pwd)
COMMIT_HASH = $(shell git rev-parse --short HEAD)

.PHONY: help
help: ## print help (this message)
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)## \(.*\)/\1;\3/p' \
	| column -t  -s ';'

.PHONY: build
build: ## build binary
	@echo "--------- running: $@ ---------"
	CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -o=gorender -pgo=default.pgo

.PHONY: build_noasm
build_noasm: ## build without assembly
	@echo "--------- running: $@ ---------"
	CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -o=gorender -tags=purego -pgo=default.pgo

.PHONY: build
build_debug: ## build with additional checks
	@echo "--------- running: $@ ---------"
	GOARCH=amd64 CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -o=gorender -pgo=default.pgo -gcflags="-m -d=ssa/check_bce" 2>&1 | tee build.log
	go-escape-lint -f build.log

PHONY: test
test: ## run tests
	@echo "--------- running: $@ ---------"
	@go test -v $(TEST_PACKAGE)

.PHONY: bench
bench: ## run benchmarks
	@echo "--------- running: $@ ---------"
	@GOGC=off GODEBUG=asyncpreemptoff=1 go test -bench=. -cpuprofile bench_cpuprof.pprof $(TEST_PACKAGE)
