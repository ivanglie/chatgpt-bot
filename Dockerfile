FROM --platform=linux/amd64 golang:1.19-alpine AS builder
WORKDIR /usr/src/chatgpt-bot
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o chatgpt-bot ./cmd/app

FROM alpine:3.17.0
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
COPY --from=builder /usr/src/chatgpt-bot/chatgpt-bot /usr/local/bin/chatgpt-bot
CMD ["chatgpt-bot"]