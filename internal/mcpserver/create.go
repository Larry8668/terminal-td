package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"terminal-td/internal/content"
	mapdata "terminal-td/internal/map"
	waves "terminal-td/internal/waves"
)

// CreateMapInput is a map definition to validate and persist to the user
// content directory.
type CreateMapInput struct {
	MapDef    mapdata.MapDef `json:"map_def" jsonschema:"the map definition to save"`
	Overwrite bool           `json:"overwrite,omitempty" jsonschema:"set true to replace an existing user map with the same id"`
}

type CreateMapOutput struct {
	ID        string `json:"id"`
	SavedPath string `json:"saved_path"`
}

// createMap implements the create_map tool. It validates before writing
// (content.SaveMap never writes an invalid map) and, unlike validate_map, a
// failed validation here IS a tool error — the caller asked for a write to
// happen and it didn't.
func createMap(_ context.Context, _ *mcp.CallToolRequest, in CreateMapInput) (*mcp.CallToolResult, CreateMapOutput, error) {
	path, err := content.SaveMap(&in.MapDef, in.Overwrite)
	if err != nil {
		return nil, CreateMapOutput{}, err
	}
	return nil, CreateMapOutput{ID: in.MapDef.ID, SavedPath: path}, nil
}

// CreateWavesInput is a wave list to validate (structurally, and against the
// target map's spawns) and persist as that map's user waves override.
type CreateWavesInput struct {
	MapID string          `json:"map_id" jsonschema:"the map id these waves are for"`
	Waves []waves.WaveDef `json:"waves" jsonschema:"the wave definitions to save"`
}

type CreateWavesOutput struct {
	SavedPath string `json:"saved_path"`
}

func createWaves(_ context.Context, _ *mcp.CallToolRequest, in CreateWavesInput) (*mcp.CallToolResult, CreateWavesOutput, error) {
	path, err := content.SaveWaves(in.MapID, in.Waves)
	if err != nil {
		return nil, CreateWavesOutput{}, err
	}
	return nil, CreateWavesOutput{SavedPath: path}, nil
}
