package waves

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// LoadWaves reads wave definitions from r.
func LoadWaves(r io.Reader) ([]WaveDef, error) {
	var defs struct {
		Waves []WaveDef `json:"waves"`
	}
	if err := json.NewDecoder(r).Decode(&defs); err != nil {
		return nil, fmt.Errorf("wave decode: %w", err)
	}
	for i, wave := range defs.Waves {
		if wave.Wave <= 0 {
			return nil, fmt.Errorf("wave %d has invalid wave number %d", i, wave.Wave)
		}
		if len(wave.Groups) == 0 {
			return nil, fmt.Errorf("wave %d has no spawn groups", wave.Wave)
		}
		for j, group := range wave.Groups {
			if group.SpawnID == "" {
				return nil, fmt.Errorf("wave %d group %d has empty spawn_id", wave.Wave, j)
			}
			if group.EnemyType == "" {
				return nil, fmt.Errorf("wave %d group %d has empty enemy_type", wave.Wave, j)
			}
			if group.Count <= 0 {
				return nil, fmt.Errorf("wave %d group %d has invalid count %d", wave.Wave, j, group.Count)
			}
			if group.Interval <= 0 {
				return nil, fmt.Errorf("wave %d group %d has invalid interval %f", wave.Wave, j, group.Interval)
			}
		}
		log.Printf("loaded wave %d with %d groups", wave.Wave, len(wave.Groups))
	}
	return defs.Waves, nil
}

// LoadWavesBytes parses wave JSON from bytes (for embed or tests).
func LoadWavesBytes(data []byte) ([]WaveDef, error) {
	return LoadWaves(bytes.NewReader(data))
}
