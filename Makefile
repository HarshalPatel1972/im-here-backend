.PHONY: run build lint

run:
	go run ./cmd/scanner/main.go

build:
	go build -o bin/im-here ./cmd/scanner

lint:
	go vet ./...
