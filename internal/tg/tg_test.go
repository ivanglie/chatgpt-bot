package tg

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

type MockBotAPI tgbotapi.BotAPI

func (m *MockBotAPI) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return nil
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return tgbotapi.Message{Text: "Pong"}, nil
}

func TestTelegramBot_Execute(t *testing.T) {
	b := &TelegramBot{bot: &MockBotAPI{}}
	res, err := b.Execute(0, "Ping")
	assert.Nil(t, err)
	assert.Equal(t, res, "Pong")
}
