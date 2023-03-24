package main

import (
	"log"
	"os"
	"strings"

	"github.com/ivanglie/chatgpt-bot/internal/oai"
	"github.com/ivanglie/chatgpt-bot/internal/tg"
)

func main() {
	users, exists := os.LookupEnv("BOT_USERS")
	log.Printf("users: %s", users)

	client, err := oai.NewClient(os.Getenv("OPENAI_API_KEY"), 1000, "")
	if err != nil {
		log.Panic(err)
	}

	bot, err := tg.NewBotAPI(os.Getenv("BOT_TOKEN"), true, 0, 60)
	if err != nil {
		log.Panic(err)
	}

	updates := bot.GetUpdatesChan()

	for update := range updates {
		if update.Message == nil || update.Message.IsCommand() {
			continue
		}

		if exists && !strings.Contains(users, update.Message.Chat.UserName) {
			continue
		}

		res, err := client.Execute(update.Message.Text)
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}

		bot.Execute(update.Message.Chat.ID, res)
	}
}
