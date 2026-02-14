package entities

import (
	"math"
	mapdata "terminal-td/internal/map"
)

type Enemy struct {
	X     float64
	Y     float64
	Speed float64
	HP    float64
	MaxHP float64

	PathIndex int
	Path      mapdata.Path

	Reward      int
	EnemyTypeID string

	ReachedBase bool
}

const reachedBaseDist = 0.5

// NewEnemy creates a basic enemy (legacy compatibility).
func NewEnemy(path mapdata.Path) *Enemy {
	start := path.Points[0]

	return &Enemy{
		X:           float64(start.X),
		Y:           float64(start.Y),
		Speed:       5,
		HP:          20.0,
		MaxHP:       20.0,
		Path:        path,
		Reward:      10,
		EnemyTypeID: "basic",
	}
}

// NewEnemyFromDef creates an enemy from a definition and path.
func NewEnemyFromDef(hp, speed float64, reward int, enemyTypeID string, path mapdata.Path) *Enemy {
	start := path.Points[0]

	return &Enemy{
		X:           float64(start.X),
		Y:           float64(start.Y),
		Speed:       speed,
		HP:          hp,
		MaxHP:       hp,
		Path:        path,
		Reward:      reward,
		EnemyTypeID: enemyTypeID,
	}
}

// Update moves the enemy along waypoints (legacy path-based).
func (e *Enemy) Update(dt float64) {
	if e.PathIndex >= len(e.Path.Points) {
		e.ReachedBase = true
		return
	}

	target := e.Path.Points[e.PathIndex]

	dx := float64(target.X) - e.X
	dy := float64(target.Y) - e.Y

	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 0.1 {
		e.X = float64(target.X)
		e.Y = float64(target.Y)
		e.PathIndex++
		return
	}

	moveDist := e.Speed * dt

	if moveDist > dist {
		moveDist = dist
	}

	e.X += (dx / dist) * moveDist
	e.Y += (dy / dist) * moveDist
}

// UpdateFlow moves the enemy along the flow field direction (no path memory).
// Caller should set ReachedBase when distance from flow field is < reachedBaseDist.
func (e *Enemy) UpdateFlow(dt float64, dirX, dirY float64) {
	moveDist := e.Speed * dt
	e.X += dirX * moveDist
	e.Y += dirY * moveDist
}
