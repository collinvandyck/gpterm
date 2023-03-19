package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Client interface {
	Complete(ctx context.Context, content string) (*CompleteResult, error)
}

type client struct {
	openai       openai.Client
	model        string
	chatContext  ChatContext
	contextCount int // how many context messages to grab
}

func New(apiKey string, opts ...Option) (Client, error) {
	res := &client{
		openai:       *openai.NewClient(apiKey),
		model:        openai.GPT3Dot5Turbo,
		chatContext:  noopChatContext{},
		contextCount: 5,
	}
	for _, o := range opts {
		o(res)
	}
	return res, nil
}

type CompleteResult struct {
	Request  openai.ChatCompletionRequest
	Response openai.ChatCompletionResponse
}

func (c *client) Complete(ctx context.Context, content string) (*CompleteResult, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "You are a helpful assistant",
		},
	}
	latest, err := c.chatContext.Latest(ctx, c.contextCount)
	if err != nil {
		return nil, fmt.Errorf("get latest: %w", err)
	}
	for _, msg := range latest {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
	}
	resp, err := c.openai.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("complete: %w", err)
	}
	err = c.chatContext.Save(ctx, req, resp)
	if err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	for _, choice := range resp.Choices {
		choice.Message.Content = strings.TrimSpace(choice.Message.Content)
	}
	res := &CompleteResult{
		Request:  req,
		Response: resp,
	}
	return res, nil
}
