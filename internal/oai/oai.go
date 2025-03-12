package oai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient is interface for OpenAI with the possibility to mock it.
type OpenAIClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// OpenAI is a wrapper for OpenAIClient.
type OpenAI struct {
	mu sync.RWMutex

	authToken     string
	client        OpenAIClient
	maxTokens     int
	prompt        string
	chatHistories map[string][]openai.ChatCompletionMessage
}

// New makes a client for ChatGPT.
func New(authToken string, maxTokens int, prompt string) (*OpenAI, error) {
	if len(authToken) == 0 {
		return nil, errors.New("OPENAI_API_KEY is empty")
	}

	client := openai.NewClient(authToken)
	log.Printf("[DEBUG] OpenAI with prompt=%s, max=%d", prompt, maxTokens)

	return &OpenAI{
		authToken:     authToken,
		client:        client,
		maxTokens:     maxTokens,
		prompt:        prompt,
		chatHistories: make(map[string][]openai.ChatCompletionMessage),
	}, nil
}

// Generate returns a response for the specific user and chat.
func (o *OpenAI) Generate(userID, chatID, request string) (response string, err error) {
	chatKey := userID + ":" + chatID

	o.mu.RLock()
	history, exists := o.chatHistories[chatKey]
	o.mu.RUnlock()

	if !exists {
		history = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You answer with no more than 50 words",
			},
		}

		if o.prompt != "" {
			history = append(history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: o.prompt,
			})
		}
	}

	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: request,
	})

	res, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT4oMini,
			MaxTokens: o.maxTokens,
			Messages:  history,
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

	history = append(history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: resp,
	})

	o.mu.Lock()
	o.chatHistories[chatKey] = history
	o.mu.Unlock()

	return resp, nil
}
