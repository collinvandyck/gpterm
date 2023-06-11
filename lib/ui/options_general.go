package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ model = (*generalOptions)(nil)

type generalOptions struct {
	Key            string
	apiKey         textinput.Model
	focus          int
	data           *optionsData
	options        []model
	selectedOption int
}

type apiKeyState int

func newGeneralOptions(uiOpts uiOpts) *generalOptions {
	opts := &generalOptions{
		apiKey: textinput.NewModel(),
	}
	opts.apiKey.Placeholder = "Enter API Key"
	opts.apiKey.EchoMode = textinput.EchoPassword
	opts.apiKey.Focus()
	opts.options = append(opts.options, newApiKeyOption(uiOpts))
	return opts
}

func (m *generalOptions) Init() tea.Cmd {
	return textinput.Blink
}

func (m *generalOptions) Update(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case *optionsData:
		m.data = msg
		m.apiKey.SetValue(msg.apiKey)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.focus {
			case 0:
				m.apiKey.Blur()
				m.focus++
			case 1:
				// enter
				m.data.apiKey = m.apiKey.Value()
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
			m.apiKey.Blur()
			switch m.focus {
			case 0:
				m.apiKey.Focus()
			}
		}
	}
	if m.focus == 0 {
		m.apiKey, cmd = m.apiKey.Update(msg)
	}
	return m, cmd
}

func (m *generalOptions) View() string {
	var b strings.Builder

	sectionHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF7DB")).
		Background(lipgloss.Color("#F25D94")).
		Padding(0, 1)
	b.WriteString(sectionHeaderStyle.Render("General Options"))
	b.WriteString("\n\n")

	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, false, true, false)
	explainStyle := lipgloss.NewStyle().MarginBottom(1)
	apiKey := lipgloss.JoinVertical(lipgloss.Top,
		headerStyle.Render("API Key"),
		explainStyle.Render("Your API key is used to authenticate with the API. You can find your API key in your account settings."),
		m.apiKey.View(),
	)
	b.WriteString(apiKey)
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
