package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"terminal-td/internal/content"
	mapdata "terminal-td/internal/map"
	waves "terminal-td/internal/waves"
)

// ListMapsInput is intentionally empty: list_maps takes no arguments.
type ListMapsInput struct{}

// GridSummary is a compact grid size for list_maps (full detail is in get_map).
type GridSummary struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// MapSummary is one entry in list_maps' output.
type MapSummary struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Source     string      `json:"source"` // "builtin" or "user"
	Grid       GridSummary `json:"grid"`
	SpawnCount int         `json:"spawn_count"`
}

type ListMapsOutput struct {
	Maps []MapSummary `json:"maps"`
}

func listMaps(_ context.Context, _ *mcp.CallToolRequest, _ ListMapsInput) (*mcp.CallToolResult, ListMapsOutput, error) {
	infos, err := content.ListMaps()
	if err != nil {
		return nil, ListMapsOutput{}, fmt.Errorf("list maps: %w", err)
	}
	out := ListMapsOutput{Maps: make([]MapSummary, 0, len(infos))}
	for _, m := range infos {
		out.Maps = append(out.Maps, MapSummary{
			ID:         m.ID,
			Name:       m.Name,
			Source:     m.Source,
			Grid:       GridSummary{Width: m.Grid.Width, Height: m.Grid.Height},
			SpawnCount: m.SpawnCount,
		})
	}
	return nil, out, nil
}

// GetMapInput identifies which map to fetch.
type GetMapInput struct {
	ID string `json:"id" jsonschema:"the map id, as returned by list_maps"`
}

// GetMapOutput returns the map exactly as authored (preserving fork path
// entries) plus its waves, if any exist.
type GetMapOutput struct {
	Map    mapdata.MapDef  `json:"map"`
	Source string          `json:"source"`
	Waves  []waves.WaveDef `json:"waves,omitempty"`
}

func getMap(_ context.Context, _ *mcp.CallToolRequest, in GetMapInput) (*mcp.CallToolResult, GetMapOutput, error) {
	if in.ID == "" {
		return nil, GetMapOutput{}, fmt.Errorf("id is required")
	}
	def, err := content.LoadMapDefByID(in.ID)
	if err != nil {
		return nil, GetMapOutput{}, fmt.Errorf("map %q not found: %w", in.ID, err)
	}

	source := mapdata.SourceBuiltin
	if infos, err := content.ListMaps(); err == nil {
		for _, m := range infos {
			if m.ID == in.ID {
				source = m.Source
				break
			}
		}
	}

	out := GetMapOutput{Map: *def, Source: source}
	if w, err := content.LoadWavesForMap(in.ID); err == nil {
		out.Waves = w
	}
	return nil, out, nil
}
