package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type apiKeyOption struct {
	Key string
	ti  textinput.Model
}

func newApiKeyOption(placeholder string) *apiKeyOption {
	ti := textinput.NewModel()
	ti.Placeholder = placeholder
	ti.Focus()
	return &apiKeyOption{ti: ti}
}

func (m *apiKeyOption) Init() tea.Cmd {
	return textinput.Blink
}

func (m *apiKeyOption) Update(msg tea.Msg) (optionInterface, tea.Cmd) {
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

func (m *apiKeyOption) View() string {
	var b strings.Builder
	b.WriteString(m.ti.View())
	b.WriteString("\n\n")
	b.WriteString("yes/no")
	return b.String()
}
