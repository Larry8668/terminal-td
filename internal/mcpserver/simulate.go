package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"terminal-td/internal/content"
	"terminal-td/internal/sim"
	waves "terminal-td/internal/waves"
)

// SimulateRunInput matches the shape committed to in Phase B, now backed by
// the real internal/sim engine.
type SimulateRunInput struct {
	MapID  string          `json:"map_id" jsonschema:"the map id to simulate"`
	Waves  []waves.WaveDef `json:"waves,omitempty" jsonschema:"optional wave override; defaults to the map's own waves"`
	Bot    string          `json:"bot,omitempty" jsonschema:"bot strategy: none|greedy (default none)"`
	Budget int             `json:"budget,omitempty" jsonschema:"starting money budget for the bot (0 = game default of 500)"`
}

// simulateRun implements the simulate_run tool: a fast, deterministic,
// headless playthrough of a map+waves pair with a scripted bot. Like
// validate_map, an in-simulation loss is a normal result reported via
// sim.Result.Outcome, not a tool error — a Go error here means the request
// itself was malformed (bad map id, invalid waves override).
func simulateRun(_ context.Context, _ *mcp.CallToolRequest, in SimulateRunInput) (*mcp.CallToolResult, sim.Result, error) {
	m, err := content.LoadMapByID(in.MapID)
	if err != nil {
		return nil, sim.Result{}, fmt.Errorf("load map %q: %w", in.MapID, err)
	}

	bot, err := sim.NewBot(in.Bot)
	if err != nil {
		return nil, sim.Result{}, err
	}

	res, err := sim.Run(sim.Config{
		Map:    m,
		Waves:  in.Waves,
		Bot:    bot,
		Budget: in.Budget,
	})
	if err != nil {
		return nil, sim.Result{}, err
	}
	return nil, *res, nil
}
