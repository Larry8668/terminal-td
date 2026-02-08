package render

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"terminal-td/internal/entities"
	"terminal-td/internal/game"
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

func DrawUI(screen tcell.Screen, g *game.Game) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	waveText := fmt.Sprintf(
		"Wave: %d/%d",
		g.Wave.CurrentWave,
		g.Wave.TotalWaves,
	)
	enemyText := fmt.Sprintf(
		"Enemies: %d",
		g.Wave.EnemiesAlive,
	)
	hpText := fmt.Sprintf(
		"Base HP: %d",
		g.Base.HP,
	)

	drawText(screen, 0, 0, style, waveText)
	drawText(screen, 0, 1, style, enemyText)
	drawText(screen, 0, 2, style, hpText)
}

func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range text {
		screen.SetContent(x+i, y, r, nil, style)
	}
}
