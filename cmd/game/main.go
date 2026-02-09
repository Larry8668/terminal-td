package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
	"terminal-td/internal/game"
	"terminal-td/internal/render"
	"time"
)

const tickRate = 100 * time.Millisecond

func main() {
	f, _ := os.Create("debug.log")
	log.SetOutput(f)

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()

	g := game.NewGame()

	events := make(chan tcell.Event, 10)
	quit := make(chan struct{})

	go screen.ChannelEvents(events, quit)

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	running := true

	for running {
		select {

		case <-ticker.C:
			dt := tickRate.Seconds()

			g.Manager.Update(dt)

			if g.Manager.IsSimulationRunning() {
				g.Update(dt)
			}
			screen.Clear()

			w, h := screen.Size()
			const uiHeight = 4
			offsetX := (w - g.Grid.Width) / 2
			if offsetX < 0 {
				offsetX = 0
			}
			offsetY := uiHeight + (h-uiHeight-g.Grid.Height)/2
			if offsetY < 0 {
				offsetY = uiHeight
			}

			render.DrawGrid(screen, g.Grid, offsetX, offsetY)
			render.DrawEnemies(screen, g.Enemies, offsetX, offsetY)
			render.DrawUI(screen, g)
			render.DrawCursor(screen, g.CursorX, g.CursorY, offsetX, offsetY)
			screen.Show()

		case ev := <-events:
			switch e := ev.(type) {

			case *tcell.EventKey:
				switch e.Key() {

				case tcell.KeyEscape:
					running = false
					close(quit)

				case tcell.KeyUp:
					g.CursorY--
					clampCursor(g)

				case tcell.KeyDown:
					g.CursorY++
					clampCursor(g)

				case tcell.KeyLeft:
					g.CursorX--
					clampCursor(g)

				case tcell.KeyRight:
					g.CursorX++
					clampCursor(g)

				case tcell.KeyRune:
					switch e.Rune() {
					case 'w', 'W':
						g.CursorY--
						clampCursor(g)

					case 's', 'S':
						g.CursorY++
						clampCursor(g)

					case 'a', 'A':
						g.CursorX--
						clampCursor(g)

					case 'd', 'D':
						g.CursorX++
						clampCursor(g)

					case '=': // if + then need to hit shift+= :/
						g.Speed = min(4.0, g.Speed*2)

					case '-':
						g.Speed = max(0.25, g.Speed/2)

					case 'p':
						g.Manager.TogglePause()

					case 'r':
						if g.Manager.State == game.StateWon || g.Manager.State == game.StateLost {
							g.Reset()
						}
					}
				}
			}
		}
	}
}

func clampCursor(g *game.Game) {
	if g.CursorX < 0 {
		g.CursorX = 0
	}
	if g.CursorX >= g.Grid.Width {
		g.CursorX = g.Grid.Width - 1
	}

	if g.CursorY < 0 {
		g.CursorY = 0
	}
	if g.CursorY >= g.Grid.Height {
		g.CursorY = g.Grid.Height - 1
	}
}
