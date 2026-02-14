package flow

import mapdata "terminal-td/internal/map"

// IsWalkable returns true only for path, spawn, and base tiles.
func IsWalkable(tile mapdata.TileType) bool {
	return tile == mapdata.PathTile || tile == mapdata.SpawnTile || tile == mapdata.BaseTile
}

// BuildWalkability returns a mask [y][x] where true = path/spawn/base.
func BuildWalkability(grid *mapdata.Grid) [][]bool {
	w := make([][]bool, grid.Height)
	for y := 0; y < grid.Height; y++ {
		w[y] = make([]bool, grid.Width)
		for x := 0; x < grid.Width; x++ {
			w[y][x] = IsWalkable(grid.Tiles[y][x])
		}
	}
	return w
}
