// Package mapcheck validates a map definition the same way the game loader
// and Game.WouldDisconnectSpawnsFromBase do, but standalone — without needing
// a running *game.Game. This makes it reusable by the MCP server's
// validate_map tool and by internal/content's SaveMap without either of them
// depending on the game package.
package mapcheck

import (
	"fmt"

	"terminal-td/internal/flow"
	mapdata "terminal-td/internal/map"
)

// shortLaneWarningThreshold is the spawn→base path length (in tiles) below
// which we warn that a lane is very short. Matches the "very short lane"
// language documented for players via get_game_rules.
const shortLaneWarningThreshold = 5

// Result is the outcome of validating a map definition.
type Result struct {
	Valid       bool           `json:"valid"`
	Errors      []string       `json:"errors"`
	Warnings    []string       `json:"warnings"`
	PathLengths map[string]int `json:"path_lengths"` // spawn_id -> steps to base, only for reachable spawns
}

// ValidateMap runs the same structural validation the game's map loader does
// (bounds, duplicate ids, dangling spawn_id references, etc.) plus a flow-field
// connectivity check per spawn: every spawn must have a finite-distance route
// to the base, exactly like Game.WouldDisconnectSpawnsFromBase enforces for
// walls at runtime.
func ValidateMap(def *mapdata.MapDef) *Result {
	res := &Result{
		Errors:      []string{},
		Warnings:    []string{},
		PathLengths: map[string]int{},
	}

	m, err := mapdata.LoadMapDef(def)
	if err != nil {
		res.Errors = append(res.Errors, err.Error())
		return res
	}

	walkable := flow.BuildWalkability(m.Grid)
	field := flow.Compute(m.Grid.Width, m.Grid.Height, walkable, m.Base.X, m.Base.Y)

	for _, spawn := range m.Spawns {
		dist, _ := field.At(spawn.X, spawn.Y)
		if dist >= flow.Inf {
			res.Errors = append(res.Errors, fmt.Sprintf("spawn %q at (%d,%d) has no path to base", spawn.ID, spawn.X, spawn.Y))
			continue
		}
		steps := int(dist)
		res.PathLengths[spawn.ID] = steps
		if steps < shortLaneWarningThreshold {
			res.Warnings = append(res.Warnings, fmt.Sprintf("spawn %q is only %d tile(s) from base (very short lane)", spawn.ID, steps))
		}
	}

	res.Valid = len(res.Errors) == 0
	return res
}
