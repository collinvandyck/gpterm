package options

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ApiKeyModel struct {
	Key string
	ti  textinput.Model
}

func NewAPIKeyModel() ApiKeyModel {
	ti := textinput.NewModel()
	ti.Placeholder = "Enter API Key"
	ti.Focus()
	return ApiKeyModel{ti: ti}
}

// Init implements tea.Model.
func (m ApiKeyModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m ApiKeyModel) Update(msg tea.Msg) (Interface, tea.Cmd) {
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m ApiKeyModel) View() string {
	return m.ti.View()
}
