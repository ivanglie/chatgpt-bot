FROM golang:1.24.0-alpine AS builder
WORKDIR /usr/src/chatgpt-bot
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -v -o chatgpt-bot ./cmd/app

FROM --platform=$BUILDPLATFORM alpine:3.17.0
WORKDIR /usr/local/bin/
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
COPY --from=builder /usr/src/chatgpt-bot/chatgpt-bot /usr/local/bin/chatgpt-bot
EXPOSE 18080
CMD ["chatgpt-bot"]