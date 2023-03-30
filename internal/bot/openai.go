package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ivanglie/chatgpt-bot/internal/utils"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient is interface for OpenAI client with the possibility to mock it.
type OpenAIClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// OpenAI bot, returns responses from ChatGPT via OpenAI API.
type OpenAI struct {
	authToken string
	client    OpenAIClient
	maxTokens int
	prompt    string

	nowFn  func() time.Time // for testing
	lastDT time.Time
}

var maxMsgLen = 14000

// NewOpenAI makes a bot for ChatGPT.
// maxTokens is hard limit for the number of tokens in the response
// https://platform.openai.com/docs/api-reference/chat/create#chat/create-max_tokens
func NewOpenAI(authToken string, maxTokens int, prompt string, httpClient *http.Client) (*OpenAI, error) {
	if len(authToken) == 0 {
		return nil, fmt.Errorf("auth token is empty")
	}

	log.Printf("[INFO] OpenAI bot with github.com/sashabaranov/go-openai, prompt=%s, max=%d", prompt, maxTokens)
	config := openai.DefaultConfig(authToken)
	config.HTTPClient = httpClient

	client := openai.NewClientWithConfig(config)
	return &OpenAI{authToken: authToken, client: client, maxTokens: maxTokens, prompt: prompt,
		nowFn: time.Now}, nil
}

// OnMessage pass msg to all bots and collects responses.
func (o *OpenAI) OnMessage(msg utils.Message) (response utils.Response) {
	ok, reqText := o.request(msg.Text)
	if !ok {
		return utils.Response{}
	}

	if o.nowFn().Sub(o.lastDT) < 30*time.Minute {
		log.Printf("[WARN] OpenAI bot is too busy, last request was %s ago, %s banned", time.Since(o.lastDT), msg.From.Username)
		return utils.Response{
			Text:        fmt.Sprintf("Too many requests, the next request can be made in %d minutes.", int(30-time.Since(o.lastDT).Minutes())),
			ReadyToSend: true,
			User:        msg.From,
			ReplyTo:     msg.ID, // reply to the message
		}
	}

	responseAI, err := o.chatGPTRequest(reqText, o.prompt, "You answer with no more than 50 words")
	if err != nil {
		log.Printf("[WARN] failed to make request to ChatGPT '%s', error=%v", reqText, err)
		return utils.Response{}
	}

	// log.Printf("[DEBUG] next request to ChatGPT can be made after %s, in %d minutes",
	// 	o.lastDT.Add(30*time.Minute), int(30-time.Since(o.lastDT).Minutes()))

	r := utils.Response{
		Text:        responseAI,
		ReadyToSend: true,
		ReplyTo:     msg.ID, // reply to the message for “group”, “supergroup” or “channel” type of chat
	}

	if msg.Chat.Type == "private" {
		r = utils.Response{
			Text:        responseAI,
			ReadyToSend: true}
	}

	return r
}

func (o *OpenAI) request(text string) (react bool, reqText string) {
	for _, prefix := range o.ReactOn() {
		if strings.HasPrefix(text, prefix) {
			return true, strings.TrimSpace(strings.TrimPrefix(text, prefix))
		}
	}
	return false, ""
}

// Help returns help message
func (o *OpenAI) Help() string {
	return utils.GenHelpMsg(o.ReactOn(), "Ask something to ChatGPT")
}

func (o *OpenAI) chatGPTRequest(request, userPrompt, sysPrompt string) (response string, err error) {

	r := request
	if userPrompt != "" {
		r = userPrompt + ".\n" + request
	}

	if len(r) > maxMsgLen {
		r = r[:maxMsgLen]
	}

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			MaxTokens: o.maxTokens,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: sysPrompt,
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
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Summary returns summary of the text.
func (o *OpenAI) Summary(text string) (response string, err error) {
	return o.chatGPTRequest(text, "", "Make a short summary, up to 50 words, followed by a list of bullet points. Each bullet point is limited to 50 words, up to 7 in total. All in markdown format and translated to russian:\n")
}

// ReactOn keys.
func (o *OpenAI) ReactOn() []string {
	return []string{"ai!", "ии!"}
}
