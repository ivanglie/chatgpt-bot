package tg

import (
	"errors"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TelegramBotAPI is interface for TelegramBot with the possibility to mock it
type TelegramBotAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

// TelegramBot is a wrapper for TelegramBotAPI
type TelegramBot struct {
	bot     TelegramBotAPI
	offset  int
	timeout int
}

// Chat represents a chat.
type Chat struct {
	ID           int64
	IsGroup      bool
	IsSuperGroup bool
}

// Message is primary record to pass data from/to bots
type Message struct {
	ID   int
	Chat *Chat `json:"chat"`
}

// NewBotAPI makes a bot for Telegram
func NewBotAPI(token string, debug bool, offset, timeout int) (*TelegramBot, error) {
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
		log.Printf("Authorized on account %s\n", b.Self.UserName)
	}

	return &TelegramBot{bot: b, offset: offset, timeout: timeout}, nil
}

// GetUpdatesChan returns a channel for receiving updates
func (b *TelegramBot) GetUpdatesChan() <-chan tgbotapi.Update {
	u := tgbotapi.NewUpdate(b.offset)
	u.Timeout = b.timeout

	return b.bot.GetUpdatesChan(u)
}

// Send request to Telegram and returns the response
func (b *TelegramBot) Execute(message Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)

	if message.Chat.IsGroup || message.Chat.IsSuperGroup {
		msg.ReplyToMessageID = message.ID
	}

	if _, err := b.bot.Send(msg); err != nil {
		return err
	}

	return nil
}
