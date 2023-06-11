package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var _ model = (*apiKeyOption)(nil)

type apiKeyOption struct {
	uiOpts
	ti textinput.Model
}

func newApiKeyOption(uiOpts uiOpts) *apiKeyOption {
	ti := textinput.New()
	ti.Placeholder = "Enter API Key"
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	return &apiKeyOption{
		uiOpts: uiOpts,
		ti:     ti,
	}
}

// Init implements model.
func (m *apiKeyOption) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements model.
func (m *apiKeyOption) Update(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

// View implements model.
func (m *apiKeyOption) View() string {
	return m.ti.View()
}
