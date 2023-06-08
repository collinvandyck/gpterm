package ui

import tea "github.com/charmbracelet/bubbletea"

type options struct {
}

// Init implements tea.Model.
func (o options) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (o options) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return o, nil
}

// View implements tea.Model.
func (o options) View() string {
	return "options"
}
