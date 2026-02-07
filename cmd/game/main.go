package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"terminal-td/internal/game"
	"terminal-td/internal/render"
	"time"
)

const tickRate = 100 * time.Millisecond

func main() {
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
			g.Update(tickRate.Seconds())

			screen.Clear()
			render.DrawGrid(screen, g.Grid)
			render.DrawEnemies(screen, g.Enemies)
			screen.Show()

		case ev := <-events:
			switch e := ev.(type) {

			case *tcell.EventKey:
				switch e.Key() {

				case tcell.KeyEscape:
					running = false
					close(quit)
				}
			}
		}
	}
}
