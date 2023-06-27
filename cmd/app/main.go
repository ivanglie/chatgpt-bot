package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/ivanglie/chatgpt-bot/internal/oai"
	"github.com/ivanglie/chatgpt-bot/internal/tg"
)

func main() {
	users, exists := os.LookupEnv("BOT_USERS")
	log.Printf("users: %s\n", users)

	openAI, err := oai.New(os.Getenv("OPENAI_API_KEY"), 1000, "")
	if err != nil {
		log.Panic(err)
	}

	tBot, err := tg.New(os.Getenv("BOT_TOKEN"), true, 0, 60)
	if err != nil {
		log.Panic(err)
	}

	updates := tBot.GetUpdatesChan()

	for update := range updates {
		if update.Message == nil || update.Message.IsCommand() {
			continue
		}

		if u := update.Message.Chat.UserName; len(u) == 0 || (exists && !strings.Contains(users, u)) {
			log.Printf("error: %v\n", errors.New("user is not allowed"))
			continue
		}

		res, err := openAI.Generate(update.Message.Text)
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}

		tBot.Send(update.Message.Chat.ID, res)
	}
}
