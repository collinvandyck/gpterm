package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type apiKeyOption struct {
	Key string
	ti  textinput.Model
}

func newApiKeyOption() *apiKeyOption {
	ti := textinput.NewModel()
	ti.Placeholder = "Enter API Key"
	ti.Focus()
	return &apiKeyOption{ti: ti}
}

// Init implements tea.Model.
func (m *apiKeyOption) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m *apiKeyOption) Update(msg tea.Msg) (optionInterface, tea.Cmd) {
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m *apiKeyOption) View() string {
	return m.ti.View()
}
