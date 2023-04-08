package gptea

import "github.com/collinvandyck/gpterm/db/query"

type ConversationSwitchedMsg struct {
	Messages []query.Message
	Err      error
}
