// Package content merges user-created maps/waves (under the OS config dir)
// with the game's embedded, built-in ones. User content shadows a built-in
// map/waves file of the same id and is logged when it does, per
// docs/agent/DECISIONS.md (2026-08-30, "user maps shadow embedded").
package content

import (
	"os"
	"path/filepath"

	"terminal-td/internal/config"
)

const (
	mapsSubdir  = "maps"
	wavesSubdir = "waves"
)

// MapsDir returns "<config dir>/maps", creating it if needed.
func MapsDir() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, mapsSubdir)
	if err := os.MkdirAll(d, 0755); err != nil {
		return "", err
	}
	return d, nil
}

// WavesDir returns "<config dir>/waves", creating it if needed.
func WavesDir() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, wavesSubdir)
	if err := os.MkdirAll(d, 0755); err != nil {
		return "", err
	}
	return d, nil
}

// wavesPathForMap returns the path a user waves override for mapID would live at.
func wavesPathForMap(mapID string) (string, error) {
	dir, err := WavesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, mapID+".json"), nil
}
