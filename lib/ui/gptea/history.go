package gptea

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/collinvandyck/gpterm/db/query"
)

type BacklogMsg struct {
	Messages []query.Message
	Err      error
}

type BacklogPrintedMsg struct {
}

type HistoryMsg struct {
	Messages []query.Message
	Err      error
}

func HistoryPrinted() tea.Msg {
	return HistoryPrintedMsg{}
}

type HistoryPrintedMsg struct{}
