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
			render.DrawGrid(screen, g.Grid)
			render.DrawEnemies(screen, g.Enemies)
			render.DrawUI(screen, g)
			screen.Show()

		case ev := <-events:
			switch e := ev.(type) {

			case *tcell.EventKey:
				switch e.Key() {

				case tcell.KeyEscape:
					running = false
					close(quit)

				case tcell.KeyRune:
					switch e.Rune() {
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
