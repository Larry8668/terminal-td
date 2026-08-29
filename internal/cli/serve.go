package cli

import (
	"fmt"
	"log"
	"os"

	"terminal-td/internal/applog"
)

// RunServe starts the MCP server over stdio. Nothing may ever be written to
// stdout here except MCP protocol traffic, so logging is routed to a session
// log file (same mechanism the play flow uses), never to stdout/stderr.
//
// Not implemented yet — the actual MCP wiring lands in Phase B. This stub
// exists so `terminal-td serve` is a stable, discoverable entry point and so
// the stdout-safety pattern (file logging) is already in place before real
// protocol traffic starts flowing through it.
func RunServe(args []string) int {
	f, err := applog.InitSessionLog()
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal-td serve: failed to init session log: %v\n", err)
	} else {
		defer f.Close()
		log.SetOutput(f)
	}

	log.Println("serve: MCP server requested (not implemented yet, Phase B)")
	fmt.Fprintln(os.Stderr, "terminal-td serve: not implemented yet (coming in Phase B)")
	return 1
}
