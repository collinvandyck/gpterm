package ui

import tea "github.com/charmbracelet/bubbletea"

type optionInterface interface {
	Init() tea.Cmd
	Update(tea.Msg) (optionInterface, tea.Cmd)
	View() string
}
