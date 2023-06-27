package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ivanglie/chatgpt-bot/internal/oai"
	"github.com/ivanglie/chatgpt-bot/internal/tg"
	"github.com/jessevdk/go-flags"
	"golang.org/x/exp/slices"
)

var (
	opts struct {
		BotToken     string   `long:"bottoken" env:"BOT_TOKEN" description:"telegram bot token"`
		OnenAIAPIKey string   `long:"openaiapikey" env:"OPENAI_API_KEY" description:"OpenAI API key"`
		BotUsers     []string `long:"botusers" env:"BOT_USERS" env-delim:"," description:"bot users"`
		Dbg          bool     `long:"dbg" env:"DEBUG" description:"use debug"`
	}

	version = "unknown"
)

func main() {
	fmt.Printf("chatgpt-bot %s\n", version)
	p := flags.NewParser(&opts, flags.Default)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			log.Printf("[ERROR] chatgpt-bot error: %v", err)
		}
		os.Exit(2)
	}

	openAI, err := oai.New(opts.OnenAIAPIKey, 1000, "")
	if err != nil {
		log.Panic(err)
	}

	tBot, err := tg.New(opts.BotToken, true, 0, 60)
	if err != nil {
		log.Panic(err)
	}

	users := opts.BotUsers
	log.Printf("users: %s, len: %d\n", users, len(users))

	updates := tBot.GetUpdatesChan()

	for update := range updates {
		if update.Message == nil || update.Message.IsCommand() {
			continue
		}

		if u := update.Message.Chat.UserName; len(u) == 0 || (len(users) == 0 || !slices.Contains(users, u)) {
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
