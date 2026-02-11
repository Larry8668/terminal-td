package enemies

type EnemyDef struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	HP     float64 `json:"hp"`
	Speed  float64 `json:"speed"`
	Size   int     `json:"size"`
	Reward int     `json:"reward"`
}

// EnemyDatabase holds all loaded enemy definitions.
type EnemyDatabase struct {
	Enemies map[string]EnemyDef // id -> definition
}

// Get returns an enemy definition by ID, or nil if not found.
func (db *EnemyDatabase) Get(id string) *EnemyDef {
	def, ok := db.Enemies[id]
	if !ok {
		return nil
	}
	return &def
}
