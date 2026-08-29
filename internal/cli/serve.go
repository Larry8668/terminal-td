package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"terminal-td/internal/applog"
	"terminal-td/internal/mcpserver"
)

// RunServe starts the MCP server over stdio. Nothing may ever be written to
// stdout here except MCP protocol traffic, so logging is routed to a session
// log file (same mechanism the play flow uses), never to stdout/stderr.
func RunServe(args []string) int {
	f, err := applog.InitSessionLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal-td serve: failed to init session log: %v\n", err)
	} else {
		defer f.Close()
		log.SetOutput(f)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("serve: starting MCP server over stdio")
	server := mcpserver.NewServer()
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("serve: server exited with error: %v", err)
		return 1
	}
	log.Println("serve: client disconnected, shutting down")
	return 0
}
