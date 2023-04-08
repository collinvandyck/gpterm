package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/collinvandyck/gpterm/lib/ui/gptea"
)

type statusModel struct {
	uiOpts
	spinner spinner.Model
	width   int
	spin    bool
	ready   bool
}

func newStatusModel(uiOpts uiOpts) statusModel {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	return statusModel{
		uiOpts: uiOpts,
		spinner: spinner.NewModel(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(style),
		),
	}
}

func (m statusModel) Init() tea.Cmd {
	return tea.Batch(m.tick())
}

func (m statusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		spinCmd tea.Cmd
		cmds    commands
	)
	switch msg := msg.(type) {

	case gptea.WindowSizeMsg:
		m.ready = msg.Ready
		m.width = msg.Width

	case gptea.StreamCompletionReq:
		m.spin = true
		cmds.Add(m.tick())

	case gptea.StreamCompletionResult:
		m.spin = false

	case spinner.TickMsg:
		if m.spin {
			m.spinner, _ = m.spinner.Update(msg)
			cmds.Add(m.tick())
		}
	}
	return m, cmds.BatchWith(spinCmd)
}

func (m statusModel) View() string {
	if !m.ready {
		return ""
	}
	spin := m.spinView()
	help := m.help(m.width - 1)
	return spin + help
}

func (m statusModel) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinner.Tick()
	})
}

func (m statusModel) spinView() string {
	return m.spinner.View()
}

func (m statusModel) help(width int) string {
	style := lipgloss.NewStyle().Background(lipgloss.Color("#222222")).Foreground(lipgloss.Color("#dddddd"))
	text := "↑/↓: History | <C-p> Prev Convo | <C-n> Next Convo | <C-c> Quit"
	return style.Width(width).Render(text)
}
