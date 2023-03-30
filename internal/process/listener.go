package process

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/ivanglie/chatgpt-bot/internal/bot"
	"github.com/ivanglie/chatgpt-bot/internal/utils"
)

// TelegramListener listens to tg update, forward to bots and send back responses.
// Not thread safe
type TelegramListener struct {
	TbAPI        tbAPI
	Bots         bot.BotInterface
	Users        string // comma separated list of users
	Debug        bool
	IdleDuration time.Duration
	chatID       int64

	msgs struct {
		once sync.Once
		ch   chan utils.Response
	}
}

type tbAPI interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// Do process all events, blocked call.
func (l *TelegramListener) Do(ctx context.Context) error {
	log.Printf("[INFO] start telegram listener for users %s", l.Users)

	l.msgs.once.Do(func() {
		l.msgs.ch = make(chan utils.Response, 100)
		if l.IdleDuration == 0 {
			l.IdleDuration = 30 * time.Second
		}
	})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := l.TbAPI.GetUpdatesChan(u)

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		case update, ok := <-updates:
			if !ok {
				return fmt.Errorf("telegram update chan closed")
			}

			if u := update.Message.Chat.UserName; len(l.Users) != 0 && !strings.Contains(l.Users, u) {
				log.Printf("[WARNING] user %s is not allowed to use this bot", u)
				continue
			}

			if update.Message == nil {
				log.Print("[DEBUG] empty message body")
				continue
			}

			msgJSON, errJSON := json.Marshal(update.Message)
			if errJSON != nil {
				log.Printf("[ERROR] failed to marshal update.Message to json: %v", errJSON)
				continue
			}
			log.Printf("[DEBUG] %s", string(msgJSON))

			if update.Message.Chat == nil {
				log.Print("[DEBUG] ignoring message not from chat")
				continue
			}

			fromChat := update.Message.Chat.ID

			msg := l.transform(update.Message)

			log.Printf("[DEBUG] incoming msg: %+v", msg)

			resp := l.Bots.OnMessage(*msg)

			if err := l.sendBotResponse(resp, fromChat); err != nil {
				log.Printf("[WARN] failed to respond on update, %v", err)
			}

		case resp := <-l.msgs.ch: // publish messages from outside clients
			if err := l.sendBotResponse(resp, l.chatID); err != nil {
				log.Printf("[WARN] failed to respond on rtjc event, %v", err)
			}

		case <-time.After(l.IdleDuration): // hit bots on idle timeout
			resp := l.Bots.OnMessage(utils.Message{Text: "idle"})
			if err := l.sendBotResponse(resp, l.chatID); err != nil {
				log.Printf("[WARN] failed to respond on idle, %v", err)
			}
		}
	}
}

// sendBotResponse sends bot's answer to tg channel and saves it to log.
func (l *TelegramListener) sendBotResponse(resp utils.Response, chatID int64) error {
	if !resp.ReadyToSend {
		return nil
	}

	log.Printf("[DEBUG] bot response - %+v, pin: %t, reply-to:%d", resp.Text, resp.Pin, resp.ReplyTo)
	tbMsg := tgbotapi.NewMessage(chatID, resp.Text)
	tbMsg.ParseMode = tgbotapi.ModeMarkdown
	tbMsg.DisableWebPagePreview = !resp.Preview
	tbMsg.ReplyToMessageID = resp.ReplyTo
	res, err := l.TbAPI.Send(tbMsg)
	if err != nil {
		return fmt.Errorf("can't send message to telegram %q: %w", resp.Text, err)
	}

	if resp.Pin {
		_, err = l.TbAPI.Request(tgbotapi.PinChatMessageConfig{ChatID: chatID, MessageID: res.MessageID, DisableNotification: true})
		if err != nil {
			return fmt.Errorf("can't pin message to telegram: %w", err)
		}
	}

	if resp.Unpin {
		_, err = l.TbAPI.Request(tgbotapi.UnpinChatMessageConfig{ChatID: chatID})
		if err != nil {
			return fmt.Errorf("can't unpin message to telegram: %w", err)
		}
	}

	return nil
}

func (l *TelegramListener) transform(msg *tgbotapi.Message) *utils.Message {
	message := utils.Message{
		ID:   msg.MessageID,
		Sent: msg.Time(),
		Text: msg.Text,
	}

	if msg.Chat != nil {
		message.ChatID = msg.Chat.ID
		message.Chat = &utils.Chat{ID: msg.Chat.ID, Type: msg.Chat.Type}
	}

	if msg.From != nil {
		message.From = utils.User{
			ID:          msg.From.ID,
			Username:    msg.From.UserName,
			DisplayName: msg.From.FirstName + " " + msg.From.LastName,
		}
	}

	if msg.SenderChat != nil {
		message.SenderChat = utils.SenderChat{
			ID:       msg.SenderChat.ID,
			UserName: msg.SenderChat.UserName,
		}
	}

	switch {
	case msg.Entities != nil && len(msg.Entities) > 0:
		message.Entities = l.transformEntities(msg.Entities)

	case msg.Photo != nil && len(msg.Photo) > 0:
		sizes := msg.Photo
		lastSize := sizes[len(sizes)-1]
		message.Image = &utils.Image{
			FileID:   lastSize.FileID,
			Width:    lastSize.Width,
			Height:   lastSize.Height,
			Caption:  msg.Caption,
			Entities: l.transformEntities(msg.CaptionEntities),
		}
	}

	// fill in the message's reply-to message
	if msg.ReplyToMessage != nil {
		message.ReplyTo.Text = msg.ReplyToMessage.Text
		message.ReplyTo.Sent = msg.ReplyToMessage.Time()
		if msg.ReplyToMessage.From != nil {
			message.ReplyTo.From = utils.User{
				ID:          msg.ReplyToMessage.From.ID,
				Username:    msg.ReplyToMessage.From.UserName,
				DisplayName: msg.ReplyToMessage.From.FirstName + " " + msg.ReplyToMessage.From.LastName,
			}
		}
		if msg.ReplyToMessage.SenderChat != nil {
			message.ReplyTo.SenderChat = utils.SenderChat{
				ID:       msg.ReplyToMessage.SenderChat.ID,
				UserName: msg.ReplyToMessage.SenderChat.UserName,
			}
		}
	}

	return &message
}

func (l *TelegramListener) transformEntities(entities []tgbotapi.MessageEntity) *[]utils.Entity {
	if len(entities) == 0 {
		return nil
	}

	result := make([]utils.Entity, 0, len(entities))
	for _, entity := range entities {
		e := utils.Entity{
			Type:   entity.Type,
			Offset: entity.Offset,
			Length: entity.Length,
			URL:    entity.URL,
		}
		if entity.User != nil {
			e.User = &utils.User{
				ID:          entity.User.ID,
				Username:    entity.User.UserName,
				DisplayName: entity.User.FirstName + " " + entity.User.LastName,
			}
		}
		result = append(result, e)
	}

	return &result
}
