package main

import (
	"github.com/gdamore/tcell/v2"
	"log"
	"os"
	"terminal-td/internal/entities"
	"terminal-td/internal/game"
	"terminal-td/internal/render"
	"time"
)

const tickRate = 100 * time.Millisecond

func main() {
	f, err := os.Create("debug.log")
	if err != nil {
		log.Printf("ERROR: Failed to create debug.log: %v", err)
	} else {
		log.SetOutput(f)
		log.Println("=== Terminal Tower Defense v0.01 ===")
		log.Println("Debug logging initialized")
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("ERROR: Failed to create screen: %v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("ERROR: Failed to initialize screen: %v", err)
	}
	defer screen.Fini()
	log.Println("Screen initialized successfully")

	g := game.NewGame()
	log.Println("Game instance created")

	events := make(chan tcell.Event, 10)
	quit := make(chan struct{})

	go screen.ChannelEvents(events, quit)

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	running := true
	menuSelection := render.MenuStart
	showControls := false

	log.Println("Entering main game loop")

	for running {
		select {

		case <-ticker.C:
			dt := tickRate.Seconds()

			screen.Clear()

			// Handle different game states
			switch g.Manager.State {
			case game.StateMenu:
				if showControls {
					render.DrawControls(screen)
				} else {
					render.DrawMainMenu(screen, menuSelection)
				}

			case game.StateQuitConfirm:
				render.DrawQuitConfirm(screen)

			default:
				// Normal game rendering
				g.Manager.Update(dt)

				if g.Manager.IsSimulationRunning() {
					g.Update(dt)
				}

				w, h := screen.Size()

				const uiHeight = 4
				const bottomHUDHeight = 5
				offsetX := (w - g.Grid.Width) / 2

				if offsetX < 0 {
					offsetX = 0
				}

				offsetY := uiHeight + (h-uiHeight-bottomHUDHeight-g.Grid.Height)/2

				if offsetY < 0 {
					offsetY = uiHeight
				}

				render.DrawGrid(screen, g.Grid, offsetX, offsetY)

				if g.Manager.Mode == game.ModeBuild {
					templates := game.GetTowerTemplates()
					template := templates[entities.TowerBasic]
					render.DrawRange(screen, g.CursorX, g.CursorY, template.Range, offsetX, offsetY)
				} else if g.Manager.Mode == game.ModeSelect {
					tower := g.GetTowerAt(g.Manager.SelectedTowerX, g.Manager.SelectedTowerY)
					if tower != nil {
						render.DrawRange(screen, tower.X, tower.Y, tower.Range, offsetX, offsetY)

						if tower.Target != nil && tower.Target.HP > 0 {
							render.DrawAttackLine(screen, tower.X, tower.Y, tower.Target.X, tower.Target.Y, offsetX, offsetY)
						}
					}
				}

				render.DrawTower(screen, g.Towers, offsetX, offsetY)
				render.DrawEnemies(screen, g.Enemies, offsetX, offsetY)
				render.DrawProjectiles(screen, g.Projectiles, offsetX, offsetY)
				render.DrawUI(screen, g)
				render.DrawCursor(screen, g.CursorX, g.CursorY, offsetX, offsetY)
				render.DrawBottomHUD(screen, g)
			}

			screen.Show()

		case ev := <-events:
			switch e := ev.(type) {

			case *tcell.EventKey:
				switch e.Key() {

				case tcell.KeyEscape:
					log.Println("DEBUG: ESC key pressed")
					if g.Manager.State == game.StateMenu {
						if showControls {
							log.Println("DEBUG: Exiting controls screen")
							showControls = false
						} else {
							log.Println("DEBUG: Quitting from menu")
							running = false
							close(quit)
						}
					} else if g.Manager.State == game.StateQuitConfirm {
						log.Println("DEBUG: Cancel quit confirmation")
						g.Manager.State = game.StateInWave
					} else if g.Manager.Mode != game.ModeNormal {
						log.Printf("DEBUG: Exiting mode %d", g.Manager.Mode)
						g.Manager.Mode = game.ModeNormal
					} else {
						log.Println("DEBUG: Showing quit confirmation")
						g.Manager.State = game.StateQuitConfirm
					}

				case tcell.KeyUp:
					if g.Manager.State == game.StateMenu && !showControls {
						if menuSelection > render.MenuStart {
							menuSelection--
							log.Printf("DEBUG: Menu selection: %d", menuSelection)
						}
					} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorY--
						clampCursor(g)
					}

				case tcell.KeyDown:
					if g.Manager.State == game.StateMenu && !showControls {
						if menuSelection < render.MenuQuit {
							menuSelection++
							log.Printf("DEBUG: Menu selection: %d", menuSelection)
						}
					} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorY++
						clampCursor(g)
					}

				case tcell.KeyLeft:
					if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorX--
						clampCursor(g)
					}

				case tcell.KeyRight:
					if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorX++
						clampCursor(g)
					}

				case tcell.KeyRune:
					switch e.Rune() {
					case 'w', 'W':
						if g.Manager.State == game.StateMenu && !showControls {
							if menuSelection > render.MenuStart {
								menuSelection--
							}
						} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.CursorY--
							clampCursor(g)
						}

					case 's', 'S':
						if g.Manager.State == game.StateMenu && !showControls {
							if menuSelection < render.MenuQuit {
								menuSelection++
							}
						} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.CursorY++
							clampCursor(g)
						}

					case 'a', 'A':
						if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.CursorX--
							clampCursor(g)
						}

					case 'd', 'D':
						if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.CursorX++
							clampCursor(g)
						}

					case '=', '+':
						if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							oldSpeed := g.Speed
							g.Speed = min(4.0, g.Speed*2)
							if oldSpeed != g.Speed {
								log.Printf("DEBUG: Speed increased to %.2fx", g.Speed)
							}
						}

					case '-':
						if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							oldSpeed := g.Speed
							g.Speed = max(0.25, g.Speed/2)
							if oldSpeed != g.Speed {
								log.Printf("DEBUG: Speed decreased to %.2fx", g.Speed)
							}
						}

					case 'p', 'P':
						if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.Manager.TogglePause()
						}

					case 'r', 'R':
						if g.Manager.State == game.StateWon || g.Manager.State == game.StateLost {
							log.Println("DEBUG: Restarting game")
							g.Reset()
						}

					case 'b', 'B':
						if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							if g.Manager.Mode == game.ModeNormal {
								log.Println("DEBUG: Entering build mode")
								g.Manager.Mode = game.ModeBuild
							} else {
								log.Println("DEBUG: Exiting build mode")
								g.Manager.Mode = game.ModeNormal
							}
						}

					case 'y', 'Y':
						if g.Manager.State == game.StateQuitConfirm {
							log.Println("DEBUG: User confirmed quit")
							running = false
							close(quit)
						}

					case 'n', 'N':
						if g.Manager.State == game.StateQuitConfirm {
							log.Println("DEBUG: User cancelled quit")
							g.Manager.State = game.StateInWave
						}

					case ' ', '\n', '\r':
						if g.Manager.State == game.StateMenu {
							if showControls {
								showControls = false
							} else {
								switch menuSelection {
								case render.MenuStart:
									log.Println("DEBUG: Starting game from menu")
									g.Manager.State = game.StatePreWave
									g.Manager.InterWaveTimer = 5.0
								case render.MenuControls:
									log.Println("DEBUG: Showing controls")
									showControls = true
								case render.MenuQuit:
									log.Println("DEBUG: Quitting from menu")
									running = false
									close(quit)
								}
							}
						} else if g.Manager.State == game.StateQuitConfirm {
							// Do nothing, use Y/N keys
						} else if g.Manager.Mode == game.ModeBuild {
							if g.PlaceTower(entities.TowerBasic) {
								log.Printf("DEBUG: Tower placed at (%d, %d)", g.CursorX, g.CursorY)
								g.Manager.Mode = game.ModeNormal
							} else {
								log.Printf("DEBUG: Failed to place tower at (%d, %d) - invalid location or insufficient funds", g.CursorX, g.CursorY)
							}
						} else if g.Manager.Mode == game.ModeNormal {
							tower := g.GetTowerAt(g.CursorX, g.CursorY)
							if tower != nil {
								log.Printf("DEBUG: Tower selected at (%d, %d)", g.CursorX, g.CursorY)
								g.Manager.Mode = game.ModeSelect
								g.Manager.SelectedTowerX = g.CursorX
								g.Manager.SelectedTowerY = g.CursorY
							} else {
								log.Println("DEBUG: Entering build mode (empty tile)")
								g.Manager.Mode = game.ModeBuild
							}
						} else if g.Manager.Mode == game.ModeSelect {
							log.Println("DEBUG: Deselecting tower")
							g.Manager.Mode = game.ModeNormal
						}
					}
				}
			}
		}
	}

	log.Println("=== Game session ended ===")
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

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
