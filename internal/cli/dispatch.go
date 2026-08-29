// Package cli implements terminal-td's subcommand dispatch: `serve`,
// `simulate`, and `version`. Running with no subcommand (or with a flag like
// --just-updated) falls through to the default play flow in cmd/game/main.go.
package cli

// Dispatch inspects args (os.Args[1:]) for a known subcommand name and, if
// found, runs it. handled=false means the caller should proceed with the
// normal play startup (this also covers plain flags such as --just-updated,
// which never match a subcommand name).
func Dispatch(args []string) (handled bool, exitCode int) {
	if len(args) == 0 {
		return false, 0
	}
	switch args[0] {
	case "serve":
		return true, RunServe(args[1:])
	case "simulate":
		return true, RunSimulate(args[1:])
	case "version":
		return true, RunVersion(args[1:])
	default:
		return false, 0
	}
}
