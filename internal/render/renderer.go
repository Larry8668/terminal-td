package render

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"terminal-td/internal/entities"
	"terminal-td/internal/game"
	mapdata "terminal-td/internal/map"
)

func DrawGrid(screen tcell.Screen, grid *mapdata.Grid, offsetX, offsetY int) {
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

			screen.SetContent(offsetX+x, offsetY+y, ch, nil, tcell.StyleDefault)
		}
	}
}

func DrawEnemies(screen tcell.Screen, enemies []*entities.Enemy, offsetX, offsetY int) {
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)

	for _, e := range enemies {
		x := offsetX + int(e.X)
		y := offsetY + int(e.Y)

		screen.SetContent(x, y, 'M', nil, style)
	}
}

func DrawUI(screen tcell.Screen, g *game.Game) {
	w, _ := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	waveText := fmt.Sprintf("Wave: %d/%d", g.Wave.CurrentWave, g.Wave.TotalWaves)
	enemyText := fmt.Sprintf("Enemies: %d", g.Wave.EnemiesAlive)
	hpText := fmt.Sprintf("Base HP: %d", g.Base.HP)

	drawText(screen, 0, 0, whiteStyle, waveText)
	drawText(screen, 0, 1, whiteStyle, enemyText)
	drawText(screen, 0, 2, whiteStyle, hpText)

	rightEdgeX := w - 2

	speedText := fmt.Sprintf("Speed: %.2fx", g.Speed)
	scoreText := fmt.Sprintf("Score: %d", g.Score.Points)
	runTimeText := fmt.Sprintf("Run Time: %s", FormatTime(g.Manager.RunTime))

	drawTextRight(screen, rightEdgeX, 0, whiteStyle, speedText)
	drawTextRight(screen, rightEdgeX, 1, whiteStyle, scoreText)
	drawTextRight(screen, rightEdgeX, 2, whiteStyle, runTimeText)

	rightRow := 3

	if g.Manager.State == game.StatePreWave {
		nextWaveTimeText := fmt.Sprintf("Next Wave In: %s", FormatTime(g.Manager.InterWaveTimer))
		drawTextRight(screen, rightEdgeX, rightRow+1, whiteStyle, nextWaveTimeText)
	}

	var stateText string
	var stateStyle tcell.Style
	showState := false

	switch g.Manager.State {
	case game.StatePaused:
		stateText = "⏸ PAUSED"
		stateStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
		showState = true
	case game.StateWon:
		stateText = "✓ VICTORY"
		stateStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		showState = true
	case game.StateLost:
		stateText = "✗ DEFEAT"
		stateStyle = tcell.StyleDefault.Foreground(tcell.ColorRed)
		showState = true
	case game.StateInWave:
	case game.StatePreWave:
		showState = false
		return
	}
	if showState {
		drawTextRight(screen, rightEdgeX, rightRow+2, stateStyle, stateText)
	}
}

func DrawCursor(screen tcell.Screen, cursorX, cursorY, offsetX, offsetY int) {
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	screen.SetContent(offsetX+cursorX, offsetY+cursorY, '+', nil, style)
}

func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range text {
		screen.SetContent(x+i, y, r, nil, style)
	}
}

func drawTextRight(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	startX := x - len(text)
	drawText(screen, startX, y, style, text)
}

func FormatTime(t float64) string {
	min := int(t) / 60
	sec := int(t) % 60
	return fmt.Sprintf("%02d:%02d", min, sec)
}
