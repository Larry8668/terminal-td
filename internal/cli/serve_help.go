package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// resolveBinaryPath returns the path to show in mcpSetupHelp for "command".
//
// os.Executable() alone isn't quite enough: package managers like Homebrew
// run the binary through a stable symlink (e.g. /opt/homebrew/bin/terminal-td)
// that points into a versioned install directory (e.g.
// .../Cellar/terminal-td/1.2.3/bin/terminal-td). Per the os.Executable docs,
// which path it returns in that case is OS-dependent — if it resolves
// through to the versioned directory, a config a user saved today would
// silently break on their next `brew upgrade`. Prefer the stable,
// PATH-discoverable location when one exists and actually points at the
// same binary; only fall back to whatever os.Executable reported otherwise.
func resolveBinaryPath() string {
	execPath, err := os.Executable()
	if err != nil {
		return "/path/to/terminal-td"
	}

	base := filepath.Base(execPath)
	if onPath, err := exec.LookPath(base); err == nil && sameFile(onPath, execPath) {
		return onPath
	}
	return execPath
}

// sameFile reports whether a and b resolve (through any symlinks) to the
// same underlying file. Used to confirm a PATH match isn't a coincidentally
// same-named, unrelated binary before preferring it.
func sameFile(a, b string) bool {
	resolvedA, errA := filepath.EvalSymlinks(a)
	resolvedB, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return false
	}
	return resolvedA == resolvedB
}

// mcpToolNames lists the tools a correctly-connected client should see.
// Kept here (not imported from internal/mcpserver) so this help text has no
// dependency on the server package — it's pure, static, user-facing copy.
var mcpToolNames = []string{
	"get_game_rules", "list_maps", "get_map", "validate_map",
	"create_map", "create_waves", "delete_map", "simulate_run",
}

// mcpSetupHelp returns the message shown when `terminal-td serve` is run
// directly in an interactive terminal instead of being spawned by an MCP
// client. It deliberately does not hardcode instructions for any specific
// client (Cursor, Claude Desktop, Windsurf, ...) — config file formats and
// locations are numerous and change independently of this binary. Instead it
// states the two facts any stdio MCP client needs (the command and its args)
// as a prompt designed to be pasted straight into whatever AI coding agent
// the user already has, so that agent can find the right config file for
// itself and verify the connection — the same way a human would do it by
// hand.
func mcpSetupHelp(binaryPath string) string {
	toolList := ""
	for i, name := range mcpToolNames {
		if i > 0 {
			toolList += ", "
		}
		toolList += name
	}

	return fmt.Sprintf(`terminal-td: this looks like it was started directly in a terminal, not by an MCP client.

'terminal-td serve' speaks the Model Context Protocol (MCP) over stdio. It is
meant to be launched as a subprocess by an MCP-capable AI tool (Cursor,
Claude Desktop, Windsurf, etc.) rather than run standalone like this —
nothing happens if you leave it running this way; it just waits for a
client that was never going to connect.

To connect it, paste the following to your AI coding assistant:

----------------------------------------------------------------------
Set up a local MCP (Model Context Protocol) server for me:
  command: %s
  args: ["serve"]

Add this as a new MCP server entry in whatever config file my current
tool uses (for example Cursor's .cursor/mcp.json, or Claude Desktop's
claude_desktop_config.json), enable it, then verify the connection by
listing its tools. A working connection exposes these tools: %s.
----------------------------------------------------------------------

Exiting now — this process doesn't need to stay running for that step.
`, binaryPath, toolList)
}
