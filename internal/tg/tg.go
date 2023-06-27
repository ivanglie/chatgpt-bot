package tg

import (
	"errors"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBotAPI is interface for TelegramBot with the possibility to mock it.
type TelegramBotAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

// TelegramBot is a wrapper for TelegramBotAPI.
type TelegramBot struct {
	bot     TelegramBotAPI
	offset  int
	timeout int
}

// New makes a bot for Telegram.
func New(token string, debug bool, offset, timeout int) (*TelegramBot, error) {
	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	if timeout == 0 {
		timeout = 60 // By default, the timeout is 60 seconds
	}

	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	b.Debug = debug
	if debug {
		log.Printf("[ERROR] Authorized on account %s\n", b.Self.UserName)
	}

	return &TelegramBot{bot: b, offset: offset, timeout: timeout}, nil
}

// GetUpdatesChan returns a channel for receiving updates.
func (b *TelegramBot) GetUpdatesChan() <-chan tgbotapi.Update {
	u := tgbotapi.NewUpdate(b.offset)
	u.Timeout = b.timeout

	return b.bot.GetUpdatesChan(u)
}

// Send request to Telegram and returns the response.
func (b *TelegramBot) Send(chatID int64, request string) (response string, err error) {
	req := tgbotapi.NewMessage(chatID, request)

	var res tgbotapi.Message
	res, err = b.bot.Send(req)
	if err != nil {
		return "", err
	}

	return res.Text, nil
}
