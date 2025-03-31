lint:
	@golangci-lint run

build:
	@go build -o bin/bot cmd/bot/main.go

run:
	@go run cmd/bot/main.go

bot-up:
	docker compose up -d

bot-down:
	docker compose down