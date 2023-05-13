package exp

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/collinvandyck/gpterm/lib/mathx"
	"github.com/spf13/cobra"
)

func scrollCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scroll",
		Short: "test out scrolling options",
		RunE: func(cmd *cobra.Command, args []string) error {
			content := ""
			for i := 0; i < 1000; i++ {
				content += fmt.Sprintf("%d\n", i)
			}
			model := scroller{
				content: content,
			}
			p := tea.NewProgram(model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)
			_, err := p.Run()
			return err
		},
	}
}

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
	highPerformanceRendering = true
)
var _ tea.Model = scroller{}

type scroller struct {
	viewport viewport.Model
	content  string
	ready    bool
}

// Init implements tea.Model
func (scroller) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m scroller) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = highPerformanceRendering
			m.viewport.SetContent(m.content)
			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		}
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
		m.ready = true
		if highPerformanceRendering {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			return m, tea.Quit
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m scroller) View() string {
	if !m.ready {
		return ""
	}
	return strings.Join([]string{
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	}, "\n")
}

func (m scroller) headerView() string {
	title := titleStyle.Render("Mr. Pager")
	line := strings.Repeat("─", mathx.Max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m scroller) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", mathx.Max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
