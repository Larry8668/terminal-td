package waves

// WaveDef is the JSON-serializable wave definition.
type WaveDef struct {
	Wave   int             `json:"wave"`
	Groups []SpawnGroupDef `json:"groups"`
}

// SpawnGroupDef defines a group of enemies to spawn from a specific spawn point.
type SpawnGroupDef struct {
	SpawnID    string  `json:"spawn_id"`
	EnemyType  string  `json:"enemy_type"`
	Count      int     `json:"count"`
	Interval   float64 `json:"interval"`
	StartDelay float64 `json:"start_delay"`
}

// ActiveSpawnGroup tracks spawning progress for a group.
type ActiveSpawnGroup struct {
	Def        SpawnGroupDef
	Spawned    int
	Timer      float64
	DelayTimer float64
	Completed  bool
}

// WaveManager manages wave definitions and active spawning groups.
type WaveManager struct {
	Waves        []WaveDef
	CurrentWave  int
	ActiveGroups []ActiveSpawnGroup
	EnemiesAlive int
}

// NewWaveManager creates a wave manager from wave definitions.
func NewWaveManager(waves []WaveDef) *WaveManager {
	return &WaveManager{
		Waves:        waves,
		CurrentWave:  0,
		ActiveGroups: nil,
		EnemiesAlive: 0,
	}
}

// StartWave initializes spawning groups for the current wave.
func (wm *WaveManager) StartWave() {
	if wm.CurrentWave < 0 || wm.CurrentWave >= len(wm.Waves) {
		return
	}
	wave := wm.Waves[wm.CurrentWave]
	wm.ActiveGroups = make([]ActiveSpawnGroup, len(wave.Groups))
	for i, group := range wave.Groups {
		wm.ActiveGroups[i] = ActiveSpawnGroup{
			Def:        group,
			Spawned:    0,
			Timer:      0,
			DelayTimer: group.StartDelay,
			Completed:  false,
		}
	}
}

// IsWaveComplete returns true if all groups are done and no enemies are alive.
func (wm *WaveManager) IsWaveComplete() bool {
	if len(wm.ActiveGroups) == 0 {
		return false
	}
	for _, group := range wm.ActiveGroups {
		if !group.Completed {
			return false
		}
	}
	return wm.EnemiesAlive == 0
}

// NextWave advances to the next wave and clears active groups.
func (wm *WaveManager) NextWave() bool {
	wm.CurrentWave++
	wm.ActiveGroups = nil
	wm.EnemiesAlive = 0
	if wm.CurrentWave >= len(wm.Waves) {
		return false
	}
	return true
}
