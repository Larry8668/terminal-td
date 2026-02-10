package render

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"terminal-td/internal/game"
)

type MenuOption int

const (
	MenuStart MenuOption = iota
	MenuControls
	MenuSettings
	MenuUpdateAvailable
	MenuQuit
)

func DrawMainMenu(screen tcell.Screen, selectedOption MenuOption, updateAvailable bool, latestVersion string) {
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

	centerX := w / 2
	row := h/2 - 2

	// Start
	startText := "START GAME"
	style := whiteStyle
	if selectedOption == MenuStart {
		style = yellowStyle
		drawText(screen, centerX-len(startText)/2-2, row, style, "> "+startText)
	} else {
		drawText(screen, centerX-len(startText)/2, row, style, startText)
	}
	row += 2

	// Controls
	controlsText := "CONTROLS"
	style = whiteStyle
	if selectedOption == MenuControls {
		style = yellowStyle
		drawText(screen, centerX-len(controlsText)/2-2, row, style, "> "+controlsText)
	} else {
		drawText(screen, centerX-len(controlsText)/2, row, style, controlsText)
	}
	row += 2

	// Settings
	settingsText := "SETTINGS"
	style = whiteStyle
	if selectedOption == MenuSettings {
		style = yellowStyle
		drawText(screen, centerX-len(settingsText)/2-2, row, style, "> "+settingsText)
	} else {
		drawText(screen, centerX-len(settingsText)/2, row, style, settingsText)
	}
	row += 2

	// Update available (only if update available)
	if updateAvailable {
		updateText := fmt.Sprintf("UPDATE AVAILABLE (%s)", latestVersion)
		style = whiteStyle
		if selectedOption == MenuUpdateAvailable {
			style = yellowStyle
			drawText(screen, centerX-len(updateText)/2-2, row, style, "> "+updateText)
		} else {
			drawText(screen, centerX-len(updateText)/2, row, style, updateText)
		}
		row += 2
	}

	// Quit
	quitText := "QUIT"
	style = whiteStyle
	if selectedOption == MenuQuit {
		style = yellowStyle
		drawText(screen, centerX-len(quitText)/2-2, row, style, "> "+quitText)
	} else {
		drawText(screen, centerX-len(quitText)/2, row, style, quitText)
	}

	// Instructions
	instructions := "Use ARROW KEYS or W/S to navigate, SPACE to select"
	instX := (w - len(instructions)) / 2
	drawText(screen, instX, h/2+7, cyanStyle, instructions)
}

func MaxMenuOption(updateAvailable bool) MenuOption {
	if updateAvailable {
		return MenuQuit
	}
	return MenuUpdateAvailable // Quit is at index 3 when update row is hidden
}

func DrawSettings(screen tcell.Screen, checkForUpdates bool) {
	w, h := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	yellowStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))

	title := "SETTINGS"
	titleX := (w - len(title)) / 2
	drawText(screen, titleX, h/2-6, greenStyle, title)

	label := "Check for updates (notified when update is available):"
	drawText(screen, w/2-len(label)/2, h/2-3, whiteStyle, label)
	value := "OFF"
	if checkForUpdates {
		value = "ON"
	}
	drawText(screen, w/2-len(value)/2, h/2-1, yellowStyle, value)

	helpText := "Press ESC to return to menu"
	drawText(screen, (w-len(helpText))/2, h/2+2, cyanStyle, helpText)
}

func DrawChangelog(screen tcell.Screen, content string) {
	w, h := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))

	title := "CHANGELOG"
	titleX := (w - len(title)) / 2
	drawText(screen, titleX, 2, greenStyle, title)

	// Word-wrap and draw content (simple: split by newline, draw lines)
	lines := splitLines(content, w-4)
	y := 4
	for _, line := range lines {
		if y >= h-3 {
			break
		}
		drawText(screen, 2, y, whiteStyle, line)
		y++
	}

	drawText(screen, w/2-20, h-2, cyanStyle, "Press any key to continue")
}

func splitLines(text string, maxWidth int) []string {
	var result []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		for len(line) > maxWidth {
			result = append(result, line[:maxWidth])
			line = line[maxWidth:]
		}
		if line != "" {
			result = append(result, line)
		}
	}
	return result
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
