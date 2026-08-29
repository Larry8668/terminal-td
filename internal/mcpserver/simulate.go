package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	waves "terminal-td/internal/waves"
)

// SimulateRunInput matches the shape simulate_run will accept once
// internal/sim exists (Phase C). Accepting it now keeps the tool's contract
// stable across phases.
type SimulateRunInput struct {
	MapID  string          `json:"map_id" jsonschema:"the map id to simulate"`
	Waves  []waves.WaveDef `json:"waves,omitempty" jsonschema:"optional wave override; defaults to the map's own waves"`
	Bot    string          `json:"bot,omitempty" jsonschema:"bot strategy: none|greedy"`
	Budget int             `json:"budget,omitempty" jsonschema:"starting money budget for the bot"`
}

type SimulateRunOutput struct {
	Implemented bool   `json:"implemented"`
	Message     string `json:"message"`
}

// simulateRun is a stub for the simulate_run tool; the real headless
// simulator (internal/sim) lands in Phase C. It deliberately returns a
// successful, structured "not implemented" result rather than a tool error,
// so a model calling it gets a clear, parseable signal instead of a generic
// failure.
func simulateRun(_ context.Context, _ *mcp.CallToolRequest, _ SimulateRunInput) (*mcp.CallToolResult, SimulateRunOutput, error) {
	return nil, SimulateRunOutput{
		Implemented: false,
		Message:     "simulate_run is not implemented yet (coming in Phase C). Use validate_map for structural and connectivity checks in the meantime.",
	}, nil
}
