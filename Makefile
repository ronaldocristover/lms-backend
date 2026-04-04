.PHONY: build run test clean migrate-up migrate-down

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin/

migrate-up:
	migrate -path ./migrations -database "postgres://app:secret@localhost:5432/lms?sslmode=disable" -up

migrate-down:
	migrate -path ./migrations -database "postgres://app:secret@localhost:5432/lms?sslmode=disable" -down

dev:
	godotenv -f .env go run ./cmd/server
