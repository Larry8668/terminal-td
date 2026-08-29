package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	mapdata "terminal-td/internal/map"
	"terminal-td/internal/mapcheck"
)

// ValidateMapInput carries an inline map definition to check, without
// writing anything to disk.
type ValidateMapInput struct {
	MapDef mapdata.MapDef `json:"map_def" jsonschema:"the map definition to validate"`
}

// validateMap implements the validate_map tool. It intentionally never
// returns a Go error for an invalid-but-well-formed request: "invalid" is a
// normal, expected result communicated via the output's valid/errors fields,
// not a tool failure.
func validateMap(_ context.Context, _ *mcp.CallToolRequest, in ValidateMapInput) (*mcp.CallToolResult, mapcheck.Result, error) {
	res := mapcheck.ValidateMap(&in.MapDef)
	return nil, *res, nil
}
