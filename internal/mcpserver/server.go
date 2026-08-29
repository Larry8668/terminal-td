// Package mcpserver implements terminal-td's MCP tools: a small "manual"
// (get_game_rules) plus map/wave CRUD and validation, all built on top of
// internal/content and internal/mapcheck. See docs/agent/PLAN.md Phase B for
// the full tool table and rationale.
package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"terminal-td/internal/buildinfo"
)

// NewServer builds an MCP server with every terminal-td tool registered.
// Callers are responsible for running it over a transport (see cli.RunServe).
func NewServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "terminal-td",
		Version: buildinfo.Version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_game_rules",
		Description: "Returns the game's rules manual: tile semantics, tower/enemy stats, wall mechanics, and the map/wave JSON schemas with examples. Call this first, before designing a map.",
	}, getGameRules)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_maps",
		Description: "Lists every available map (built-in and user-created), with id, name, source, grid size, and spawn count.",
	}, listMaps)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_map",
		Description: "Returns the full map definition (as authored, including any fork paths) and its waves (if any) for a given map id.",
	}, getMap)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "validate_map",
		Description: "Validates a map definition without saving it: structural checks (bounds, duplicate/dangling ids) plus a flow-field connectivity check per spawn. Returns valid/errors/warnings/path_lengths. Use this to iterate before create_map.",
	}, validateMap)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_map",
		Description: "Validates and saves a map definition to the user content directory, where it becomes playable in-game. Fails if invalid. Set overwrite=true to replace an existing user map with the same id.",
	}, createMap)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_waves",
		Description: "Validates and saves wave definitions for a map (structurally, and against that map's spawn ids), overriding its default waves.",
	}, createWaves)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "simulate_run",
		Description: "Runs a fast, deterministic headless simulation of a map+waves pair with a scripted bot (none|greedy) to check difficulty, without needing a human to play it. Returns outcome (won/lost/timeout), per-wave kills/leaks, and other stats. See get_game_rules' difficulty_guidance for target numbers.",
	}, simulateRun)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_map",
		Description: "Deletes a user-created map (and its waves override, if any). Built-in maps cannot be deleted.",
	}, deleteMap)

	return server
}
