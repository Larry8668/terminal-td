package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"terminal-td/internal/content"
)

// DeleteMapInput identifies the user map to delete.
type DeleteMapInput struct {
	ID string `json:"id" jsonschema:"the user map id to delete; built-in maps cannot be deleted"`
}

type DeleteMapOutput struct {
	Deleted bool `json:"deleted"`
}

func deleteMap(_ context.Context, _ *mcp.CallToolRequest, in DeleteMapInput) (*mcp.CallToolResult, DeleteMapOutput, error) {
	if err := content.DeleteMap(in.ID); err != nil {
		return nil, DeleteMapOutput{Deleted: false}, err
	}
	return nil, DeleteMapOutput{Deleted: true}, nil
}
