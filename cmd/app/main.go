package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ivanglie/chatgpt-bot/internal/bot"
	"github.com/ivanglie/chatgpt-bot/internal/process"
)

var revision = "local"

func main() {
	ctx := context.TODO()

	fmt.Printf("chatgpt-bot, %s\n", revision)

	setupLog(true)

	tbAPI, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("[ERROR] can't make telegram bot, %v", err)
	}

	tbAPI.Debug = true

	openAIBot, err := bot.NewOpenAI(os.Getenv("OPENAI_API_KEY"), 1000, "", &http.Client{Timeout: 120 * time.Second})
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	bots := bot.BotSlice{
		openAIBot,
	}

	tgListener := process.TelegramListener{
		TbAPI:        tbAPI,
		Bots:         bots,
		Debug:        true,
		IdleDuration: 30 * time.Second,
	}

	if users, exists := os.LookupEnv("BOT_USERS"); exists {
		tgListener = process.TelegramListener{
			TbAPI:        tbAPI,
			Bots:         bots,
			Users:        users,
			Debug:        true,
			IdleDuration: 30 * time.Second,
		}
	}

	if err := tgListener.Do(ctx); err != nil {
		log.Fatalf("[ERROR] telegram listener failed, %v", err)
	}
}

func setupLog(dbg bool) {

}
