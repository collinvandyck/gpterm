package client

import (
	"context"

	"github.com/collinvandyck/gpterm/db/query"
	"github.com/sashabaranov/go-openai"
)

type ChatContext interface {
	Latest(ctx context.Context, max int) ([]query.Message, error)
	Save(ctx context.Context, req openai.ChatCompletionRequest, res openai.ChatCompletionResponse) error
}

type noopChatContext struct{}

func (n noopChatContext) Latest(_ context.Context, _ int) ([]query.Message, error) {
	return nil, nil
}

func (n noopChatContext) Save(_ context.Context, _ openai.ChatCompletionRequest, _ openai.ChatCompletionResponse) error {
	return nil
}
