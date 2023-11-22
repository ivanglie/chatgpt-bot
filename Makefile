BOT_TOKEN=$(shell cat bottoken.secret)
OPENAI_API_KEY=$(shell cat openaiapikey.secret)
BOT_USERS=$(shell cat botusers.secret)

tests:
	go test -v -cover -race ./...

run:
	go run -race ./cmd/app/ --bottoken=$(BOT_TOKEN) --openaiapikey=$(OPENAI_API_KEY) --botusers=$(BOT_USERS)

dc:
	BOT_TOKEN=$(BOT_TOKEN) OPENAI_API_KEY=$(OPENAI_API_KEY) BOT_USERS=$(BOT_USERS) docker compose up -d