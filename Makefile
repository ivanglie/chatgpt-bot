include .env

.PHONY: run tests docker

run:
	BOT_TOKEN=$(BOT_TOKEN) OPENAI_API_KEY=$(OPENAI_API_KEY) BOT_USERS=$(BOT_USERS) go run -race ./cmd/app/

tests:
	go test -v -cover -race ./...

docker:
	BOT_TOKEN=$(BOT_TOKEN) OPENAI_API_KEY=$(OPENAI_API_KEY) BOT_USERS=$(BOT_USERS) docker compose down -v && docker compose up --build -d