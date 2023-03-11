package gpterm

import (
	"context"
	"fmt"
	"strings"

	"github.com/collinvandyck/gpterm/db/query"
	"github.com/collinvandyck/gpterm/lib/slicex"
	"github.com/sashabaranov/go-openai"
)

type Client struct {
	openai  openai.Client
	model   string
	store   *Store
	queries *query.Queries
}

func NewClient(ctx context.Context, store *Store) (*Client, error) {
	apiKey, err := store.GetAPIKey(ctx)
	if err != nil {
		return nil, err
	}
	res := &Client{
		openai: *openai.NewClient(apiKey),
		model:  openai.GPT3Dot5Turbo,
		store:  store,
	}
	return res, nil
}

func (c *Client) Complete(ctx context.Context, content string) ([]string, error) {
	resp, err := c.openai.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "You are a helpful assistant",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("complete: %w", err)
	}
	res := slicex.Map(resp.Choices, func(in openai.ChatCompletionChoice) string {
		return strings.TrimSpace(in.Message.Content)
	})
	return res, nil
}

func (c *Client) Close() error {
	return c.store.Close()
}
