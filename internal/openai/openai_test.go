package openai

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
	c, err := NewClient("", 0, "")
	assert.Nil(t, c)
	assert.NotNil(t, err)

	c, err = NewClient("OPENAI_API_KEY", 0, "")
	assert.NotNil(t, c)
	assert.Nil(t, err)
}

func TestOpenAI_Execute(t *testing.T) {
	c, _ := NewClient("OPENAI_API_KEY", 0, "")
	c.client = &MockOpenAI{}

	res, err := c.Execute("Ping", false)
	assert.Nil(t, err)
	assert.Equal(t, res, "Pong")

	res, err = c.Execute("Ping", true)
	assert.NotNil(t, err)
	assert.Equal(t, res, "")

	res, err = c.Execute("ai! Ping", true)
	assert.Nil(t, err)
	assert.Equal(t, res, "Pong")
}

func TestOpenAI_request(t *testing.T) {
	o := &OpenAI{}

	ok, req := o.request("Ping")
	assert.False(t, ok)
	assert.Equal(t, req, "")

	ok, req = o.request("ai! Ping")
	assert.True(t, ok)
	assert.Equal(t, req, "Ping")
}
