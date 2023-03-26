package exp

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func scrollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scrollback",
		Short: "experiment to see about clearing the screen buffer",
		RunE: func(cmd *cobra.Command, args []string) error {
			model := scrollback{}
			for i := 0; i < 100; i++ {
				model.lines = append(model.lines, fmt.Sprintf("Line %d", i))
			}
			p := tea.NewProgram(model)
			_, err := p.Run()
			return err
		},
	}
}

var _ tea.Model = scrollback{}

type scrollback struct {
	lines []string
}

func (m scrollback) Init() tea.Cmd {
	return nil
}

func (m scrollback) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ClearScrollback()
		return m, tea.Sequence(tea.ClearScreen)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, nil
}

type errMsg error

func ClearScrollback() tea.Msg {
	//fmt.Print("\033[3J\033[H\033[2J")
	fmt.Print("\033[3J")
	return errMsg(nil)
}

func (m scrollback) View() string {
	lines := strings.Join(m.lines, "\n")
	return lines + "\nprompt: "
}
