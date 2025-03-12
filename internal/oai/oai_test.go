package oai

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

type MockOpenAI struct{}

func (m *MockOpenAI) CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	res := openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "Pong"}}}}
	return res, nil
}

func TestNewClient(t *testing.T) {
	c, err := New("", 0, "")
	assert.Nil(t, c)
	assert.NotNil(t, err)

	c, err = New("OPENAI_API_KEY", 0, "")
	assert.NotNil(t, c)
	assert.Nil(t, err)
}

func TestOpenAI_Execute(t *testing.T) {
	c, _ := New("OPENAI_API_KEY", 0, "")
	c.client = &MockOpenAI{}

	res, err := c.Generate("userID", "chatID", "Ping")
	assert.Nil(t, err)
	assert.Equal(t, res, "Pong")
}
