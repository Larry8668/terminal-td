package waves

import (
	"embed"
	"fmt"
	"log"
)

//go:embed data/*.json
var wavesFS embed.FS

// LoadWavesForMap loads waves for a specific map ID (e.g., "classic", "desert").
func LoadWavesForMap(mapID string) ([]WaveDef, error) {
	filename := fmt.Sprintf("data/%s.json", mapID)
	data, err := wavesFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("waves for map %q: %w", mapID, err)
	}
	waves, err := LoadWavesBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parse waves for map %q: %w", mapID, err)
	}
	log.Printf("loaded %d waves for map %q", len(waves), mapID)
	return waves, nil
}

// DefaultWaves returns waves for the classic map (backward compatibility).
func DefaultWaves() ([]WaveDef, error) {
	return LoadWavesForMap("classic")
}

// ValidateWavesAgainstMap checks that all spawn_ids in waves exist in the map.
func ValidateWavesAgainstMap(waves []WaveDef, spawnIDs map[string]bool) error {
	for _, wave := range waves {
		for _, group := range wave.Groups {
			if !spawnIDs[group.SpawnID] {
				return fmt.Errorf("wave %d group references spawn_id %q which doesn't exist in map", wave.Wave, group.SpawnID)
			}
		}
	}
	return nil
}
