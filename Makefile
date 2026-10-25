BIN := orders

.PHONY: build run test vet fmt lint

build:
	go build -o bin/$(BIN) ./cmd/orders

run:
	go run ./cmd/orders

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

lint:
	golangci-lint run