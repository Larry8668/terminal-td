package entities

import (
	"math"
	mapdata "terminal-td/internal/map"
)

type Enemy struct {
	X     float64
	Y     float64
	Speed float64

	PathIndex int
	Path      mapdata.Path

	ReachedBase bool
}

func NewEnemy(path mapdata.Path) *Enemy {
	start := path.Points[0]

	return &Enemy{
		X:     float64(start.X),
		Y:     float64(start.Y),
		Speed: 5,
		Path:  path,
	}
}

func (e *Enemy) Update(dt float64) {
	if e.PathIndex >= len(e.Path.Points) {
		e.ReachedBase = true
		return
	}
	if e.PathIndex >= len(e.Path.Points) {
		return
	}

	target := e.Path.Points[e.PathIndex]

	dx := float64(target.X) - e.X
	dy := float64(target.Y) - e.Y

	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 0.1 {
		e.PathIndex++
		return
	}

	e.X += (dx / dist) * e.Speed * dt
	e.Y += (dy / dist) * e.Speed * dt
}
