package oai

import (
	"context"
	"errors"
	"fmt"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient is interface for OpenAI with the possibility to mock it
type OpenAIClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// OpenAI is a wrapper for OpenAIClient
type OpenAI struct {
	authToken string
	client    OpenAIClient
	maxTokens int
	prompt    string
}

// NewClient makes a client for ChatGPT
// maxTokens is hard limit for the number of tokens in the response
// https://platform.openai.com/docs/api-reference/chat/create#chat/create-max_tokens
// Returns OpenAI client and error if authToken is empty
func NewClient(authToken string, maxTokens int, prompt string) (*OpenAI, error) {
	if len(authToken) == 0 {
		return nil, errors.New("OPENAI_API_KEY is empty")
	}

	client := openai.NewClient(authToken)
	log.Printf("OpenAI with prompt=%s, max=%d", prompt, maxTokens)

	return &OpenAI{authToken: authToken, client: client, maxTokens: maxTokens, prompt: prompt}, nil
}

// Send request to OpenAI and returns the response
func (o *OpenAI) Execute(request string) (response string, err error) {
	r := request
	if o.prompt != "" {
		r = o.prompt + ".\n" + request
	}

	res, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: o.maxTokens,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You answer with no more than 50 words",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: r,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	// OpenAI platform supports to return multiple chat completion choices
	// but we use only the first one
	// https://platform.openai.com/docs/api-reference/chat/create#chat/create-n
	if len(res.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	resp := res.Choices[0].Message.Content
	if len(resp) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return resp, nil
}
