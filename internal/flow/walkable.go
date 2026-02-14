package flow

import mapdata "terminal-td/internal/map"

// IsWalkable returns true only for path, spawn, and base tiles.
func IsWalkable(tile mapdata.TileType) bool {
	return tile == mapdata.PathTile || tile == mapdata.SpawnTile || tile == mapdata.BaseTile
}

// BuildWalkability returns a mask [y][x] where true = path/spawn/base. No blocking.
func BuildWalkability(grid *mapdata.Grid) [][]bool {
	return BuildWalkabilityWithBlocked(grid, nil)
}

// BuildWalkabilityWithBlocked returns walkable mask; path tiles in blocked set are not walkable.
func BuildWalkabilityWithBlocked(grid *mapdata.Grid, blockedTiles [][2]int) [][]bool {
	blocked := make(map[[2]int]bool)
	for _, p := range blockedTiles {
		blocked[p] = true
	}
	w := make([][]bool, grid.Height)
	for y := 0; y < grid.Height; y++ {
		w[y] = make([]bool, grid.Width)
		for x := 0; x < grid.Width; x++ {
			if blocked[[2]int{x, y}] {
				w[y][x] = false
				continue
			}
			w[y][x] = IsWalkable(grid.Tiles[y][x])
		}
	}
	return w
}
