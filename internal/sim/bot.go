package sim

import (
	"fmt"

	"terminal-td/internal/game"
)

// Bot decides what to build during the pre-wave window, once per wave
// boundary, before that wave's enemies spawn. Implementations must be
// deterministic — same game state in, same decisions out — since that's
// what makes a Run's Result reproducible.
type Bot interface {
	PreWave(g *game.Game)
}

// NoneBot builds nothing. It's the baseline: how much raw pressure does a
// map's waves apply with zero defense? A map where NoneBot survives too long
// is probably too easy.
type NoneBot struct{}

// PreWave is a no-op: NoneBot never builds anything.
func (NoneBot) PreWave(*game.Game) {}

// NewBot resolves a bot strategy name to a Bot. Empty string and "none" both
// mean NoneBot. Returns an error for unrecognized names rather than silently
// falling back, so a typo in a CLI flag or MCP tool call surfaces clearly.
func NewBot(name string) (Bot, error) {
	switch name {
	case "", "none":
		return NoneBot{}, nil
	case "greedy":
		return GreedyBot{}, nil
	default:
		return nil, fmt.Errorf("sim: unknown bot strategy %q (expected \"none\" or \"greedy\")", name)
	}
}
