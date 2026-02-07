package mapdata

type TileType int

const (
	Empty TileType = iota
	PathTile
	SpawnTile
	BaseTile
)

type Grid struct {
	Width  int
	Height int
	Tiles  [][]TileType
}

func NewGrid(w, h int) *Grid {
	tiles := make([][]TileType, h)

	for y := range tiles {
		tiles[y] = make([]TileType, w)
	}

	return &Grid{
		Width:  w,
		Height: h,
		Tiles:  tiles,
	}
}
