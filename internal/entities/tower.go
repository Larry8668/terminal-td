package entities

import (
	"math"
)

type Tower struct {
	X, Y int
	Type TowerType

	Range    float64
	Damage   float64
	FireRate float64
	Cost     int

	Target   *Enemy
	Cooldown float64

	Symbol rune
	Color  int
}

type TowerType int

const (
	TowerBasic TowerType = iota
)

func NewTower(x, y int, towerType TowerType) *Tower {
	t := &Tower{
		X:        x,
		Y:        y,
		Type:     towerType,
		Cooldown: 0,
	}

	switch towerType {
	case TowerBasic:
		t.Range = 5.0
		t.Damage = 10.0
		t.FireRate = 1.0
		t.Cost = 50
		t.Symbol = 'T'
		t.Color = 3
	}

	return t
}

func (t *Tower) DistanceTo(x, y float64) float64 {
	dx := float64(t.X) - x
	dy := float64(t.Y) - y

	return math.Sqrt(dx*dx + dy*dy)
}
