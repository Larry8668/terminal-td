package sim

import (
	"terminal-td/internal/entities"
	"terminal-td/internal/flow"
	"terminal-td/internal/game"
)

// GreedyBot spends its money before every wave on the tower placement that
// covers the most spawn→base traffic, repeating until it can no longer
// afford a tower or no remaining empty tile would cover any traffic at all.
type GreedyBot struct{}

// PreWave implements Bot.
func (GreedyBot) PreWave(g *game.Game) {
	template, ok := game.GetTowerTemplates()[entities.TowerBasic]
	if !ok {
		return
	}

	paths := tracedPaths(g)
	for g.Money >= template.Cost {
		x, y, coverage := bestTowerSpot(g, template.Range, paths)
		if x < 0 || coverage == 0 {
			break
		}
		g.CursorX, g.CursorY = x, y
		if !g.PlaceTower(entities.TowerBasic) {
			// CanPlaceTower already confirmed this tile in bestTowerSpot;
			// this would only trip on an unexpected engine change. Stop
			// rather than loop forever.
			break
		}
	}
}

// tracedPaths returns the flow-field route from every spawn to base, used to
// score how much enemy traffic a candidate tile would cover.
func tracedPaths(g *game.Game) [][]flow.Tile {
	if g.Map == nil || g.FlowField == nil {
		return nil
	}
	var paths [][]flow.Tile
	for _, spawn := range g.Map.Spawns {
		path := g.FlowField.TracePath(spawn.X, spawn.Y, g.Base.X, g.Base.Y)
		if len(path) > 0 {
			paths = append(paths, path)
		}
	}
	return paths
}

// bestTowerSpot scans every placeable tile and returns the one whose
// in-range circle covers the most distinct spawn paths, along with that
// coverage count. Scans in increasing (x,y) order and only replaces the
// current best on strictly greater coverage, so ties resolve deterministically
// to the lowest x, then lowest y. Returns x=-1 if no placeable tile exists.
func bestTowerSpot(g *game.Game, towerRange float64, paths [][]flow.Tile) (bestX, bestY, bestCoverage int) {
	bestX, bestY = -1, -1
	for x := 0; x < g.Grid.Width; x++ {
		for y := 0; y < g.Grid.Height; y++ {
			if !g.CanPlaceTower(x, y) {
				continue
			}
			coverage := coverageAt(x, y, towerRange, paths)
			if bestX < 0 || coverage > bestCoverage {
				bestX, bestY, bestCoverage = x, y, coverage
			}
		}
	}
	return bestX, bestY, bestCoverage
}

// coverageAt counts how many distinct spawn→base paths pass within
// towerRange of tile (x,y) — at most once per path, regardless of how many
// of that path's tiles are in range.
func coverageAt(x, y int, towerRange float64, paths [][]flow.Tile) int {
	count := 0
	rangeSq := towerRange * towerRange
	for _, path := range paths {
		for _, tile := range path {
			dx := float64(tile.X - x)
			dy := float64(tile.Y - y)
			if dx*dx+dy*dy <= rangeSq {
				count++
				break
			}
		}
	}
	return count
}
