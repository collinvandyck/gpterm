package store

import (
	"context"

	"github.com/collinvandyck/gpterm/db/query"
	"github.com/collinvandyck/gpterm/lib/client"
	"github.com/sashabaranov/go-openai"
)

var _ client.ChatContext = &chatContext{}

type chatContext struct {
	*Store
}

func (c *chatContext) Latest(ctx context.Context, max int) ([]query.Message, error) {
	return c.GetLastMessages(ctx, max)
}

func (c *chatContext) Save(ctx context.Context, req openai.ChatCompletionRequest, res openai.ChatCompletionResponse) error {
	return c.SaveRequestResponse(ctx, req, res)
}
