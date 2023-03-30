package bot

import (
	"strings"
	"testing"

	"github.com/ivanglie/chatgpt-bot/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenHelpMsg(t *testing.T) {
	require.Equal(t, "cmd _- description_\n", utils.GenHelpMsg([]string{"cmd"}, "description"))
}

func TestMultiBotHelp(t *testing.T) {
	b1 := &InterfaceMock{HelpFunc: func() string {
		return "b1 help"
	}}
	b2 := &InterfaceMock{HelpFunc: func() string {
		return "b2 help"
	}}

	// Must return concatenated b1 and b2 without space
	// Line formatting only in genHelpMsg()
	require.Equal(t, "b1 help\nb2 help\n", BotSlice{b1, b2}.Help())
}

func TestMultiBotReactsOnHelp(t *testing.T) {
	b := &InterfaceMock{
		ReactOnFunc: func() []string {
			return []string{"help"}
		},
		HelpFunc: func() string {
			return "help"
		},
	}

	mb := BotSlice{b}
	resp := mb.OnMessage(utils.Message{Text: "help"})

	require.True(t, resp.ReadyToSend)
	require.Equal(t, "help\n", resp.Text)
}

func TestMultiBotCombinesAllBotResponses(t *testing.T) {
	msg := utils.Message{Text: "cmd"}

	b1 := &InterfaceMock{
		ReactOnFunc: func() []string { return []string{"cmd"} },
		OnMessageFunc: func(m utils.Message) utils.Response {
			return utils.Response{ReadyToSend: true, Text: "b1 resp", ReplyTo: 789}
		},
	}
	b2 := &InterfaceMock{
		ReactOnFunc:   func() []string { return []string{"cmd"} },
		OnMessageFunc: func(m utils.Message) utils.Response { return utils.Response{ReadyToSend: true, Text: "b2 resp"} },
	}

	mb := BotSlice{b1, b2}
	resp := mb.OnMessage(msg)
	t.Logf("resp: %+v", resp)

	require.True(t, resp.ReadyToSend)
	parts := strings.Split(resp.Text, "\n")
	require.Len(t, parts, 2)
	require.Contains(t, parts, "b1 resp")
	require.Contains(t, parts, "b2 resp")
	assert.Equal(t, 789, resp.ReplyTo)
}
