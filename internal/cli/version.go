package cli

import (
	"fmt"

	"terminal-td/internal/game"
)

// RunVersion prints the running build's version to stdout.
func RunVersion(args []string) int {
	fmt.Println(game.Version)
	return 0
}
