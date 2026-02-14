package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"

	"terminal-td/internal/entities"
	"terminal-td/internal/flow"
	"terminal-td/internal/game"
	mapdata "terminal-td/internal/map"
)

func DrawGrid(screen tcell.Screen, grid *mapdata.Grid, offsetX, offsetY int) {
	DrawGridWithHighlights(screen, grid, nil, offsetX, offsetY, nil, 0)
}

func DrawGridWithHighlights(screen tcell.Screen, grid *mapdata.Grid, mapData *mapdata.GameMap, offsetX, offsetY int, highlightSpawns map[string]bool, blinkTimer float64) {
	defaultStyle := tcell.StyleDefault
	redStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	blinkRedStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	shouldBlink := int(blinkTimer*4)%2 == 0

	spawnPositions := make(map[int]map[int]string)
	if mapData != nil {
		for _, spawn := range mapData.Spawns {
			if spawnPositions[spawn.Y] == nil {
				spawnPositions[spawn.Y] = make(map[int]string)
			}
			spawnPositions[spawn.Y][spawn.X] = spawn.ID
		}
	}

	for y := 0; y < grid.Height; y++ {
		for x := 0; x < grid.Width; x++ {
			var ch rune
			style := defaultStyle

			switch grid.Tiles[y][x] {
			case mapdata.Empty:
				ch = '.'
			case mapdata.PathTile:
				ch = '='
			case mapdata.SpawnTile:
				ch = 'S'
				if highlightSpawns != nil && spawnPositions[y] != nil {
					if spawnID, ok := spawnPositions[y][x]; ok && highlightSpawns[spawnID] {
						if shouldBlink {
							style = blinkRedStyle
						} else {
							style = redStyle
						}
					}
				}
			case mapdata.BaseTile:
				ch = 'E'
			}

			screen.SetContent(offsetX+x, offsetY+y, ch, nil, style)
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

// DrawPathPreview draws traced paths from spawns to base (dim overlay). Call during pre-wave.
func DrawPathPreview(screen tcell.Screen, paths [][]flow.Tile, offsetX, offsetY int) {
	style := tcell.StyleDefault.Foreground(tcell.Color(6)).Dim(true)
	for _, path := range paths {
		for _, t := range path {
			screen.SetContent(offsetX+t.X, offsetY+t.Y, '·', nil, style)
		}
	}
}

// DrawBlockedOverlay draws blocked path tiles (wall segments) as '#'.
func DrawBlockedOverlay(screen tcell.Screen, blockedTiles [][2]int, offsetX, offsetY int) {
	style := tcell.StyleDefault.Foreground(tcell.Color(8)).Dim(true)
	for _, p := range blockedTiles {
		screen.SetContent(offsetX+p[0], offsetY+p[1], '#', nil, style)
	}
}

func DrawUI(screen tcell.Screen, g *game.Game) {
	w, _ := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	waveText := fmt.Sprintf("Wave: %d/%d", g.GetCurrentWave(), g.GetTotalWaves())
	enemyText := fmt.Sprintf("Enemies: %d", g.GetEnemiesAlive())
	hpText := fmt.Sprintf("Base HP: %d", g.Base.HP)

	drawText(screen, 0, 0, whiteStyle, waveText)
	drawText(screen, 0, 1, whiteStyle, enemyText)
	drawText(screen, 0, 2, whiteStyle, hpText)
	if flowDebug := g.FlowDebugString(); flowDebug != "" {
		cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))
		drawText(screen, 0, 3, cyanStyle, flowDebug)
	}

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

func DrawTower(screen tcell.Screen, towers []*entities.Tower, offsetX, offsetY int, linkableTowers, removeWallTowers [][2]int) {
	linkableSet := make(map[[2]int]bool)
	for _, p := range linkableTowers {
		linkableSet[p] = true
	}
	removeSet := make(map[[2]int]bool)
	for _, p := range removeWallTowers {
		removeSet[p] = true
	}
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	yellowStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	for _, tower := range towers {
		pos := [2]int{tower.X, tower.Y}
		switch {
		case removeSet[pos]:
			screen.SetContent(offsetX+tower.X, offsetY+tower.Y, tower.Symbol, nil, yellowStyle)
		case linkableSet[pos]:
			screen.SetContent(offsetX+tower.X, offsetY+tower.Y, tower.Symbol, nil, greenStyle)
		default:
			style := tcell.StyleDefault.Foreground(tcell.Color(tower.Color))
			screen.SetContent(offsetX+tower.X, offsetY+tower.Y, tower.Symbol, nil, style)
		}
	}
}

func DrawBottomHUD(screen tcell.Screen, g *game.Game) {
	w, h := screen.Size()

	hudStartY := h - 5

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	redStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)

	separator := strings.Repeat("_", w)
	drawText(screen, 0, hudStartY, whiteStyle, separator)

	switch g.Manager.Mode {
	case game.ModeBuild:
		templates := game.GetTowerTemplates()
		template := templates[entities.TowerBasic]

		canAfford := g.Money >= template.Cost
		costStyle := whiteStyle

		if !canAfford {
			costStyle = redStyle
		}

		buildText := fmt.Sprintf("Build: [%c] %s - Cost: %d", template.Symbol, template.Name, template.Cost)
		moneyText := fmt.Sprintf("Money: %d", g.Money)
		helpText := "Press SPACE/ENTER to build, ESC/B to cancel"

		drawText(screen, 0, hudStartY+1, costStyle, buildText)
		drawText(screen, 0, hudStartY+2, whiteStyle, moneyText)
		drawText(screen, 0, hudStartY+3, cyanStyle, helpText)

		if g.CanPlaceTower(g.CursorX, g.CursorY) {
			drawText(screen, 0, hudStartY+4, greenStyle, "✓ Valid placement")
		} else {
			drawText(screen, 0, hudStartY+4, redStyle, "✗ Invalid placement (on path or existing tower)")
		}

	case game.ModeSelect:
		tower := g.GetTowerAt(g.Manager.SelectedTowerX, g.Manager.SelectedTowerY)
		if tower != nil {
			templates := game.GetTowerTemplates()
			template := templates[tower.Type]
			dps := tower.Damage * tower.FireRate
			linkable := g.GetLinkableTowers(g.Manager.SelectedTowerX, g.Manager.SelectedTowerY)
			greyStyle := tcell.StyleDefault.Foreground(tcell.Color(8)).Dim(true)

			wallsForTower := g.GetWallsForTower(g.Manager.SelectedTowerX, g.Manager.SelectedTowerY)
			drawText(screen, 0, hudStartY+1, whiteStyle, fmt.Sprintf("Tower: [%c] %s", tower.Symbol, template.Name))
			drawText(screen, 0, hudStartY+2, whiteStyle, fmt.Sprintf("DPS: %.1f | Range: %.1f", dps, tower.Range))
			if g.Manager.SelectingWallTarget {
				drawText(screen, 0, hudStartY+3, cyanStyle, "Select a green tower to link (SPACE/ENTER), 0 cancel")
			} else if g.Manager.SelectingWallRemoveTarget {
				drawText(screen, 0, hudStartY+3, cyanStyle, "Select a yellow tower to remove wall (SPACE/ENTER), 0 cancel")
			} else {
				if len(linkable) == 0 {
					drawText(screen, 0, hudStartY+3, greyStyle, "1. Build wall (can't block last path to base)")
				} else {
					drawText(screen, 0, hudStartY+3, greenStyle, "1. Build wall")
				}
				if len(wallsForTower) == 0 {
					drawText(screen, 0, hudStartY+4, greyStyle, "2. Remove wall (none)  ")
					drawText(screen, 24, hudStartY+4, cyanStyle, "3. Sell tower  0. Deselect")
				} else {
					drawText(screen, 0, hudStartY+4, greenStyle, "2. Remove wall  ")
					drawText(screen, 16, hudStartY+4, cyanStyle, "3. Sell tower  0. Deselect")
				}
			}
		}
	case game.ModeNormal:
		moneyText := fmt.Sprintf("Money: %d", g.Money)
		helpText := "Press SPACE on empty tile to build, on tower to select"
		drawText(screen, 0, hudStartY+1, whiteStyle, moneyText)
		drawText(screen, 0, hudStartY+2, cyanStyle, helpText)
	}
}

func DrawRange(screen tcell.Screen, centerX, centerY int, rangeVal float64, offsetX, offsetY int) {
	rangeInt := int(rangeVal)

	for dy := -rangeInt; dy <= rangeInt; dy++ {
		for dx := -rangeInt; dx <= rangeInt; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist <= rangeVal+0.5 && dist >= rangeVal-0.5 {
				x := offsetX + centerX + dx
				y := offsetY + centerY + dy

				if x >= 0 && y >= 0 {
					style := tcell.StyleDefault.Foreground(tcell.Color(6)).Dim(true)
					screen.SetContent(x, y, '.', nil, style)
				}
			}
		}
	}
}

func DrawProjectiles(screen tcell.Screen, projectiles []*entities.Projectile, offsetX, offsetY int) {
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow)

	for _, proj := range projectiles {
		x := offsetX + int(proj.X)
		y := offsetY + int(proj.Y)

		screen.SetContent(x, y, '*', nil, style)
	}
}

func DrawAttackLine(screen tcell.Screen, fromX, fromY int, toX, toY float64, offsetX, offsetY int) {
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow).Dim(true)

	screenToX := offsetX + int(toX)
	screenToY := offsetY + int(toY)
	screenFromX := offsetX + fromX
	screenFromY := offsetY + fromY

	dx := screenToX - screenFromX
	dy := screenToY - screenFromY

	steps := max(abs(dx), abs(dy))
	if steps == 0 {
		return
	}

	for i := 0; i <= steps; i++ {
		if i%2 == 0 {
			x := screenFromX + (dx*i)/steps
			y := screenFromY + (dy*i)/steps
			screen.SetContent(x, y, '.', nil, style)
		}
	}
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
