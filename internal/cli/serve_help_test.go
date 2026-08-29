package cli

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestSameFile(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real-binary")
	if err := os.WriteFile(real, []byte("fake"), 0755); err != nil {
		t.Fatalf("write real binary: %v", err)
	}
	symlink := filepath.Join(dir, "symlink-to-real")
	if err := os.Symlink(real, symlink); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	unrelated := filepath.Join(dir, "unrelated-binary")
	if err := os.WriteFile(unrelated, []byte("different"), 0755); err != nil {
		t.Fatalf("write unrelated binary: %v", err)
	}

	if !sameFile(real, symlink) {
		t.Error("expected a symlink and its target to be the same file")
	}
	if sameFile(real, unrelated) {
		t.Error("expected two distinct files to not be the same file")
	}
	if sameFile(real, filepath.Join(dir, "does-not-exist")) {
		t.Error("expected a nonexistent path to never match")
	}
}

// TestResolveBinaryPathPrefersStableSymlink simulates the Homebrew scenario:
// the running binary's real (post os.Executable-style resolution) location
// is a versioned path, but a stable symlink pointing at the same file exists
// somewhere on PATH. resolveBinaryPath doesn't call os.Executable() itself
// (that's only meaningful for the actual running test binary), so this test
// exercises sameFile + the PATH-lookup logic directly via a temp PATH,
// mirroring what resolveBinaryPath does internally.
func TestResolveBinaryPathPrefersStableSymlinkLogic(t *testing.T) {
	dir := t.TempDir()
	versioned := filepath.Join(dir, "cellar", "1.2.3", "terminal-td")
	if err := os.MkdirAll(filepath.Dir(versioned), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(versioned, []byte("fake"), 0755); err != nil {
		t.Fatalf("write versioned binary: %v", err)
	}

	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	stableSymlink := filepath.Join(binDir, "terminal-td")
	if err := os.Symlink(versioned, stableSymlink); err != nil {
		t.Fatalf("create stable symlink: %v", err)
	}

	t.Setenv("PATH", binDir)

	// This mirrors resolveBinaryPath's body without depending on
	// os.Executable(), which always points at the real `go test` binary.
	base := filepath.Base(versioned)
	onPath, err := exec.LookPath(base)
	if err != nil {
		t.Fatalf("LookPath: %v", err)
	}
	if !sameFile(onPath, versioned) {
		t.Fatal("expected the PATH-resolved symlink to match the versioned binary")
	}
	if onPath != stableSymlink {
		t.Fatalf("expected to resolve to the stable symlink %q, got %q", stableSymlink, onPath)
	}
}
