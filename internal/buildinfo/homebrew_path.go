package buildinfo

import "strings"

// pathLooksLikeHomebrew is the path-heuristic half of IsHomebrew: Homebrew
// installs casks under a Cellar or Caskroom directory (both macOS and
// Linuxbrew use these names), so a running binary whose own path contains
// either segment was almost certainly installed via `brew install --cask`.
func pathLooksLikeHomebrew(executablePath string) bool {
	return strings.Contains(executablePath, "/Cellar/") ||
		strings.Contains(executablePath, "/Caskroom/") ||
		strings.Contains(executablePath, "\\Cellar\\") ||
		strings.Contains(executablePath, "\\Caskroom\\")
}
