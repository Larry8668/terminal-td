package cli

import (
	"strings"
	"testing"
)

func TestMCPSetupHelpContainsCommandAndTools(t *testing.T) {
	msg := mcpSetupHelp("/opt/homebrew/bin/terminal-td")

	if !strings.Contains(msg, "/opt/homebrew/bin/terminal-td") {
		t.Fatal("expected the help message to include the actual binary path")
	}
	if !strings.Contains(msg, `args: ["serve"]`) {
		t.Fatal("expected the help message to include the serve arg")
	}
	for _, tool := range mcpToolNames {
		if !strings.Contains(msg, tool) {
			t.Fatalf("expected the help message to mention tool %q", tool)
		}
	}
	// Must not hardcode a specific MCP client's config format/location —
	// that's the whole point of delegating to the user's own AI agent.
	for _, vendorTerm := range []string{"mcpServers\":", ".json\":", "~/.config"} {
		if strings.Contains(msg, vendorTerm) {
			t.Fatalf("help message should not hardcode client-specific config syntax, found %q", vendorTerm)
		}
	}
}
