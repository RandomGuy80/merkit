.PHONY: up down build run test kill

up:
	docker-compose up -d

down:
	docker-compose down

build:
	go build -o bin/server ./cmd/server

run: up
	go run ./cmd/server

test:
	go test ./... -v -count=1

kill:
	-fuser -k 8080/tcp 2>/dev/null || true

lint:
	go vet ./...
