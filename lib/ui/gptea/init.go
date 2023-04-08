package gptea

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/collinvandyck/gpterm/lib/term"
)

type ScrollbackClearedMsg struct{}

func ClearScrollback() tea.Msg {
	term.ClearScrollback()
	return ScrollbackClearedMsg{}
}

func WindowResized(msg tea.WindowSizeMsg, ready bool) func() tea.Msg {
	return func() tea.Msg {
		return WindowSizeMsg{
			WindowSizeMsg: msg,
			Ready:         ready,
		}
	}
}

type WindowSizeMsg struct {
	tea.WindowSizeMsg
	Ready bool
}

func Init() tea.Msg {
	return InitMsg{}
}

type InitMsg struct {
}
