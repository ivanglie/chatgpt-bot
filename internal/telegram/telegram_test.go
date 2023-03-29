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
	err := b.Execute(Message{ID: 0, Chat: &Chat{ID: 1, IsGroup: true, IsSuperGroup: false}}, "Ping")
	assert.Nil(t, err)
}

func TestNewBotAPI(t *testing.T) {
	bot, err := NewBotAPI("", true, 0, 0)
	assert.NotNil(t, err)
	assert.Nil(t, bot)

	bot, err = NewBotAPI("qwerty", true, 0, 0)
	assert.NotNil(t, err)
	assert.Nil(t, bot)
}
