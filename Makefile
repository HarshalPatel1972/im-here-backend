.PHONY: run build build-linux deploy lint

run:
	go run ./cmd/scanner/main.go

build:
	go build -o bin/im-here ./cmd/scanner

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/im-here-linux ./cmd/scanner

deploy: build-linux
	./deploy.sh

lint:
	go vet ./...
