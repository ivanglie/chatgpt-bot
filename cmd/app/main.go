package main

import (
	"context"
	"log"
	"os"
	"strings"

	openai "github.com/ivanglie/chatgpt-bot/internal/openai"
	telegram "github.com/ivanglie/chatgpt-bot/internal/telegram"
)

func main() {
	users, exists := os.LookupEnv("BOT_USERS")
	log.Printf("users: %s", users)

	client, err := openai.NewClient(os.Getenv("OPENAI_API_KEY"), 1000, "")
	if err != nil {
		log.Panic(err)
	}

	bot, err := telegram.NewBotAPI(os.Getenv("BOT_TOKEN"), true, 0, 60)
	if err != nil {
		log.Panic(err)
	}

	ctx := context.Background()
	updates := bot.GetUpdatesChan()

	for {
		select {

		case <-ctx.Done():
			log.Printf("ctx.Done(): %v\n", ctx.Err())
			return

		case update, ok := <-updates:
			if !ok {
				log.Printf("updates channel closed")
				return
			}

			msg := update.Message

			if msg == nil || msg.IsCommand() {
				continue
			}

			if user := msg.Chat.UserName; exists && !strings.Contains(users, user) {
				log.Printf("User %s is not in BOT_USERS", user)
				continue
			}

			res, err := client.Execute(msg.Text, msg.Chat.IsGroup() || msg.Chat.IsSuperGroup())
			if err != nil {
				log.Printf("error: %v\n", err)
				continue
			}

			m := telegram.Message{ID: msg.MessageID,
				Chat: &telegram.Chat{ID: msg.Chat.ID, IsGroup: msg.Chat.IsGroup(), IsSuperGroup: msg.Chat.IsSuperGroup()}}
			bot.Execute(m, res)
		}
	}
}
