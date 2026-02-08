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
			{0, 10},
			{10, 10},
			{10, 5},
			{25, 5},
			{25, 15},
			{39, 15},
		},
	}
}

func ApplyPath(grid *Grid, path Path) {
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

	start := path.Points[0]
	end := path.Points[len(path.Points)-1]

	grid.Tiles[start.Y][start.X] = SpawnTile
	grid.Tiles[end.Y][end.X] = BaseTile
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
