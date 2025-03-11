package main

import (
	"fmt"
	"os"

	"github.com/ivanglie/chatgpt-bot/internal/oai"
	"github.com/ivanglie/chatgpt-bot/internal/tg"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

var (
	opts struct {
		BotToken     string   `long:"bottoken" env:"BOT_TOKEN" description:"bot token for Telegram"`
		OnenAIAPIKey string   `long:"openaiapikey" env:"OPENAI_API_KEY" description:"key for OpenAI API"`
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

	setupLog(opts.Dbg)

	openAI, err := oai.New(opts.OnenAIAPIKey, 1000, "")
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	telegramBot, err := tg.New(opts.BotToken, opts.Dbg, 0, 60)
	if err != nil {
		log.Panic().Msg(err.Error())
	}

	users := opts.BotUsers
	log.Debug().Msgf("users: %v, len: %d", users, len(users))

	updates := telegramBot.GetUpdatesChan()

	for update := range updates {
		if update.Message == nil || update.Message.IsCommand() {
			continue
		}

		if user := update.Message.From.UserName; len(users) != 0 && !slices.Contains(users, user) {
			log.Error().Msgf("user %s is not allowed", update.Message.From.String())
			telegramBot.Send(update.Message.Chat.ID, "Access denied.")

			continue
		}

		log.Debug().Msgf("user: %s, request: %s", update.Message.From.String(), update.Message.Text)

		res, err := openAI.Generate(update.Message.Text)
		if err != nil {
			log.Error().Msg(err.Error())
			continue
		}

		log.Debug().Msgf("user: %s, response: %s", update.Message.From.String(), res)

		telegramBot.Send(update.Message.Chat.ID, res)
	}
}

func setupLog(dbg bool) {
	if dbg {
		log.Level(zerolog.DebugLevel)
		return
	}

	log.Level(zerolog.InfoLevel)
}
