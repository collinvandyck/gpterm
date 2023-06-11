package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ model = (*apiKeyOption)(nil)

type apiKeyOption struct {
	Key   string
	ti    textinput.Model
	focus int
	data  *optionsData
}

type apiKeyState int

func newApiKeyOption(placeholder string) *apiKeyOption {
	ti := textinput.NewModel()
	ti.Placeholder = placeholder
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	return &apiKeyOption{ti: ti}
}

func (m *apiKeyOption) Init() tea.Cmd {
	return textinput.Blink
}

func (m *apiKeyOption) Update(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case *optionsData:
		m.data = msg
		m.ti.SetValue(msg.apiKey)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.focus {
			case 0:
				m.ti.Blur()
				m.focus++
			case 1:
				// enter
				m.data.apiKey = m.ti.Value()
			case 2:
				// cancel
			}
		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			m.focus++
			if s == "up" || s == "shift+tab" {
				m.focus = (m.focus - 2)
			}
			m.focus = (m.focus + 3) % 3
			m.ti.Blur()
			switch m.focus {
			case 0:
				m.ti.Focus()
			}
		}
	}
	if m.focus == 0 {
		m.ti, cmd = m.ti.Update(msg)
	}
	return m, cmd
}

func (m *apiKeyOption) View() string {
	var b strings.Builder
	b.WriteString(m.ti.View())
	b.WriteString("\n\n")

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF7DB")).
		Background(lipgloss.Color("#888B7E")).
		Padding(0, 3).
		MarginRight(2).
		MarginTop(1)

	activeButtonStyle := buttonStyle.Copy().
		Foreground(lipgloss.Color("#FFF7DB")).
		Background(lipgloss.Color("#F25D94")).
		Underline(true)

	var okButtonStyle = buttonStyle
	var cancelButtonStyle = buttonStyle
	switch m.focus {
	case 0:
	case 1:
		okButtonStyle = activeButtonStyle
	case 2:
		cancelButtonStyle = activeButtonStyle
	}

	okButton := okButtonStyle.Render("OK")
	cancelButton := cancelButtonStyle.Render("Cancel")
	choice := lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
	b.WriteString(choice)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("focus: %v", m.focus))
	return b.String()
}
