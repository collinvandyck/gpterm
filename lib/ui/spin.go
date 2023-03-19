package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spin struct {
	uiOpts
	spinner  spinner.Model
	duration time.Duration
}

func newSpinner(uiOpts uiOpts) spin {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	res := spin{
		uiOpts:   uiOpts,
		spinner:  sp,
		duration: 500 * time.Millisecond,
	}
	return res
}

func (m spin) tick() tea.Cmd {
	return tea.Tick(m.duration, func(t time.Time) tea.Msg {
		return spinner.TickMsg{
			Time: t,
			ID:   1,
		}
	})
}

func (m spin) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spin) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if msg.ID == 1 {
			m.spinner, _ = m.spinner.Update(msg)
			return m, m.tick()
		}
	}
	return m, nil
}

func (m spin) View() string {
	return m.spinner.View()
}
