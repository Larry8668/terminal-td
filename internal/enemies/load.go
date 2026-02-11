package enemies

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// LoadEnemies reads enemy definitions from r and returns a database.
func LoadEnemies(r io.Reader) (*EnemyDatabase, error) {
	var defs struct {
		Enemies []EnemyDef `json:"enemies"`
	}
	if err := json.NewDecoder(r).Decode(&defs); err != nil {
		return nil, fmt.Errorf("enemy decode: %w", err)
	}
	db := &EnemyDatabase{
		Enemies: make(map[string]EnemyDef),
	}
	for _, def := range defs.Enemies {
		if def.ID == "" {
			return nil, fmt.Errorf("enemy with empty id")
		}
		if def.HP <= 0 {
			return nil, fmt.Errorf("enemy %q has invalid hp %f", def.ID, def.HP)
		}
		if def.Speed <= 0 {
			return nil, fmt.Errorf("enemy %q has invalid speed %f", def.ID, def.Speed)
		}
		if def.Size <= 0 {
			return nil, fmt.Errorf("enemy %q has invalid size %d", def.ID, def.Size)
		}
		if _, ok := db.Enemies[def.ID]; ok {
			return nil, fmt.Errorf("duplicate enemy id %q", def.ID)
		}
		db.Enemies[def.ID] = def
		log.Printf("loaded enemy: id=%q name=%q hp=%.1f speed=%.1f size=%d reward=%d",
			def.ID, def.Name, def.HP, def.Speed, def.Size, def.Reward)
	}
	return db, nil
}

// LoadEnemiesBytes parses enemy JSON from bytes (for embed or tests).
func LoadEnemiesBytes(data []byte) (*EnemyDatabase, error) {
	return LoadEnemies(bytes.NewReader(data))
}
