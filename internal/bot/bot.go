package bot

import (
	"context"
	"log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ivanglie/chatgpt-bot/internal/utils"
)

type BotInterface interface {
	Help() string
	ReactOn() []string
	OnMessage(message utils.Message) (response utils.Response)
}

// BotSlice is a slice of BotInterface.
type BotSlice []BotInterface

// Help returns help message for all bots.
func (b BotSlice) Help() string {
	sb := strings.Builder{}

	for _, bot := range b {
		help := bot.Help()
		if len(help) != 0 {
			// Because WriteString always returns nil error
			if !strings.HasSuffix(help, "\n") {
				help += "\n"
			}

			sb.WriteString(help)
		}
	}

	return sb.String()
}

// ReactOn returns all keywords of all bots.
func (b BotSlice) ReactOn() (res []string) {
	for _, bot := range b {
		res = append(res, bot.ReactOn()...)
	}

	return
}

// OnMessage passes message to bots array and combines all responses.
func (b BotSlice) OnMessage(message utils.Message) (response utils.Response) {
	if utils.Contains([]string{"help", "/help", "help!"}, message.Text) {
		return utils.Response{Text: b.Help(), ReadyToSend: true}
	}

	resps := make(chan string)
	var pin, unpin int32
	var channelID int64
	var user utils.User
	var replyTo int

	wg := &sync.WaitGroup{}
	for _, bot := range b {
		bot := bot

		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()

			if resp := bot.OnMessage(message); resp.ReadyToSend {
				resps <- resp.Text
				if resp.Pin {
					atomic.AddInt32(&pin, 1)
				}
				if resp.Unpin {
					atomic.AddInt32(&unpin, 1)
				}
				if resp.ReplyTo > 0 {
					replyTo = resp.ReplyTo
				}
			}
		}(context.Background())
	}

	go func() {
		wg.Wait()
		close(resps)
	}()

	lines := make([]string, 0, len(resps))
	for r := range resps {
		log.Printf("[DEBUG] collect %q", r)
		lines = append(lines, r)
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i] < lines[j]
	})

	log.Printf("[DEBUG] answers %d, readyToSend %v", len(lines), len(lines) > 0)
	return utils.Response{
		Text:        strings.Join(lines, "\n"),
		ReadyToSend: len(lines) > 0,
		Pin:         atomic.LoadInt32(&pin) > 0,
		Unpin:       atomic.LoadInt32(&unpin) > 0,
		User:        user,
		ChannelID:   channelID,
		ReplyTo:     replyTo,
	}
}
