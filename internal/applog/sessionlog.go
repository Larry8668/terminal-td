// Package applog centralizes session log file creation so every entry point
// (the game's play loop, the MCP server, etc.) logs to disk instead of
// stdout/stderr. This matters most for `serve`, where stdout must carry only
// MCP protocol traffic.
package applog

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"terminal-td/internal/config"
)

const (
	sessionLogPrefix     = "terminal-td-session-"
	sessionLogSuffix     = ".log"
	maxSessionLogs       = 5
	sessionLogTimeFormat = "20060102-150405"
)

// InitSessionLog creates a new timestamped session log file under the user
// config dir and prunes old ones beyond maxSessionLogs. Callers are
// responsible for routing log output to it (log.SetOutput(f)) and closing it.
func InitSessionLog() (*os.File, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	cleanupSessionLogs(dir)
	name := sessionLogPrefix + time.Now().Format(sessionLogTimeFormat) + sessionLogSuffix
	path := filepath.Join(dir, name)
	return os.Create(path)
}

// cleanupSessionLogs removes the oldest session logs once more than
// maxSessionLogs are present, so the config dir doesn't grow unbounded.
func cleanupSessionLogs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, sessionLogPrefix) && strings.HasSuffix(n, sessionLogSuffix) {
			matches = append(matches, n)
		}
	}
	if len(matches) <= maxSessionLogs {
		return
	}
	sort.Strings(matches)
	for i := 0; i < len(matches)-maxSessionLogs; i++ {
		_ = os.Remove(filepath.Join(dir, matches[i]))
	}
}
