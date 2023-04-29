package client

import (
	"context"
	"embed"
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

//go:embed preambles/*
var preambles embed.FS

func getPreamble(name string) string {
	bs, err := preambles.ReadFile(fmt.Sprintf("preambles/%s.md", name))
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(bs))
}

func (c *client) preamble() []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: getPreamble("default"),
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
