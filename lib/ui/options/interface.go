package options

import tea "github.com/charmbracelet/bubbletea"

type Interface interface {
	Init() tea.Cmd
	Update(tea.Msg) (Interface, tea.Cmd)
	View() string
}
