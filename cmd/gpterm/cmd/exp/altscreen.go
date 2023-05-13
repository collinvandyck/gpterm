package exp

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func altScreenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "alt",
		Short: "Run alt screen experiment",
		RunE: func(cmd *cobra.Command, args []string) error {
			model := altScreen{}
			p := tea.NewProgram(model)
			_, err := p.Run()
			return err
		},
	}
}

var _ tea.Model = altScreen{}

type altScreen struct {
	ta    textarea.Model
	ready bool
	alt   bool
}

// Init implements tea.Model
func (m altScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m altScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		screenCmd tea.Cmd
		taCmd     tea.Cmd
	)

	m.ta, taCmd = m.ta.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.ta = textarea.New()
			m.ta.Placeholder = "..."
			m.ta.Prompt = "┃ "
			m.ta.CharLimit = 4096
			m.ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
			m.ta.ShowLineNumbers = false
			m.ta.KeyMap.InsertNewline.SetEnabled(false)
		}
		m.ta.SetWidth(msg.Width)
		m.ta.SetHeight(3)
		m.ready = true
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			return m, tea.Quit
		case tea.KeyCtrlG:
			m.alt = !m.alt
			if m.alt {
				screenCmd = tea.EnterAltScreen
			} else {
				screenCmd = tea.Sequence(
					tea.ExitAltScreen,
					tea.ClearScreen,
				)
			}
		}
	}
	return m, tea.Sequence(taCmd, screenCmd)
}

// View implements tea.Model
func (m altScreen) View() string {
	if !m.ready {
		return ""
	}
	if m.alt {
		return "alt"
	}
	return m.ta.View()
}
