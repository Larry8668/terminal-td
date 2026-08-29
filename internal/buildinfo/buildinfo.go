// Package buildinfo holds values injected at build time via linker flags
// (see .goreleaser.yaml's ldflags), replacing the old hardcoded version
// const. Defaults here are what you get from `go build`/`go run` with no
// ldflags — i.e. a local dev build — so nothing breaks for development.
package buildinfo

// Version is the released version tag (e.g. "v0.2.0"). Overridden at build
// time via -X terminal-td/internal/buildinfo.Version={{.Version}}. Defaults
// to "dev" for local builds that didn't go through goreleaser.
var Version = "dev"

// Channel identifies how this binary was distributed, which changes
// self-update behavior (see internal/updater): "source" for local/dev
// builds, "github" for binaries downloaded directly from a GitHub release,
// "homebrew" for the Homebrew cask build. Overridden at build time via
// -X terminal-td/internal/buildinfo.Channel={{...}}.
var Channel = "source"

// IsHomebrew reports whether this binary should be treated as
// Homebrew-managed: either because it was built with Channel=homebrew
// (ldflags, set on the dedicated cask build), or — belt and suspenders,
// since ldflags can't retroactively fix already-distributed binaries or
// catch a user who manually moves a binary into a Cellar-like path — because
// the running executable's own path looks like a Homebrew install location.
// See docs/agent/DECISIONS.md for why both checks exist.
func IsHomebrew(executablePath string) bool {
	if Channel == "homebrew" {
		return true
	}
	return pathLooksLikeHomebrew(executablePath)
}
