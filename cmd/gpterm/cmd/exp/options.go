package exp

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func optionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "options",
		Short: "test out options rendering",
		RunE: func(cmd *cobra.Command, args []string) error {
			model := optionsModel{}
			p := tea.NewProgram(model,
				tea.WithAltScreen(),
			)
			_, err := p.Run()
			return err
		},
	}
}

type optionsModel struct {
	width  int
	height int
}

// Init implements tea.Model.
func (m optionsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m optionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m optionsModel) View() string {
	return "view"
}
