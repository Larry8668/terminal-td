package render

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"terminal-td/internal/game"
)

type MenuOption int

const (
	MenuStart MenuOption = iota
	MenuControls
	MenuQuit
)

func DrawMainMenu(screen tcell.Screen, selectedOption MenuOption) {
	w, h := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))
	yellowStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)

	// Title
	title := "TERMINAL TOWER DEFENSE"
	titleX := (w - len(title)) / 2
	drawText(screen, titleX, h/2-8, greenStyle, title)

	// Version
	version := fmt.Sprintf("Version %s", game.Version)
	versionX := (w - len(version)) / 2
	drawText(screen, versionX, h/2-6, cyanStyle, version)

	// Menu options
	startText := "START GAME"
	controlsText := "CONTROLS"
	quitText := "QUIT"

	centerX := w / 2

	// Start option
	style := whiteStyle
	if selectedOption == MenuStart {
		style = yellowStyle
		drawText(screen, centerX-len(startText)/2-2, h/2-2, style, "> "+startText)
	} else {
		drawText(screen, centerX-len(startText)/2, h/2-2, style, startText)
	}

	// Controls option
	style = whiteStyle
	if selectedOption == MenuControls {
		style = yellowStyle
		drawText(screen, centerX-len(controlsText)/2-2, h/2, style, "> "+controlsText)
	} else {
		drawText(screen, centerX-len(controlsText)/2, h/2, style, controlsText)
	}

	// Quit option
	style = whiteStyle
	if selectedOption == MenuQuit {
		style = yellowStyle
		drawText(screen, centerX-len(quitText)/2-2, h/2+2, style, "> "+quitText)
	} else {
		drawText(screen, centerX-len(quitText)/2, h/2+2, style, quitText)
	}

	// Instructions
	instructions := "Use ARROW KEYS or W/S to navigate, SPACE to select"
	instX := (w - len(instructions)) / 2
	drawText(screen, instX, h/2+5, cyanStyle, instructions)
}

func DrawControls(screen tcell.Screen) {
	w, h := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	yellowStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)

	// Title
	title := "CONTROLS"
	titleX := (w - len(title)) / 2
	drawText(screen, titleX, h/2-10, greenStyle, title)

	controls := []string{
		"MOVEMENT:",
		"  Arrow Keys or WASD - Move cursor",
		"",
		"BUILDING:",
		"  B - Toggle build mode",
		"  SPACE/ENTER - Place tower / Select tower",
		"  ESC - Cancel build mode / Deselect",
		"",
		"GAMEPLAY:",
		"  P - Pause / Unpause",
		"  +/- - Increase / Decrease game speed",
		"  R - Restart (when game over)",
		"",
		"QUIT:",
		"  ESC - Quit game",
		"",
		"Press ESC to return to menu",
	}

	y := h/2 - 8
	for i, line := range controls {
		if i == 0 || i == 3 || i == 8 || i == 13 {
			// Section headers
			drawText(screen, w/2-len(line)/2, y, yellowStyle, line)
		} else if line == "" {
			// Empty line
		} else {
			drawText(screen, w/2-len(line)/2, y, whiteStyle, line)
		}
		y++
	}
}

func DrawQuitConfirm(screen tcell.Screen) {
	w, h := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	redStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))

	// Dialog box
	message := "Are you sure you want to quit?"
	yesText := "YES (Y)"
	noText := "NO (N or ESC)"

	boxWidth := 50
	boxHeight := 7
	boxX := (w - boxWidth) / 2
	boxY := (h - boxHeight) / 2

	// Draw box border
	for x := boxX; x < boxX+boxWidth; x++ {
		screen.SetContent(x, boxY, '-', nil, whiteStyle)
		screen.SetContent(x, boxY+boxHeight-1, '-', nil, whiteStyle)
	}
	for y := boxY; y < boxY+boxHeight; y++ {
		screen.SetContent(boxX, y, '|', nil, whiteStyle)
		screen.SetContent(boxX+boxWidth-1, y, '|', nil, whiteStyle)
	}

	// Message
	msgX := boxX + (boxWidth-len(message))/2
	drawText(screen, msgX, boxY+2, whiteStyle, message)

	// Options
	yesX := boxX + (boxWidth-len(yesText))/2
	noX := boxX + (boxWidth-len(noText))/2
	drawText(screen, yesX, boxY+4, redStyle, yesText)
	drawText(screen, noX, boxY+5, cyanStyle, noText)
}
