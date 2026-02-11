package mapdata

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
)

//go:embed data/*.json
var defaultMapFS embed.FS

// MapInfo holds map metadata for selection.
type MapInfo struct {
	ID   string
	Name string
}

// ListMaps returns all available embedded maps.
func ListMaps() ([]MapInfo, error) {
	entries, err := defaultMapFS.ReadDir("data")
	if err != nil {
		return nil, fmt.Errorf("read maps dir: %w", err)
	}
	var maps []MapInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := defaultMapFS.ReadFile("data/" + e.Name())
		if err != nil {
			continue
		}
		m, err := LoadMapBytes(data)
		if err != nil {
			continue
		}
		maps = append(maps, MapInfo{ID: m.ID, Name: m.Name})
	}
	return maps, nil
}

// LoadMapByID loads a map by its ID from embedded maps.
func LoadMapByID(id string) (*GameMap, error) {
	entries, err := defaultMapFS.ReadDir("data")
	if err != nil {
		return nil, fmt.Errorf("read maps dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := defaultMapFS.ReadFile("data/" + e.Name())
		if err != nil {
			continue
		}
		m, err := LoadMapBytes(data)
		if err != nil {
			continue
		}
		if m.ID == id {
			return m, nil
		}
	}
	return nil, fmt.Errorf("map %q not found", id)
}

// DefaultMap returns the built-in classic map (same layout as legacy hardcoded map).
func DefaultMap() (*GameMap, error) {
	return LoadMapByID("classic")
}

// LoadMapBytes parses map JSON from bytes (for embed or tests).
func LoadMapBytes(data []byte) (*GameMap, error) {
	return LoadMap(bytes.NewReader(data))
}
