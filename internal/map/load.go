package mapdata

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// LoadMap reads a map definition from r and builds a GameMap (grid from paths/spawns/base).
func LoadMap(r io.Reader) (*GameMap, error) {
	var def MapDef
	if err := json.NewDecoder(r).Decode(&def); err != nil {
		return nil, fmt.Errorf("map decode: %w", err)
	}
	return buildGameMap(&def)
}

// LoadMapFile reads a map from a JSON file path.
func LoadMapFile(path string) (*GameMap, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open map: %w", err)
	}
	defer f.Close()
	return LoadMap(f)
}

func buildGameMap(def *MapDef) (*GameMap, error) {
	if def.Grid.Width <= 0 || def.Grid.Height <= 0 {
		return nil, fmt.Errorf("invalid grid size %dx%d", def.Grid.Width, def.Grid.Height)
	}
	spawnByID := make(map[string]SpawnPoint)
	for _, s := range def.Spawns {
		if s.ID == "" {
			return nil, fmt.Errorf("spawn with empty id")
		}
		if _, ok := spawnByID[s.ID]; ok {
			return nil, fmt.Errorf("duplicate spawn id %q", s.ID)
		}
		spawnByID[s.ID] = SpawnPoint{ID: s.ID, X: s.X, Y: s.Y}
	}

	pathsBySpawn := make(map[string]Path)
	for _, pd := range def.Paths {
		if pd.SpawnID == "" {
			return nil, fmt.Errorf("path with empty spawn_id")
		}
		if _, ok := spawnByID[pd.SpawnID]; !ok {
			return nil, fmt.Errorf("path spawn_id %q not in spawns", pd.SpawnID)
		}
		if len(pd.Points) < 2 {
			return nil, fmt.Errorf("path for spawn %q has fewer than 2 points", pd.SpawnID)
		}
		points := make([]Point, len(pd.Points))
		for i, p := range pd.Points {
			points[i] = Point{X: p.X, Y: p.Y}
			if p.X < 0 || p.X >= def.Grid.Width || p.Y < 0 || p.Y >= def.Grid.Height {
				return nil, fmt.Errorf("path point (%d,%d) out of bounds", p.X, p.Y)
			}
		}
		pathsBySpawn[pd.SpawnID] = Path{Points: points}
	}

	if def.Base.X < 0 || def.Base.X >= def.Grid.Width || def.Base.Y < 0 || def.Base.Y >= def.Grid.Height {
		return nil, fmt.Errorf("base (%d,%d) out of bounds", def.Base.X, def.Base.Y)
	}
	if def.Base.HP <= 0 {
		return nil, fmt.Errorf("base hp must be positive, got %d", def.Base.HP)
	}

	grid := NewGrid(def.Grid.Width, def.Grid.Height)
	for _, pd := range def.Paths {
		path := pathsBySpawn[pd.SpawnID]
		applyPathToGrid(grid, path, def.Base.X, def.Base.Y)
	}

	spawns := make([]SpawnPoint, 0, len(def.Spawns))
	for _, s := range def.Spawns {
		spawns = append(spawns, spawnByID[s.ID])
	}

	log.Printf("map loaded: id=%s name=%q grid=%dx%d spawns=%d paths=%d base=(%d,%d) hp=%d",
		def.ID, def.Name, def.Grid.Width, def.Grid.Height, len(spawns), len(def.Paths), def.Base.X, def.Base.Y, def.Base.HP)

	return &GameMap{
		ID:     def.ID,
		Name:   def.Name,
		Grid:   grid,
		Spawns: spawns,
		Paths:  pathsBySpawn,
		Base:   BaseInfo{X: def.Base.X, Y: def.Base.Y, HP: def.Base.HP},
	}, nil
}
