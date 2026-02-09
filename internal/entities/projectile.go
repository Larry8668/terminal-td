package entities

import (
	"math"
)

type Projectile struct {
	X, Y        float64
	TargetX     float64
	TargetY     float64
	TargetEnemy *Enemy
	Speed       float64
	Damage      float64
	HasHit      bool
}

func NewProjectile(startX, startY float64, targetEnemy *Enemy, speed, damage float64) *Projectile {
	return &Projectile{
		X:           startX,
		Y:           startY,
		TargetX:     targetEnemy.X,
		TargetY:     targetEnemy.Y,
		TargetEnemy: targetEnemy,
		Speed:       speed,
		Damage:      damage,
		HasHit:      false,
	}
}

func (p *Projectile) Update(dt float64) {
	if p.HasHit {
		return
	}

	if p.TargetEnemy == nil || p.TargetEnemy.HP <= 0 {
		p.HasHit = true
		return
	}

	p.TargetX = p.TargetEnemy.X
	p.TargetY = p.TargetEnemy.Y

	dx := p.TargetX - p.X
	dy := p.TargetY - p.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 0.8 {
		p.X = p.TargetX
		p.Y = p.TargetY
		p.HasHit = true
		return
	}

	moveDist := p.Speed * dt
	if moveDist > dist {
		moveDist = dist
	}

	p.X += (dx / dist) * moveDist
	p.Y += (dy / dist) * moveDist
}
