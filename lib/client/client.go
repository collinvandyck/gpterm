package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/collinvandyck/gpterm/db/query"
	"github.com/sashabaranov/go-openai"
)

type Client interface {
	Complete(ctx context.Context, latest []query.Message, content string) (*CompleteResult, error)
	Stream(ctx context.Context, latest []query.Message, content string) (*StreamResult, error)
}

type client struct {
	openai openai.Client
	model  string
}

func New(apiKey string, opts ...Option) (Client, error) {
	config := openai.DefaultConfig(apiKey)
	rt := &roundTripper{
		RoundTripper: http.DefaultTransport,
	}
	httpClient := &http.Client{Transport: rt}
	config.HTTPClient = httpClient
	oaiClient := openai.NewClientWithConfig(config)
	res := &client{
		openai: *oaiClient,
		model:  openai.GPT3Dot5Turbo,
	}
	for _, o := range opts {
		o(res, rt)
	}
	return res, nil
}

type StreamResult struct {
	Req      openai.ChatCompletionRequest
	Response *openai.ChatCompletionStream
}

type CompleteResult struct {
	Req      openai.ChatCompletionRequest
	Response openai.ChatCompletionResponse
}

func (c *client) preamble() []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "You are a helpful assistant. If you write code make sure it is in a code block with language annotations.",
		},
	}
}

func (c *client) Stream(ctx context.Context, latest []query.Message, content string) (*StreamResult, error) {
	messages := c.preamble()
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
		Stream:   true,
	}
	resp, err := c.openai.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("stream: %w", err)
	}
	res := &StreamResult{
		Req:      req,
		Response: resp,
	}
	return res, nil
}

func (c *client) Complete(ctx context.Context, latest []query.Message, content string) (*CompleteResult, error) {
	messages := c.preamble()
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
	for _, choice := range resp.Choices {
		choice.Message.Content = strings.TrimSpace(choice.Message.Content)
	}
	res := &CompleteResult{
		Req:      req,
		Response: resp,
	}
	return res, nil
}

func (c *client) Close() error {
	return nil
}
