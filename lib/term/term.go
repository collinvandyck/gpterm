package term

import (
	"fmt"
)

func ClearScrollback() {
	// standard
	fmt.Print("\033[H\033[2J")

	// tmux
	fmt.Print("\033[3J")
}
