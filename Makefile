# Makefile for reverseProxyExample

BINARY=./bin/reverseproxy
CMD=./cmd/reverseproxy

.PHONY: build run test fmt vet clean help

build:
	@mkdir -p ./bin
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

run-bin: build
	$(BINARY)

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf ./bin

help:
	@echo "Available targets: build run run-bin test fmt vet clean"
