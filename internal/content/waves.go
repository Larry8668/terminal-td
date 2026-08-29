package content

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	waves "terminal-td/internal/waves"
)

// LoadWavesForMap returns waves for mapID, preferring a user override
// ("<waves dir>/<mapID>.json") over the map's embedded waves.
func LoadWavesForMap(mapID string) ([]waves.WaveDef, error) {
	path, err := wavesPathForMap(mapID)
	if err == nil {
		if data, readErr := os.ReadFile(path); readErr == nil {
			w, parseErr := waves.LoadWavesBytes(data)
			if parseErr != nil {
				log.Printf("content: invalid user waves for map %q (%s): %v, falling back to built-in waves", mapID, path, parseErr)
			} else {
				log.Printf("content: loaded user waves for map %q from %s", mapID, path)
				return w, nil
			}
		}
	}
	return waves.LoadWavesForMap(mapID)
}

// SaveWaves validates defs against mapID's spawns (loading the map through
// LoadMapByID so a user map override is respected) and against the wave
// schema itself, then writes them to "<waves dir>/<mapID>.json". Returns the
// path written.
func SaveWaves(mapID string, defs []waves.WaveDef) (string, error) {
	m, err := LoadMapByID(mapID)
	if err != nil {
		return "", fmt.Errorf("load map %q to validate waves: %w", mapID, err)
	}

	payload := struct {
		Waves []waves.WaveDef `json:"waves"`
	}{Waves: defs}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal waves for map %q: %w", mapID, err)
	}
	if _, err := waves.LoadWavesBytes(data); err != nil {
		return "", fmt.Errorf("waves for map %q failed validation: %w", mapID, err)
	}

	spawnIDs := make(map[string]bool, len(m.Spawns))
	for _, s := range m.Spawns {
		spawnIDs[s.ID] = true
	}
	if err := waves.ValidateWavesAgainstMap(defs, spawnIDs); err != nil {
		return "", fmt.Errorf("waves for map %q failed spawn validation: %w", mapID, err)
	}

	path, err := wavesPathForMap(mapID)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write waves for map %q: %w", mapID, err)
	}
	log.Printf("content: saved %d wave(s) for map %q to %s", len(defs), mapID, path)
	return path, nil
}
