package gptea

import "github.com/collinvandyck/gpterm/db/query"

type ConversationHistoryMsg struct {
	Val int
	Err error
}

type ConversationSwitchedMsg struct {
	Messages []query.Message
	Err      error
}
