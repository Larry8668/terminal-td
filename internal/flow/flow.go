package flow

import "math"

const Inf = 1e9

type Field struct {
	Width      int
	Height     int
	Distances  [][]float64
	Directions [][]Vec2
}

func NewField(width, height int) *Field {
	dist := make([][]float64, height)
	dirs := make([][]Vec2, height)

	for y := 0; y < height; y++ {
		dist[y] = make([]float64, width)
		dirs[y] = make([]Vec2, width)
		for x := 0; x < width; x++ {
			dist[y][x] = Inf
		}
	}

	return &Field{
		Width:      width,
		Height:     height,
		Distances:  dist,
		Directions: dirs,
	}
}

func (f *Field) At(x, y int) (dist float64, dir Vec2) {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return Inf, Vec2{}
	}
	return f.Distances[y][x], f.Directions[y][x]
}

func (f *Field) AtFloat(x, y float64) (dist float64, dir Vec2) {
	tx := int(x)
	ty := int(y)

	if tx < 0 || tx >= f.Width || ty < 0 || ty >= f.Height {
		return Inf, Vec2{}
	}
	return f.Distances[ty][tx], f.Directions[ty][tx]
}

// Tile is a grid cell (for path trace).
type Tile struct {
	X, Y int
}

const maxTraceSteps = 500

// TracePath follows flow directions from (startX, startY) toward base until reaching (baseX, baseY) or max steps.
// Returns a slice of tiles from spawn toward base (excluding spawn, including base).
func (f *Field) TracePath(startX, startY, baseX, baseY int) []Tile {
	var path []Tile
	x, y := startX, startY
	for i := 0; i < maxTraceSteps; i++ {
		dist, dir := f.At(x, y)
		if dist >= Inf || (dir.X == 0 && dir.Y == 0) {
			break
		}
		nx := x + int(math.Round(dir.X))
		ny := y + int(math.Round(dir.Y))
		if nx < 0 || nx >= f.Width || ny < 0 || ny >= f.Height {
			break
		}
		path = append(path, Tile{X: nx, Y: ny})
		x, y = nx, ny
		if x == baseX && y == baseY {
			break
		}
	}
	return path
}
