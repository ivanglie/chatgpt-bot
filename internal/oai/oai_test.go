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

func TestOpenAI_Execute(t *testing.T) {
	c := NewClient("OPENAI_API_KEY", 0, "")
	c.client = &MockOpenAI{}

	res, err := c.Execute("Ping")
	assert.Nil(t, err)
	assert.Equal(t, res, "Pong")
}
