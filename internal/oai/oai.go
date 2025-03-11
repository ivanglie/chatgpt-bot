package oai

import (
	"context"
	"errors"
	"fmt"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient is interface for OpenAI with the possibility to mock it.
type OpenAIClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// OpenAI is a wrapper for OpenAIClient.
type OpenAI struct {
	authToken string
	client    OpenAIClient
	maxTokens int
	prompt    string
	history   []openai.ChatCompletionMessage
}

// New makes a client for ChatGPT.
// maxTokens is hard limit for the number of tokens in the response
// https://platform.openai.com/docs/api-reference/chat/create#chat/create-max_tokens
// Returns OpenAI client and error if authToken is empty
func New(authToken string, maxTokens int, prompt string) (*OpenAI, error) {
	if len(authToken) == 0 {
		return nil, errors.New("OPENAI_API_KEY is empty")
	}
	client := openai.NewClient(authToken)

	history := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You answer with no more than 50 words",
		},
	}

	if prompt != "" {
		history = append(history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		})
	}

	log.Printf("[DEBUG] OpenAI with prompt=%s, max=%d", prompt, maxTokens)
	return &OpenAI{authToken: authToken, client: client, maxTokens: maxTokens, prompt: prompt, history: history}, nil
}

// Generate returns a response for the request using ChatGPT.
func (o *OpenAI) Generate(request string) (response string, err error) {
	o.history = append(o.history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: request,
	})

	res, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT4oMini,
			MaxTokens: o.maxTokens,
			Messages:  o.history,
		},
	)

	if err != nil {
		return "", err
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	resp := res.Choices[0].Message.Content
	if len(resp) == 0 {
		return "", fmt.Errorf("empty response")
	}

	o.history = append(o.history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: resp,
	})

	return resp, nil
}
