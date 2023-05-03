package exp

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func lipglossCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lipgloss",
		Short: "test out lipgloss rendering",
		RunE: func(cmd *cobra.Command, args []string) error {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
			str := `
import "fmt"

func main() {
  fmt.Println("Hello, World!")
}
`
			out := style.Render(str)
			fmt.Println(out)
			s := bufio.NewScanner(strings.NewReader(str))
			for s.Scan() {
				t := s.Text()
				out := style.Render(t)
				fmt.Println(out)
			}

			fmt.Println()
			sentence := "Now is the time for all good men to come to the aid of the party."
			for i := 0; i < 2; i++ {
				sentence += " " + sentence
			}
			style = lipgloss.NewStyle().Width(80)
			fmt.Println(style.Render(sentence))

			style = lipgloss.NewStyle().Width(80)
			spanStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f81ce5"))

			fmt.Println()
			sentence = "Now is the time for " + spanStyle.Render("all") + " good men to come to the aid of the party."
			for i := 0; i < 2; i++ {
				sentence += " " + sentence
			}
			fmt.Println(style.Render(sentence))

			return nil
		},
	}
}
