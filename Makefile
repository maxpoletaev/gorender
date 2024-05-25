.DEFAULT_GOAL := help

TEST_PACKAGE = ./...
PWD = $(shell pwd)
GO_MODULE = github.com/maxpoletaev/goxgl
COMMIT_HASH = $(shell git rev-parse --short HEAD)
C2GOASM_CLANG_FLAGS=-masm=intel -mno-red-zone -mstackrealign -mllvm -inline-threshold=1000 -fno-asynchronous-unwind-tables -fno-exceptions -fno-rtti

.PHONY: goat_docker_build
goat_docker_build: ## build goat docker image
	@echo "--------- running: $@ ---------"
	docker build -t goat -f goat.Dockerfile .

.PHONY: goat_docker_run
goat_docker_run: ## run goat
	@echo "--------- running: $@ ---------"
	docker run --rm -v $(PWD):/src goat make asm

.PHONY: asm
asm: ## generate assembly
	@echo "--------- running: $@ ---------"
	clang $(C2GOASM_CLANG_FLAGS) -O3 -mavx2 -S -o tmp.s matrix_avx256.c
	c2goasm -a tmp.s matrix_avx256.s

.PHONY: help
help: ## print help (this message)
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)## \(.*\)/\1;\3/p' \
	| column -t  -s ';'

.PHONY: build
build: ## build binary
	@echo "--------- running: $@ ---------"
	GOARCH=amd64 CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -o=goxgl -pgo=default.pgo

.PHONY: build
build_debug: ## build with additional checks
	@echo "--------- running: $@ ---------"
	GOARCH=amd64 CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -o=goxgl -pgo=default.pgo -gcflags="-m -d=ssa/check_bce" 2>&1 | tee build.log
	go-escape-lint -f build.log

PHONY: test
test: ## run tests
	@echo "--------- running: $@ ---------"
	@go test -v $(TEST_PACKAGE)


.PHONY: bench
bench: ## run benchmarks
	@echo "--------- running: $@ ---------"
	@GOGC=off GODEBUG=asyncpreemptoff=1 go test -bench=. -cpuprofile bench_cpuprof.pprof $(TEST_PACKAGE)