package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"time"
)

const (
	screenWidth  = 40
	screenHeight = 20
	tickRate     = 200 * time.Millisecond
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()

	screen.Clear()

	running := true
	cursorX, cursorY := 5, 5

	for running {
		start := time.Now()

		event := screen.PollEvent()
		switch ev := event.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				running = false

			case tcell.KeyUp:
				cursorY--

			case tcell.KeyDown:
				cursorY++

			case tcell.KeyLeft:
				cursorX--

			case tcell.KeyRight:
				cursorX++
			}
		}

		if cursorX < 0 {
			cursorX = 0
		}
		if cursorY < 0 {
			cursorY = 0
		}
		if cursorX >= screenWidth {
			cursorX = screenWidth - 1
		}
		if cursorY >= screenHeight {
			cursorY = screenHeight - 1
		}

		screen.Clear()
		drawGrid(screen)
		drawCursor(screen, cursorX, cursorY)
		screen.Show()

		elapsed := time.Since(start)
		sleep := tickRate - elapsed

		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
}

func drawGrid(screen tcell.Screen) {
	style := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)

	for y := 0; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			screen.SetContent(x, y, '.', nil, style)
		}
	}
}

func drawCursor(screen tcell.Screen, x, y int) {
	style := tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorGreen)

	screen.SetContent(x, y, 'X', nil, style)
}
