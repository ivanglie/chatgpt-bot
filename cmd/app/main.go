package main

import (
	"log"
	"os"

	"github.com/ivanglie/chatgpt-bot/internal/oai"
	"github.com/ivanglie/chatgpt-bot/internal/tg"
)

func main() {
	client := oai.NewClient(os.Getenv("OPENAI_API_KEY"), 1000, "")

	bot, err := tg.NewBotAPI(os.Getenv("BOT_TOKEN"), true, 0, 60)
	if err != nil {
		log.Panic(err)
	}

	updates := bot.GetUpdatesChan()

	for update := range updates {
		if update.Message == nil || update.Message.IsCommand() {
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
