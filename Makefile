.DEFAULT_GOAL := run

.PHONY: run
run: build
	bin/gokoban

.PHONY: build
build: gofmt
	go build -o bin/gokoban

.PHONY: gofmt
gofmt:
	GO111MODULE=off go fmt ./...
