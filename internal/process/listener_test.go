package process

import (
	"context"
	"testing"
	"time"

	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"

	"github.com/ivanglie/chatgpt-bot/internal/bot"
	"github.com/ivanglie/chatgpt-bot/internal/utils"
)

func TestTelegramListener_DoNoBots(t *testing.T) {
	tbAPI := &tbAPIMock{GetChatFunc: func(config tbapi.ChatInfoConfig) (tbapi.Chat, error) {
		return tbapi.Chat{ID: 123}, nil
	}}
	bots := &bot.InterfaceMock{OnMessageFunc: func(msg utils.Message) utils.Response {
		return utils.Response{ReadyToSend: false}
	}}

	l := TelegramListener{
		TbAPI: tbAPI,
		Bots:  bots,
		Users: "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	updMsg := tbapi.Update{
		Message: &tbapi.Message{
			ReplyToMessage: &tbapi.Message{
				SenderChat: &tbapi.Chat{
					ID:        4321,
					UserName:  "another_user",
					FirstName: "first",
					LastName:  "last",
				},
				Chat: &tbapi.Chat{ID: 123},
				Text: "text 123",
				From: &tbapi.User{UserName: "user"},
			},
			Chat: &tbapi.Chat{ID: 123},
			Text: "text 123",
			From: &tbapi.User{UserName: "user"},
		},
	}

	updChan := make(chan tbapi.Update, 1)
	updChan <- updMsg
	close(updChan)
	tbAPI.GetUpdatesChanFunc = func(config tbapi.UpdateConfig) tbapi.UpdatesChannel { return updChan }

	err := l.Do(ctx)
	assert.EqualError(t, err, "telegram update chan closed")
}

func TestTelegramListener_DoWithBots(t *testing.T) {
	tbAPI := &tbAPIMock{
		GetChatFunc: func(config tbapi.ChatInfoConfig) (tbapi.Chat, error) {
			return tbapi.Chat{ID: 123}, nil
		},
		SendFunc: func(c tbapi.Chattable) (tbapi.Message, error) {
			return tbapi.Message{Text: c.(tbapi.MessageConfig).Text, From: &tbapi.User{UserName: "user"}}, nil
		},
	}
	bots := &bot.InterfaceMock{OnMessageFunc: func(msg utils.Message) utils.Response {
		t.Logf("on-message: %+v", msg)
		if msg.Text == "text 123" && msg.From.Username == "user" {
			return utils.Response{ReadyToSend: true, Text: "bot's answer"}
		}
		return utils.Response{}
	}}

	l := TelegramListener{
		TbAPI: tbAPI,
		Bots:  bots,
		Users: "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Minute)
	defer cancel()

	updMsg := tbapi.Update{
		Message: &tbapi.Message{
			Chat: &tbapi.Chat{ID: 123},
			Text: "text 123",
			From: &tbapi.User{UserName: "user"},
			Date: int(time.Date(2020, 2, 11, 19, 35, 55, 9, time.UTC).Unix()),
		},
	}

	updChan := make(chan tbapi.Update, 1)
	updChan <- updMsg
	close(updChan)
	tbAPI.GetUpdatesChanFunc = func(config tbapi.UpdateConfig) tbapi.UpdatesChannel { return updChan }

	err := l.Do(ctx)
	assert.EqualError(t, err, "telegram update chan closed")
	assert.Equal(t, 1, len(tbAPI.SendCalls()))
	assert.Equal(t, "bot's answer", tbAPI.SendCalls()[0].C.(tbapi.MessageConfig).Text)
}

func TestTelegramListener_DoPinMessages(t *testing.T) {
	tbAPI := &tbAPIMock{
		GetChatFunc: func(config tbapi.ChatInfoConfig) (tbapi.Chat, error) {
			return tbapi.Chat{ID: 123}, nil
		},
		SendFunc: func(c tbapi.Chattable) (tbapi.Message, error) {
			if c.(tbapi.MessageConfig).Text == "bot's answer" {
				return tbapi.Message{MessageID: 456, Text: c.(tbapi.MessageConfig).Text, From: &tbapi.User{UserName: "user"}}, nil
			}
			return tbapi.Message{}, nil
		},
		RequestFunc: func(c tbapi.Chattable) (*tbapi.APIResponse, error) {
			return &tbapi.APIResponse{Ok: true}, nil
		},
	}
	bots := &bot.InterfaceMock{OnMessageFunc: func(msg utils.Message) utils.Response {
		t.Logf("on-message: %+v", msg)
		if msg.Text == "text 123" && msg.From.Username == "user" {
			return utils.Response{ReadyToSend: true, Text: "bot's answer", Pin: true}
		}
		return utils.Response{}
	}}

	l := TelegramListener{

		TbAPI: tbAPI,
		Bots:  bots,
		Users: "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Minute)
	defer cancel()

	updMsg := tbapi.Update{
		Message: &tbapi.Message{
			Chat: &tbapi.Chat{ID: 123},
			Text: "text 123",
			From: &tbapi.User{UserName: "user"},
			Date: int(time.Date(2020, 2, 11, 19, 35, 55, 9, time.UTC).Unix()),
		},
	}

	updChan := make(chan tbapi.Update, 1)
	updChan <- updMsg
	close(updChan)
	tbAPI.GetUpdatesChanFunc = func(config tbapi.UpdateConfig) tbapi.UpdatesChannel { return updChan }

	err := l.Do(ctx)
	assert.EqualError(t, err, "telegram update chan closed")
	assert.Equal(t, 1, len(bots.OnMessageCalls()))
	assert.Equal(t, 1, len(tbAPI.SendCalls()))
	assert.Equal(t, 1, len(tbAPI.RequestCalls()))
	assert.Equal(t, 456, tbAPI.RequestCalls()[0].C.(tbapi.PinChatMessageConfig).MessageID)
}

func TestTelegramListener_DoUnpinMessages(t *testing.T) {
	tbAPI := &tbAPIMock{
		GetChatFunc: func(config tbapi.ChatInfoConfig) (tbapi.Chat, error) {
			return tbapi.Chat{ID: 123}, nil
		},
		SendFunc: func(c tbapi.Chattable) (tbapi.Message, error) {
			if c.(tbapi.MessageConfig).Text == "bot's answer" {
				return tbapi.Message{Text: c.(tbapi.MessageConfig).Text, From: &tbapi.User{UserName: "user"}}, nil
			}
			return tbapi.Message{}, nil
		},
		RequestFunc: func(c tbapi.Chattable) (*tbapi.APIResponse, error) {
			return &tbapi.APIResponse{Ok: true}, nil
		},
	}
	bots := &bot.InterfaceMock{OnMessageFunc: func(msg utils.Message) utils.Response {
		t.Logf("on-message: %+v", msg)
		if msg.Text == "text 123" && msg.From.Username == "user" {
			return utils.Response{ReadyToSend: true, Text: "bot's answer", Unpin: true}
		}
		return utils.Response{}
	}}

	l := TelegramListener{
		TbAPI: tbAPI,
		Bots:  bots,
		Users: "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Minute)
	defer cancel()

	updMsg := tbapi.Update{
		Message: &tbapi.Message{
			Chat: &tbapi.Chat{ID: 123},
			Text: "text 123",
			From: &tbapi.User{UserName: "user"},
			Date: int(time.Date(2020, 2, 11, 19, 35, 55, 9, time.UTC).Unix()),
		},
	}

	updChan := make(chan tbapi.Update, 1)
	updChan <- updMsg
	close(updChan)
	tbAPI.GetUpdatesChanFunc = func(config tbapi.UpdateConfig) tbapi.UpdatesChannel { return updChan }

	err := l.Do(ctx)
	assert.EqualError(t, err, "telegram update chan closed")
	assert.Equal(t, 1, len(bots.OnMessageCalls()))
	assert.Equal(t, 1, len(tbAPI.SendCalls()))
	assert.Equal(t, 1, len(tbAPI.RequestCalls()))
	assert.Equal(t, int64(123), tbAPI.RequestCalls()[0].C.(tbapi.UnpinChatMessageConfig).ChatID)
}

func TestTelegram_transformTextMessage(t *testing.T) {
	l := TelegramListener{}
	assert.Equal(
		t,
		&utils.Message{
			ID: 30,
			From: utils.User{
				ID:          100000001,
				Username:    "username",
				DisplayName: "First Last",
			},
			Sent:   time.Unix(1578627415, 0),
			Text:   "Message",
			ChatID: 123456,
			Chat:   &utils.Chat{ID: 123456, Type: ""},
		},
		l.transform(
			&tbapi.Message{
				Chat: &tbapi.Chat{
					ID: 123456,
				},
				From: &tbapi.User{
					ID:        100000001,
					UserName:  "username",
					FirstName: "First",
					LastName:  "Last",
				},
				MessageID: 30,
				Date:      1578627415,
				Text:      "Message",
			},
		),
	)
}

func TestTelegram_transformPhoto(t *testing.T) {
	l := TelegramListener{}
	assert.Equal(
		t,
		&utils.Message{
			Sent: time.Unix(1578627415, 0),
			Image: &utils.Image{
				FileID:  "AgADAgADFKwxG8r0qUiQByxwp9Gi4s1qwQ8ABAEAAwIAA3kAA5K9AgABFgQ",
				Width:   1280,
				Height:  597,
				Caption: "caption",
				Entities: &[]utils.Entity{
					{
						Type:   "bold",
						Offset: 0,
						Length: 7,
					},
				},
			},
		},
		l.transform(
			&tbapi.Message{
				Date: 1578627415,
				Photo: []tbapi.PhotoSize{
					{
						FileID:   "AgADAgADFKwxG8r0qUiQByxwp9Gi4s1qwQ8ABAEAAwIAA20AA5C9AgABFgQ",
						Width:    320,
						Height:   149,
						FileSize: 6262,
					},
					{
						FileID:   "AgADAgADFKwxG8r0qUiQByxwp9Gi4s1qwQ8ABAEAAwIAA3gAA5G9AgABFgQ",
						Width:    800,
						Height:   373,
						FileSize: 30240,
					},
					{
						FileID:   "AgADAgADFKwxG8r0qUiQByxwp9Gi4s1qwQ8ABAEAAwIAA3kAA5K9AgABFgQ",
						Width:    1280,
						Height:   597,
						FileSize: 55267,
					},
				},
				Caption: "caption",
				CaptionEntities: []tbapi.MessageEntity{
					{
						Type:   "bold",
						Offset: 0,
						Length: 7,
					},
				},
			},
		),
	)
}

func TestTelegram_transformEntities(t *testing.T) {
	l := TelegramListener{}
	assert.Equal(
		t,
		&utils.Message{
			Sent: time.Unix(1578627415, 0),
			Text: "@username тебя слишком много, отдохни...",
			Entities: &[]utils.Entity{
				{
					Type:   "mention",
					Offset: 0,
					Length: 9,
				},
				{
					Type:   "italic",
					Offset: 10,
					Length: 30,
				},
			},
		},
		l.transform(
			&tbapi.Message{
				Date: 1578627415,
				Text: "@username тебя слишком много, отдохни...",
				Entities: []tbapi.MessageEntity{
					{
						Type:   "mention",
						Offset: 0,
						Length: 9,
					},
					{
						Type:   "italic",
						Offset: 10,
						Length: 30,
					},
				},
			},
		),
	)
}
