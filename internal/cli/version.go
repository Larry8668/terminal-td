package cli

import (
	"fmt"

	"terminal-td/internal/buildinfo"
)

// RunVersion prints the running build's version to stdout.
func RunVersion(args []string) int {
	fmt.Println(buildinfo.Version)
	return 0
}
