package flow

import (
	"math"
)

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Normalize() Vec2 {
	len := v.Len()
	if len < 1e-9 {
		return Vec2{}
	}

	return Vec2{X: v.X / len, Y: v.Y / len}
}

func (v Vec2) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}
