package mapdata

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
)

//go:embed data/*.json
var defaultMapFS embed.FS

// Source values for MapInfo, distinguishing built-in (embedded) maps from
// user-created ones layered in from the config dir (see internal/content).
const (
	SourceBuiltin = "builtin"
	SourceUser    = "user"
)

// MapInfo holds map metadata for selection and for the MCP list_maps tool.
type MapInfo struct {
	ID         string
	Name       string
	Source     string // SourceBuiltin or SourceUser
	Grid       GridDef
	SpawnCount int
}

// InfoFromMap extracts MapInfo summary fields from a loaded GameMap.
func InfoFromMap(m *GameMap, source string) MapInfo {
	return MapInfo{
		ID:         m.ID,
		Name:       m.Name,
		Source:     source,
		Grid:       GridDef{Width: m.Grid.Width, Height: m.Grid.Height},
		SpawnCount: len(m.Spawns),
	}
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
		maps = append(maps, InfoFromMap(m, SourceBuiltin))
	}
	return maps, nil
}

// embeddedMapFile returns the raw JSON bytes of the embedded map whose "id"
// field matches id, or an error if none match. Shared by LoadMapByID and
// LoadMapDefByID so the embed directory is only walked in one place.
func embeddedMapFile(id string) ([]byte, error) {
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
		def, err := LoadMapDefBytes(data)
		if err != nil {
			continue
		}
		if def.ID == id {
			return data, nil
		}
	}
	return nil, fmt.Errorf("map %q not found", id)
}

// LoadMapByID loads a map by its ID from embedded maps.
func LoadMapByID(id string) (*GameMap, error) {
	data, err := embeddedMapFile(id)
	if err != nil {
		return nil, err
	}
	return LoadMapBytes(data)
}

// LoadMapDefByID returns the raw MapDef (not a built GameMap) for an embedded
// map id, preserving every authored path entry — see LoadMapDefBytes.
func LoadMapDefByID(id string) (*MapDef, error) {
	data, err := embeddedMapFile(id)
	if err != nil {
		return nil, err
	}
	return LoadMapDefBytes(data)
}

// DefaultMap returns the built-in classic map (same layout as legacy hardcoded map).
func DefaultMap() (*GameMap, error) {
	return LoadMapByID("classic")
}

// LoadMapBytes parses map JSON from bytes (for embed or tests).
func LoadMapBytes(data []byte) (*GameMap, error) {
	return LoadMap(bytes.NewReader(data))
}
