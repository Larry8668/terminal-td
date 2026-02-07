package game

import (
	"terminal-td/internal/entities"
	mapdata "terminal-td/internal/map"
)

type Game struct {
	Grid    *mapdata.Grid
	Path    mapdata.Path
	Enemies []*entities.Enemy
}

func NewGame() *Game {
	grid := mapdata.NewGrid(40, 20)
	path := mapdata.DefaultPath()
	mapdata.ApplyPath(grid, path)

	enemy := entities.NewEnemy(path)

	return &Game{
		Grid:    grid,
		Path:    path,
		Enemies: []*entities.Enemy{enemy},
	}
}

func (g *Game) Update(dt float64) {
	for _, e := range g.Enemies {
		e.Update(dt)
	}
}
