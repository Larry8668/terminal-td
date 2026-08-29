package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	_ "embed"

	"terminal-td/internal/enemies"
	"terminal-td/internal/entities"
	"terminal-td/internal/game"
)

//go:embed examples/map_example.json
var mapExampleJSON []byte

//go:embed examples/waves_example.json
var wavesExampleJSON []byte

// GetGameRulesInput is intentionally empty: get_game_rules takes no arguments.
type GetGameRulesInput struct{}

// TileRule describes what a single tile type means to the model.
type TileRule struct {
	Tile        string `json:"tile"`
	Description string `json:"description"`
}

// WallRules documents wall mechanics (see internal/game.AddWall/RecomputeFlow).
type WallRules struct {
	LinkRadiusManhattan int    `json:"link_radius_manhattan"`
	SellRefundPercent   int    `json:"sell_refund_percent"`
	Description         string `json:"description"`
}

// TowerRule documents one buildable tower type (from game.GetTowerTemplates).
type TowerRule struct {
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Cost     int     `json:"cost"`
	Range    float64 `json:"range"`
	Damage   float64 `json:"damage"`
	FireRate float64 `json:"fire_rate"`
}

// EnemyRule documents one enemy type (from enemies.json).
type EnemyRule struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	HP     float64 `json:"hp"`
	Speed  float64 `json:"speed"`
	Size   int     `json:"size"`
	Reward int     `json:"reward"`
}

// SchemaDoc pairs a short description with a real, valid example so the model
// doesn't have to guess field shapes. Example is map[string]any (not
// json.RawMessage or bare `any`): the SDK infers an output JSON Schema from
// this struct's shape via reflection, and json.RawMessage would be inferred
// as a JSON array (wrong — the examples are objects), while bare `any` is
// reflected as an unconstrained `true` schema, which some MCP clients reject
// as non-portable (confirmed via `mcp-inspector --cli ... --strict`).
// map[string]any gives a proper "type": "object" schema instead.
type SchemaDoc struct {
	Description string         `json:"description"`
	Example     map[string]any `json:"example"`
}

// mustParseExample unmarshals an embedded example JSON object file into a
// map[string]any so it can populate SchemaDoc.Example. Panics on malformed
// embedded JSON, which would be a build-time bug (the file is ours), not a
// runtime condition.
func mustParseExample(data []byte) map[string]any {
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		panic(fmt.Sprintf("mcpserver: malformed embedded example JSON: %v", err))
	}
	return v
}

// GetGameRulesOutput is the model's "manual": everything needed to design a
// valid, appropriately difficult map+waves pair without trial and error
// against validate_map.
type GetGameRulesOutput struct {
	GameVersion        string      `json:"game_version"`
	Tiles              []TileRule  `json:"tiles"`
	Walls              WallRules   `json:"walls"`
	Towers             []TowerRule `json:"towers"`
	EnemyTypes         []EnemyRule `json:"enemy_types"`
	MapSchema          SchemaDoc   `json:"map_schema"`
	WaveSchema         SchemaDoc   `json:"wave_schema"`
	DifficultyGuidance string      `json:"difficulty_guidance"`
	Notes              []string    `json:"notes"`
}

func towerTypeName(t entities.TowerType) string {
	switch t {
	case entities.TowerBasic:
		return "basic"
	default:
		return fmt.Sprintf("type_%d", int(t))
	}
}

func buildTowerRules() []TowerRule {
	templates := game.GetTowerTemplates()
	rules := make([]TowerRule, 0, len(templates))
	for towerType, tmpl := range templates {
		rules = append(rules, TowerRule{
			Type:     towerTypeName(towerType),
			Name:     tmpl.Name,
			Cost:     tmpl.Cost,
			Range:    tmpl.Range,
			Damage:   tmpl.Damage,
			FireRate: tmpl.FireRate,
		})
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Type < rules[j].Type })
	return rules
}

func buildEnemyRules() ([]EnemyRule, error) {
	db, err := enemies.DefaultEnemies()
	if err != nil {
		return nil, fmt.Errorf("load enemy defs: %w", err)
	}
	rules := make([]EnemyRule, 0, len(db.Enemies))
	for _, def := range db.Enemies {
		rules = append(rules, EnemyRule{
			ID: def.ID, Name: def.Name, HP: def.HP, Speed: def.Speed, Size: def.Size, Reward: def.Reward,
		})
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	return rules, nil
}

// getGameRules implements the get_game_rules tool.
func getGameRules(_ context.Context, _ *mcp.CallToolRequest, _ GetGameRulesInput) (*mcp.CallToolResult, GetGameRulesOutput, error) {
	enemyRules, err := buildEnemyRules()
	if err != nil {
		return nil, GetGameRulesOutput{}, err
	}

	out := GetGameRulesOutput{
		GameVersion: game.Version,
		Tiles: []TileRule{
			{Tile: "empty", Description: "Buildable ground. Towers can be placed here; enemies never walk on it."},
			{Tile: "path", Description: "Enemies walk here toward the base, following the shortest route (flow field). Cannot build a tower here. Can be blocked by a wall between two towers, as long as at least one route from every spawn to the base remains open."},
			{Tile: "spawn", Description: "Where enemies for a given spawn_id enter the map. Always walkable; never blockable."},
			{Tile: "base", Description: "The player's base. Every enemy that reaches it removes 1 HP from base.hp; base.hp reaching 0 ends the game."},
		},
		Walls: WallRules{
			LinkRadiusManhattan: game.MaxWallLinkDist,
			SellRefundPercent:   game.SellRefundPercent,
			Description: fmt.Sprintf(
				"A wall links two towers whose Manhattan distance apart is <= %d. It blocks every 'path' tile on the straight (Bresenham) line between them; 'spawn' and 'base' tiles are never blocked. A wall (or tower sale, which removes its walls) is rejected if it would leave any spawn with no route to the base — map design does not need to defend against players fully sealing a lane, the game already prevents it at runtime.",
				game.MaxWallLinkDist),
		},
		Towers:     buildTowerRules(),
		EnemyTypes: enemyRules,
		MapSchema: SchemaDoc{
			Description: "A map is JSON with: id (unique string), name, grid{width,height}, spawns[{id,x,y}], paths[{spawn_id,points[{x,y}]}] (multiple path entries may share a spawn_id to create forks — all are drawn as walkable, but only one is used as the enemy's default route), base{x,y,hp}. All x/y must be within the grid. Validate with validate_map before calling create_map.",
			Example:     mustParseExample(mapExampleJSON),
		},
		WaveSchema: SchemaDoc{
			Description: "Waves are JSON with a top-level \"waves\" array. Each wave has a wave number (1-indexed, increasing) and groups[{spawn_id,enemy_type,count,interval,start_delay}]: spawn_id must exist on the target map, enemy_type must be one of enemy_types above, interval is seconds between spawns within the group, start_delay is seconds after the wave starts before the group begins spawning.",
			Example:     mustParseExample(wavesExampleJSON),
		},
		DifficultyGuidance: "Use simulate_run to check difficulty instead of guessing. Target for 'challenging but beatable': bot=greedy with the default budget (500) should WIN, ending with base HP at or below 30% of base_hp_start; bot=none should LOSE by wave 2-3 (a map with zero defense surviving much longer than that is too forgiving). Iterate: simulate with bot=none first to confirm the map applies real pressure, then bot=greedy to confirm a reasonable defense can still win. simulate_run is deterministic — same map_id/waves/bot/budget always produces the same result, so you can compare runs directly after each waves edit.",
		Notes: []string{
			"Call validate_map before create_map; create_map re-validates and will reject a structurally invalid or disconnected map anyway, but validate_map lets you iterate without writing files.",
			"create_map/create_waves only ever write to the user content directory; they never modify built-in maps. A user map with the same id as a built-in one shadows it (logged) rather than being rejected.",
			"delete_map only removes user-created maps; built-in maps cannot be deleted.",
		},
	}
	return nil, out, nil
}
