package game

import (
	"log"
)

type GameState int
type InteractionMode int

const (
	StateMenu GameState = iota
	StatePreWave
	StateInWave
	StatePaused
	StateWon
	StateLost
	StateQuitConfirm
)

const (
	ModeNormal InteractionMode = iota
	ModeBuild
	ModeSelect
)

type GameManager struct {
	State GameState
	Mode  InteractionMode

	SelectedTowerX            int
	SelectedTowerY            int
	SelectingWallTarget       bool
	SelectingWallRemoveTarget bool

	CurrentWave int
	TotalWaves  int

	InterWaveTimer float64
	InterWaveDelay float64

	RunTime float64

	Paused bool
}

func NewGameManager(totalWaves int, interWaveDelay float64) *GameManager {
	return &GameManager{
		State:          StateMenu,
		Mode:           ModeNormal,
		SelectedTowerX: -1,
		SelectedTowerY: -1,
		TotalWaves:     totalWaves,
		InterWaveDelay: interWaveDelay,
		InterWaveTimer: 5,
	}
}

func (m *GameManager) IsSimulationRunning() bool {
	return m.State == StateInWave && !m.Paused
}

func (m *GameManager) Update(dt float64) {
	switch m.State {
	case StatePreWave:
		m.InterWaveTimer -= dt
		if m.InterWaveTimer <= 0 {
			m.StartWave()
		}
	case StateInWave:
		m.RunTime += dt
	}
}

func (m *GameManager) StartWave() {
	m.State = StateInWave
	m.CurrentWave++
	log.Printf("DEBUG: Wave %d started", m.CurrentWave)
}

func (m *GameManager) EndWave() {
	if m.CurrentWave >= m.TotalWaves {
		m.State = StateWon
		log.Printf("DEBUG: All waves completed - game won! Run time: %.2fs", m.RunTime)
		return
	}

	m.State = StatePreWave
	m.InterWaveTimer = m.InterWaveDelay
	log.Printf("DEBUG: Wave %d completed, waiting for next wave (Timer: %.1fs)", m.CurrentWave, m.InterWaveTimer)
}

func (m *GameManager) OnBaseDestroyed() {
	m.State = StateLost
	log.Printf("DEBUG: Base destroyed - game lost! Run time: %.2fs, Wave: %d/%d", m.RunTime, m.CurrentWave, m.TotalWaves)
}

func (m *GameManager) TogglePause() {
	if m.State != StateInWave && m.State != StatePaused {
		return
	}

	m.Paused = !m.Paused

	if m.Paused {
		m.State = StatePaused
		log.Println("Pausing game")
	} else {
		m.State = StateInWave
		log.Println("Unpausing game")
	}
}

func (m *GameManager) Reset() {
	m.State = StatePreWave
	m.Mode = ModeNormal
	m.SelectedTowerX = -1
	m.SelectedTowerY = -1
	m.SelectingWallTarget = false
	m.SelectingWallRemoveTarget = false
	m.CurrentWave = 0
	m.RunTime = 0
	m.InterWaveTimer = m.InterWaveDelay
	m.Paused = false
	log.Println("DEBUG: Game reset")
}
