package game

import (
	"log"
)

type GameState int

const (
	StatePreWave GameState = iota
	StateInWave
	StatePaused
	StateWon
	StateLost
)

type GameManager struct {
	State GameState

	CurrentWave int
	TotalWaves  int

	InterWaveTimer float64
	InterWaveDelay float64

	RunTime float64

	Paused bool
}

func NewGameManager(totalWaves int, interWaveDelay float64) *GameManager {
	return &GameManager{
		State:          StatePreWave,
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
}

func (m *GameManager) EndWave() {
	if m.CurrentWave >= m.TotalWaves {
		m.State = StateWon
		return
	}

	m.State = StatePreWave
	m.InterWaveTimer = m.InterWaveDelay
}

func (m *GameManager) OnBaseDestroyed() {
	m.State = StateLost
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
	m.CurrentWave = 0
	m.RunTime = 0
	m.InterWaveTimer = m.InterWaveDelay
	m.Paused = false
}
