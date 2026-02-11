package mapdata

// MapDef is the JSON-serializable map definition (Option B: no tile matrix).
type MapDef struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Grid   GridDef    `json:"grid"`
	Spawns []SpawnDef `json:"spawns"`
	Paths  []PathDef  `json:"paths"`
	Base   BaseDef    `json:"base"`
}

type GridDef struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type SpawnDef struct {
	ID string `json:"id"`
	X  int    `json:"x"`
	Y  int    `json:"y"`
}

type PathDef struct {
	SpawnID string     `json:"spawn_id"`
	Points  []PointDef `json:"points"`
}

type PointDef struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type BaseDef struct {
	X  int `json:"x"`
	Y  int `json:"y"`
	HP int `json:"hp"`
}

// SpawnPoint is the runtime spawn (one per lane).
type SpawnPoint struct {
	ID string
	X  int
	Y  int
}

// GameMap is the loaded map: grid, spawns, paths by spawn_id, base.
type GameMap struct {
	ID     string
	Name   string
	Grid   *Grid
	Spawns []SpawnPoint
	Paths  map[string]Path // spawn_id -> path
	Base   BaseInfo
}

// BaseInfo is base position and HP (runtime).
type BaseInfo struct {
	X  int
	Y  int
	HP int
}

// PrimaryPath returns the path for the first spawn (single-lane / backward compat).
func (m *GameMap) PrimaryPath() Path {
	if len(m.Spawns) == 0 {
		return Path{}
	}
	return m.Paths[m.Spawns[0].ID]
}
