package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"

	"terminal-td/internal/config"
	"terminal-td/internal/entities"
	"terminal-td/internal/game"
	mapdata "terminal-td/internal/map"
	"terminal-td/internal/render"
	"terminal-td/internal/updater"
)

const (
	tickRate             = 100 * time.Millisecond
	sessionLogPrefix     = "terminal-td-session-"
	sessionLogSuffix     = ".log"
	maxSessionLogs       = 5
	sessionLogTimeFormat = "20060102-150405"
)

func initSessionLog() (*os.File, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	cleanupSessionLogs(dir)
	name := sessionLogPrefix + time.Now().Format(sessionLogTimeFormat) + sessionLogSuffix
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func cleanupSessionLogs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, sessionLogPrefix) && strings.HasSuffix(n, sessionLogSuffix) {
			matches = append(matches, n)
		}
	}
	if len(matches) <= maxSessionLogs {
		return
	}
	sort.Strings(matches)
	for i := 0; i < len(matches)-maxSessionLogs; i++ {
		_ = os.Remove(filepath.Join(dir, matches[i]))
	}
}

func main() {
	justUpdated := flag.Bool("just-updated", false, "Show changelog after update")
	changelogPath := flag.String("changelog", "", "Path to changelog file")
	flag.Parse()

	f, err := initSessionLog()
	if err != nil {
		log.Printf("ERROR: Failed to create session log: %v", err)
	} else {
		defer f.Close()
		log.SetOutput(f)
		log.Printf("=== Terminal Tower Defense %s ===", game.Version)
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

	if *justUpdated && *changelogPath != "" {
		showChangelogScreen(screen, *changelogPath)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Printf("config load: %v, using defaults", err)
		cfg = config.Default()
	}

	var updateAvailable bool
	var latestVersion string
	var latestRelease *updater.Release
	if cfg.CheckForUpdates {
		release, err := updater.FetchLatest(updater.DefaultOwner, updater.DefaultRepo)
		if err != nil {
			log.Printf("check for update: %v", err)
		} else if updater.IsNewer(game.Version, release.TagName) {
			updateAvailable = true
			latestVersion = release.TagName
			latestRelease = release
			log.Printf("Update available: %s", latestVersion)
		}
	}

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
	showSettings := false
	showChangelog := false
	changelogContent := ""
	quitConfirmYes := false
	showUpdateScreen := false
	var updateProgress *updater.Progress
	var updateStarted bool
	showMapSelection := false
	var availableMaps []mapdata.MapInfo
	var mapSelectionIndex int
	if maps, err := mapdata.ListMaps(); err != nil {
		log.Printf("load maps: %v", err)
		availableMaps = []mapdata.MapInfo{{ID: "classic", Name: "Tutorial"}}
	} else {
		availableMaps = maps
	}

	handleMenuSelect := func() bool {
		if g.Manager.State != game.StateMenu {
			return false
		}
		if showSettings {
			cfg.CheckForUpdates = !cfg.CheckForUpdates
			if err := config.Save(cfg); err != nil {
				log.Printf("config save: %v", err)
			}
			return false
		}
		if showControls {
			showControls = false
			return false
		}
		if showChangelog {
			showChangelog = false
			return false
		}
		if showUpdateScreen && updateProgress != nil && updateProgress.Done {
			if updateProgress.Err != nil {
				showUpdateScreen = false
			}
			return false
		}
		if menuSelection == render.MenuUpdateAvailable && updateAvailable && latestRelease != nil {
			showUpdateScreen = true
			updateProgress = &updater.Progress{}
			updateStarted = false
			return false
		}
		if menuSelection == render.MenuUpdateAvailable && !updateAvailable || menuSelection == render.MenuQuit {
			log.Println("DEBUG: Quitting from menu")
			running = false
			close(quit)
			return true
		}
		switch menuSelection {
		case render.MenuStart:
			log.Println("DEBUG: Showing map selection")
			showMapSelection = true
			mapSelectionIndex = 0
		case render.MenuControls:
			log.Println("DEBUG: Showing controls")
			showControls = true
		case render.MenuSettings:
			log.Println("DEBUG: Showing settings")
			showSettings = true
		case render.MenuChangelog:
			release, err := updater.FetchLatest(updater.DefaultOwner, updater.DefaultRepo)
			if err != nil {
				changelogContent = "Failed to load changelog: " + err.Error()
			} else {
				changelogContent = strings.TrimSpace(release.Body)
				if changelogContent == "" {
					changelogContent = "No changelog for this release."
				}
			}
			showChangelog = true
		}
		return false
	}

	log.Println("Entering main game loop")

	for running {
		select {

		case <-ticker.C:
			dt := tickRate.Seconds()

			screen.Clear()

			switch g.Manager.State {
			case game.StateMenu:
				if showUpdateScreen && updateProgress != nil {
					if !updateStarted {
						updateStarted = true
						go updater.RunUpdateWithProgress(latestRelease, updateProgress)
					}
					render.DrawUpdateScreen(screen, updateProgress.Step, updateProgress.Percent, updateProgress.Done, updateProgress.Err)
				} else if showMapSelection {
					render.DrawMapSelection(screen, availableMaps, mapSelectionIndex)
				} else if showSettings {
					render.DrawSettings(screen, cfg.CheckForUpdates)
				} else if showControls {
					render.DrawControls(screen)
				} else if showChangelog {
					render.DrawChangelog(screen, changelogContent)
				} else {
					render.DrawMainMenu(screen, menuSelection, updateAvailable, latestVersion)
				}

			case game.StateQuitConfirm:
				render.DrawQuitConfirm(screen, quitConfirmYes)

			default:
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

				var highlightSpawns map[string]bool
				blinkTimer := g.Manager.RunTime
				if g.Manager.State == game.StatePreWave {
					highlightSpawns = g.GetNextWaveSpawnIDs()
				}
				render.DrawGridWithHighlights(screen, g.Grid, g.Map, offsetX, offsetY, highlightSpawns, blinkTimer)

				if g.Manager.State == game.StatePreWave && g.FlowField != nil {
					pathPreview := g.TracePathsForNextWave()
					render.DrawPathPreview(screen, pathPreview, offsetX, offsetY)
				}

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
						if showUpdateScreen && updateProgress != nil && updateProgress.Done {
							showUpdateScreen = false
						} else if showMapSelection {
							showMapSelection = false
						} else if showSettings {
							showSettings = false
						} else if showControls {
							log.Println("DEBUG: Exiting controls screen")
							showControls = false
						} else if showChangelog {
							showChangelog = false
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
						quitConfirmYes = false
						g.Manager.State = game.StateQuitConfirm
					}

				case tcell.KeyUp:
					if g.Manager.State == game.StateQuitConfirm {
						quitConfirmYes = true
					} else if g.Manager.State == game.StateMenu && showMapSelection {
						if mapSelectionIndex > 0 {
							mapSelectionIndex--
						}
					} else if g.Manager.State == game.StateMenu && !showControls && !showSettings && !showChangelog && !showUpdateScreen && !showMapSelection {
						if menuSelection > render.MenuStart {
							menuSelection--
						}
					} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorY--
						clampCursor(g)
					}

				case tcell.KeyDown:
					if g.Manager.State == game.StateQuitConfirm {
						quitConfirmYes = false
					} else if g.Manager.State == game.StateMenu && showMapSelection {
						if mapSelectionIndex < len(availableMaps)-1 {
							mapSelectionIndex++
						}
					} else if g.Manager.State == game.StateMenu && !showControls && !showSettings && !showChangelog && !showUpdateScreen && !showMapSelection {
						maxOpt := render.MaxMenuOption(updateAvailable)
						if menuSelection < maxOpt {
							menuSelection++
						}
					} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorY++
						clampCursor(g)
					}

				case tcell.KeyLeft:
					if g.Manager.State == game.StateQuitConfirm {
						quitConfirmYes = true
					} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorX--
						clampCursor(g)
					}

				case tcell.KeyRight:
					if g.Manager.State == game.StateQuitConfirm {
						quitConfirmYes = false
					} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
						g.CursorX++
						clampCursor(g)
					}

				case tcell.KeyEnter:
					if showUpdateScreen && updateProgress != nil && updateProgress.Done && updateProgress.Err == nil {
						os.Exit(0)
					}
					if g.Manager.State == game.StateQuitConfirm {
						if quitConfirmYes {
							running = false
							close(quit)
						} else {
							g.Manager.State = game.StateInWave
						}
						continue
					}
					if g.Manager.State == game.StateMenu && showMapSelection {
						if mapSelectionIndex >= 0 && mapSelectionIndex < len(availableMaps) {
							selectedMapID := availableMaps[mapSelectionIndex].ID
							log.Printf("DEBUG: Starting game with map %q", selectedMapID)
							m, err := mapdata.LoadMapByID(selectedMapID)
							if err != nil {
								log.Printf("ERROR: Failed to load map %q: %v", selectedMapID, err)
								m, _ = mapdata.DefaultMap()
							}
							g = game.NewGameFromMap(m)
							g.Manager.State = game.StatePreWave
							g.Manager.InterWaveTimer = 5.0
							showMapSelection = false
						}
						continue
					}
					if handleMenuSelect() {
						continue
					}

				case tcell.KeyRune:
					switch e.Rune() {
					case 'w', 'W':
						if g.Manager.State == game.StateQuitConfirm {
							quitConfirmYes = true
						} else if g.Manager.State == game.StateMenu && showMapSelection {
							if mapSelectionIndex > 0 {
								mapSelectionIndex--
							}
						} else if g.Manager.State == game.StateMenu && !showControls && !showSettings && !showChangelog && !showUpdateScreen && !showMapSelection {
							if menuSelection > render.MenuStart {
								menuSelection--
							}
						} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.CursorY--
							clampCursor(g)
						}

					case 's', 'S':
						if g.Manager.State == game.StateQuitConfirm {
							quitConfirmYes = false
						} else if g.Manager.State == game.StateMenu && showMapSelection {
							if mapSelectionIndex < len(availableMaps)-1 {
								mapSelectionIndex++
							}
						} else if g.Manager.State == game.StateMenu && !showControls && !showSettings && !showChangelog && !showUpdateScreen && !showMapSelection {
							maxOpt := render.MaxMenuOption(updateAvailable)
							if menuSelection < maxOpt {
								menuSelection++
							}
						} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.CursorY++
							clampCursor(g)
						}

					case 'a', 'A':
						if g.Manager.State == game.StateQuitConfirm {
							quitConfirmYes = true
						} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
							g.CursorX--
							clampCursor(g)
						}

					case 'd', 'D':
						if g.Manager.State == game.StateQuitConfirm {
							quitConfirmYes = false
						} else if g.Manager.State != game.StateMenu && g.Manager.State != game.StateQuitConfirm {
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
						if g.Manager.State == game.StateMenu && showMapSelection {
							if mapSelectionIndex >= 0 && mapSelectionIndex < len(availableMaps) {
								selectedMapID := availableMaps[mapSelectionIndex].ID
								log.Printf("DEBUG: Starting game with map %q", selectedMapID)
								m, err := mapdata.LoadMapByID(selectedMapID)
								if err != nil {
									log.Printf("ERROR: Failed to load map %q: %v", selectedMapID, err)
									m, _ = mapdata.DefaultMap()
								}
								g = game.NewGameFromMap(m)
								g.Manager.State = game.StatePreWave
								g.Manager.InterWaveTimer = 5.0
								showMapSelection = false
							}
							continue
						}
						if showUpdateScreen && updateProgress != nil && updateProgress.Done && updateProgress.Err == nil {
							os.Exit(0)
						}
						if g.Manager.State == game.StateMenu {
							if handleMenuSelect() {
								continue
							}
						} else if g.Manager.State == game.StateQuitConfirm {
							if quitConfirmYes {
								running = false
								close(quit)
							} else {
								g.Manager.State = game.StateInWave
							}
						} else if g.Manager.Mode == game.ModeBuild {
							if g.PlaceTower(entities.TowerBasic) {
								log.Printf("DEBUG: Tower placed at (%d, %d)", g.CursorX, g.CursorY)
								g.Manager.Mode = game.ModeNormal
							} else {
								log.Printf("DEBUG: Failed to place tower at (%d, %d)", g.CursorX, g.CursorY)
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

func showChangelogScreen(screen tcell.Screen, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("changelog read: %v", err)
		return
	}
	changelog := strings.TrimSpace(string(content))
	if changelog == "" {
		changelog = "No changelog available."
	}

	events := make(chan tcell.Event, 10)
	go screen.ChannelEvents(events, nil)

	for {
		screen.Clear()
		render.DrawChangelog(screen, changelog)
		screen.Show()

		ev := <-events
		if _, ok := ev.(*tcell.EventKey); ok {
			break
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
