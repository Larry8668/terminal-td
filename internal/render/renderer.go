package render

import (
	"github.com/gdamore/tcell/v2"
	"terminal-td/internal/entities"
	mapdata "terminal-td/internal/map"
)

func DrawGrid(screen tcell.Screen, grid *mapdata.Grid) {
	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			var ch rune

			switch grid.Tiles[y][x] {
			case mapdata.Empty:
				ch = '.'
			case mapdata.PathTile:
				ch = '='
			case mapdata.SpawnTile:
				ch = 'S'
			case mapdata.BaseTile:
				ch = 'E'
			}

			screen.SetContent(x, y, ch, nil, tcell.StyleDefault)
		}
	}
}

func DrawEnemies(screen tcell.Screen, enemies []*entities.Enemy) {
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)

	for _, e := range enemies {
		x := int(e.X)
		y := int(e.Y)

		screen.SetContent(x, y, 'M', nil, style)
	}
}
