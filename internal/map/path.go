package mapdata

type Point struct {
	X int
	Y int
}

type Path struct {
	Points []Point
}

func DefaultPath() Path {
	return Path{
		Points: []Point{
			{0, 12},
			{20, 12},
			{20, 5},
			{50, 5},
			{50, 22},
			{79, 22},
		},
	}
}

func ApplyPath(grid *Grid, path Path) {
	applyPathToGrid(grid, path, -1, -1)
}

// ApplyPathSegmentsOnly marks tiles along path segments as PathTile only (no spawn/base).
// Used by the loader when spawn and base come from map def, so fork start points stay path.
func ApplyPathSegmentsOnly(grid *Grid, path Path) {
	for i := 0; i < len(path.Points)-1; i++ {
		a := path.Points[i]
		b := path.Points[i+1]
		if a.X == b.X {
			start, end := min(a.Y, b.Y), max(a.Y, b.Y)
			for y := start; y <= end; y++ {
				grid.Tiles[y][a.X] = PathTile
			}
		} else if a.Y == b.Y {
			start, end := min(a.X, b.X), max(a.X, b.X)
			for x := start; x <= end; x++ {
				grid.Tiles[a.Y][x] = PathTile
			}
		}
	}
}

// applyPathToGrid draws path segments and spawn; if baseX, baseY >= 0 uses them for base tile, else path end.
func applyPathToGrid(grid *Grid, path Path, baseX, baseY int) {
	ApplyPathSegmentsOnly(grid, path)
	start := path.Points[0]
	grid.Tiles[start.Y][start.X] = SpawnTile
	if baseX >= 0 && baseY >= 0 {
		grid.Tiles[baseY][baseX] = BaseTile
	} else {
		end := path.Points[len(path.Points)-1]
		grid.Tiles[end.Y][end.X] = BaseTile
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
