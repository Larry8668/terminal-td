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
	MenuChangelog
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

	// Changelog
	changelogText := "CHANGELOG"
	style = whiteStyle
	if selectedOption == MenuChangelog {
		style = yellowStyle
		drawText(screen, centerX-len(changelogText)/2-2, row, style, "> "+changelogText)
	} else {
		drawText(screen, centerX-len(changelogText)/2, row, style, changelogText)
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

	row += 2

	// Instructions
	instructions := "Use ARROW KEYS or W/S to navigate, SPACE to select"
	instX := (w - len(instructions)) / 2
	drawText(screen, instX, row, cyanStyle, instructions)
}

func MaxMenuOption(updateAvailable bool) MenuOption {
	if updateAvailable {
		return MenuQuit
	}
	return MenuChangelog + 1 // Quit is at index 4 when update row is hidden
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

func DrawUpdateScreen(screen tcell.Screen, step string, percent int, done bool, err error) {
	w, h := screen.Size()
	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	greenStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))
	yellowStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	redStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	title := "UPDATING"
	drawText(screen, (w-len(title))/2, 2, greenStyle, title)
	drawText(screen, (w-len(step))/2, 5, whiteStyle, step)
	barWidth := 40
	if barWidth > w-4 {
		barWidth = w - 4
	}
	filled := (barWidth * percent) / 100
	if filled > barWidth {
		filled = barWidth
	}
	bar := "[" + strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled) + "]"
	drawText(screen, (w-len(bar))/2, 7, cyanStyle, bar)
	pctStr := fmt.Sprintf("%d%%", percent)
	drawText(screen, (w-len(pctStr))/2, 8, yellowStyle, pctStr)

	if done {
		if err != nil {
			drawText(screen, (w-14)/2, 11, redStyle, "Update failed")
			lines := splitLines(err.Error(), w-4)
			y := 12
			for _, line := range lines {
				if y >= h-4 {
					break
				}
				drawText(screen, 2, y, whiteStyle, line)
				y++
			}
			drawText(screen, (w-32)/2, h-2, cyanStyle, "Press any key to return to menu")
		} else {
			drawText(screen, (w-52)/2, 12, greenStyle, "The exe has been updated.")
			drawText(screen, (w-46)/2, 13, whiteStyle, "Reopen the game to play the latest version.")
			drawText(screen, (w-38)/2, h-2, cyanStyle, "Press Space or Enter to quit")
		}
	}
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

func DrawQuitConfirm(screen tcell.Screen, selectedYes bool) {
	w, h := screen.Size()

	whiteStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	cyanStyle := tcell.StyleDefault.Foreground(tcell.Color(6))
	yellowStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)

	message := "Are you sure you want to quit?"
	yesLabel := "YES"
	noLabel := "NO"
	hint := "Use ARROW KEYS or W/S to navigate, SPACE to select"

	msgX := (w - len(message)) / 2
	row := h/2 - 2
	drawText(screen, msgX, row, whiteStyle, message)

	optX := w/2 - 2
	if selectedYes {
		drawText(screen, optX, row+2, yellowStyle, "> "+yesLabel)
		drawText(screen, optX, row+3, whiteStyle, "  "+noLabel)
	} else {
		drawText(screen, optX, row+2, whiteStyle, "  "+yesLabel)
		drawText(screen, optX, row+3, yellowStyle, "> "+noLabel)
	}
	hintX := (w - len(hint)) / 2
	drawText(screen, hintX, row+5, cyanStyle, hint)
}
