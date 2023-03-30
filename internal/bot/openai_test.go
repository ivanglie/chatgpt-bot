package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	ai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivanglie/chatgpt-bot/internal/bot/mocks"
	"github.com/ivanglie/chatgpt-bot/internal/utils"
)

func TestOpenAI_Help(t *testing.T) {
	require.Contains(t, (&OpenAI{}).Help(), "ai!")
}

func TestOpenAI_OnMessage(t *testing.T) {
	// Example of response from OpenAI API
	// https://platform.openai.com/docs/api-reference/chat
	jsonResponse, err := os.ReadFile("testdata/chat_completion_response.json")
	require.NoError(t, err)

	tbl := []struct {
		request    string
		prompt     string
		json       []byte
		mockResult bool
		response   utils.Response
	}{
		{"Good result", "prompt", jsonResponse, true, utils.Response{Text: "Mock response", ReadyToSend: true, ReplyTo: 756}},
		{"Good result", "", jsonResponse, true, utils.Response{Text: "Mock response", ReadyToSend: true, ReplyTo: 756}},
		{"Error result", "", jsonResponse, false, utils.Response{}},
		{"Empty result", "", []byte(`{}`), true, utils.Response{}},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mockOpenAIClient := &mocks.OpenAIClient{
				CreateChatCompletionFunc: func(ctx context.Context, request ai.ChatCompletionRequest) (ai.ChatCompletionResponse, error) {
					var response ai.ChatCompletionResponse

					err = json.Unmarshal(tt.json, &response)
					require.NoError(t, err)

					if !tt.mockResult {
						return ai.ChatCompletionResponse{}, fmt.Errorf("mock error")
					}

					return response, nil
				},
			}

			o, _ := NewOpenAI("ss-mockToken", 100, tt.prompt, &http.Client{Timeout: 10 * time.Second})
			o.client = mockOpenAIClient

			assert.Equal(t,
				tt.response,
				o.OnMessage(utils.Message{Text: fmt.Sprintf("ai! %s", tt.request), ID: 756, Chat: &utils.Chat{ID: 123, Type: "group"}}),
			)
			calls := mockOpenAIClient.CreateChatCompletionCalls()
			assert.Equal(t, 1, len(calls))
			// First message is system role setup
			expRequest := tt.request
			if tt.prompt != "" {
				expRequest = tt.prompt + ".\n" + tt.request
			}
			assert.Equal(t, expRequest, calls[0].ChatCompletionRequest.Messages[1].Content)
		})
	}
}

func TestOpenAI_request(t *testing.T) {
	tbl := []struct {
		text string
		ok   bool
		req  string
	}{
		{"ai! valid request", true, "valid request"},
		{"", false, ""},
		{"not valid request", false, ""},
		{"chat not valid request", false, ""},
		{"blah ai! test", false, ""},
		{"ai! chat test", true, "chat test"},
	}

	o := &OpenAI{}
	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ok, req := o.request(tt.text)
			if !tt.ok {
				assert.False(t, ok)
				return
			}
			assert.True(t, ok)
			assert.Equal(t, tt.req, req)
		})
	}
}
