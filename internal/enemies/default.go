package enemies

import (
	"embed"
)

//go:embed data/enemies.json
var defaultEnemiesFS embed.FS

// DefaultEnemies returns the built-in enemy definitions.
func DefaultEnemies() (*EnemyDatabase, error) {
	data, err := defaultEnemiesFS.ReadFile("data/enemies.json")
	if err != nil {
		return nil, err
	}
	return LoadEnemiesBytes(data)
}
