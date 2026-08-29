package cli

import (
	"flag"
	"fmt"
	"os"
)

// RunSimulate runs a headless simulation for a map and prints a JSON result
// to stdout. The flag surface matches the PLAN.md Phase C spec so the command
// signature is stable now; the engine itself (internal/sim) lands in Phase C.
func RunSimulate(args []string) int {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	mapID := fs.String("map", "", "map id to simulate")
	wavesPath := fs.String("waves", "", "optional path to a wave definitions JSON file (defaults to the map's own waves)")
	budget := fs.Int("budget", 0, "starting money budget for the bot (0 = use the map/game default)")
	bot := fs.String("bot", "none", "bot strategy: none|greedy")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	fmt.Fprintf(os.Stderr,
		"terminal-td simulate: not implemented yet (coming in Phase C) — requested map=%q waves=%q budget=%d bot=%q\n",
		*mapID, *wavesPath, *budget, *bot)
	return 1
}
