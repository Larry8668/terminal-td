package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"terminal-td/internal/content"
	"terminal-td/internal/sim"
	"terminal-td/internal/waves"
)

// RunSimulate runs a headless simulation for a map and prints a JSON Result
// (internal/sim.Result) to stdout. Errors and usage problems go to stderr so
// stdout only ever carries the JSON result, making this safe to pipe into
// other tools (e.g. `jq`).
func RunSimulate(args []string) int {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	mapID := fs.String("map", "", "map id to simulate")
	wavesPath := fs.String("waves", "", "optional path to a wave definitions JSON file (defaults to the map's own waves)")
	budget := fs.Int("budget", 0, "starting money budget for the bot (0 = use the map/game default)")
	botName := fs.String("bot", "none", "bot strategy: none|greedy")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *mapID == "" {
		fmt.Fprintln(os.Stderr, "terminal-td simulate: --map is required")
		return 2
	}

	m, err := content.LoadMapByID(*mapID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal-td simulate: load map %q: %v\n", *mapID, err)
		return 1
	}

	var waveOverride []waves.WaveDef
	if *wavesPath != "" {
		data, err := os.ReadFile(*wavesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "terminal-td simulate: read waves file %q: %v\n", *wavesPath, err)
			return 1
		}
		waveOverride, err = waves.LoadWaves(bytes.NewReader(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "terminal-td simulate: parse waves file %q: %v\n", *wavesPath, err)
			return 1
		}
	}

	bot, err := sim.NewBot(*botName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal-td simulate: %v\n", err)
		return 2
	}

	result, err := sim.Run(sim.Config{
		Map:    m,
		Waves:  waveOverride,
		Bot:    bot,
		Budget: *budget,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal-td simulate: %v\n", err)
		return 1
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "terminal-td simulate: encode result: %v\n", err)
		return 1
	}
	return 0
}
