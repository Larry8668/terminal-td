package game

import (
	"log"
	"terminal-td/internal/entities"
	mapdata "terminal-td/internal/map"
	"time"
)

type Game struct {
	Grid    *mapdata.Grid
	Path    mapdata.Path
	Enemies []*entities.Enemy
	Towers  []*entities.Tower

	Wave  WaveManager
	Base  Base
	Speed float64

	Money int

	Score      Score
	Difficulty Difficulty

	Manager *GameManager

	CursorX int
	CursorY int
}

func NewGame() *Game {
	grid := mapdata.NewGrid(80, 25)
	path := mapdata.DefaultPath()
	mapdata.ApplyPath(grid, path)

	g := &Game{
		Grid:   grid,
		Path:   path,
		Towers: []*entities.Tower{},

		Money: 500,

		CursorX: grid.Width / 2,
		CursorY: grid.Height / 2,
	}

	g.Base = Base{
		HP: 10,
	}

	g.Speed = 1.0
	g.Difficulty = Difficulty{
		SpeedMultiplier: 1.0,
		SpawnMultiplier: 1.0,
		CountBonus:      0,
	}

	g.Wave = WaveManager{
		CurrentWave:    1,
		TotalWaves:     5,
		EnemiesPerWave: 5,

		SpawnInterval: 1 * time.Second,
	}

	g.Manager = NewGameManager(g.Wave.TotalWaves, 5)

	return g
}

func (g *Game) spawnEnemy() {
	enemy := entities.NewEnemy(g.Path)

	enemy.Speed *= g.Difficulty.SpeedMultiplier

	g.Enemies = append(g.Enemies, enemy)

	g.Wave.EnemiesSpawned++
	g.Wave.EnemiesAlive++
}

func (g *Game) updateSpawning(dt float64) {
	w := &g.Wave

	delta := time.Duration(dt * float64(time.Second))

	log.Println(
		"Wave:", w.CurrentWave,
		"Spawning:", w.Spawning,
		"Spawned:", w.EnemiesSpawned,
		"Alive:", w.EnemiesAlive,
	)

	if !g.Manager.IsSimulationRunning() {
		return
	}

	if !w.Spawning && !w.SpawnFinished {
		log.Println("Wave started")
		w.Spawning = true
		w.EnemiesSpawned = 0
		w.SpawnTimer = 0
	}

	w.SpawnTimer += delta

	if w.SpawnTimer >= w.SpawnInterval && w.EnemiesSpawned < w.EnemiesPerWave {
		w.SpawnTimer = 0
		g.spawnEnemy()
	}

	if w.EnemiesSpawned == w.EnemiesPerWave {
		log.Println("Wave spawn finished")
		w.Spawning = false
		w.SpawnFinished = true
	}
}

func (g *Game) updateEnemies(dt float64) {
	alive := []*entities.Enemy{}

	for _, e := range g.Enemies {
		e.Update(dt)

		if e.ReachedBase {
			g.Base.HP--
			g.Wave.EnemiesAlive--
			if g.Base.HP <= 0 {
				g.Manager.OnBaseDestroyed()
			}
			continue
		}

		alive = append(alive, e)
	}

	g.Enemies = alive
}

func (g *Game) updateWaveState() {
	w := &g.Wave

	if !w.Spawning && w.EnemiesAlive == 0 && w.EnemiesSpawned == w.EnemiesPerWave {

		log.Println("Wave cleared")

		g.Score.WavesCleared++
		g.Score.Points += 100

		g.Difficulty.SpeedMultiplier += 0.1
		g.Difficulty.SpawnMultiplier += 0.05
		g.Difficulty.CountBonus += 1

		g.Manager.EndWave()
		w.SpawnFinished = false

		if w.CurrentWave < w.TotalWaves {
			w.CurrentWave++
			w.EnemiesPerWave += 2 + g.Difficulty.CountBonus

			w.SpawnInterval = time.Duration(float64(w.SpawnInterval) / g.Difficulty.SpawnMultiplier)
		}
	}
}

func (g *Game) Update(dt float64) {
	scaled := dt * g.Speed

	g.updateSpawning(scaled)
	g.updateEnemies(scaled)
	g.updateWaveState()
}

func (g *Game) Reset() {
	g.Enemies = nil
	g.Towers = []*entities.Tower{}

	g.Base.HP = 10
	g.Money = 500

	g.CursorX = g.Grid.Width / 2
	g.CursorY = g.Grid.Height / 2

	g.Wave = WaveManager{
		CurrentWave:    1,
		TotalWaves:     5,
		EnemiesPerWave: 5,
		SpawnInterval:  1 * time.Second,
	}

	g.Difficulty = Difficulty{
		SpeedMultiplier: 1.0,
		SpawnMultiplier: 1.0,
		CountBonus:      0,
	}

	g.Score = Score{}

	g.Manager.Reset()
}

func (g *Game) CanPlaceTower(x, y int) bool {
	if x < 0 || x >= g.Grid.Width || y < 0 || y >= g.Grid.Height {
		return false
	}

	if g.Grid.Tiles[y][x] == mapdata.PathTile || g.Grid.Tiles[y][x] == mapdata.SpawnTile || g.Grid.Tiles[y][x] == mapdata.BaseTile {
		return false
	}

	for _, tower := range g.Towers {
		if tower.X == x && tower.Y == y {
			return false
		}
	}

	return true
}

func (g *Game) PlaceTower(towerType entities.TowerType) bool {
	if !g.CanPlaceTower(g.CursorX, g.CursorY) {
		return false
	}

	templates := GetTowerTemplates()
	template, exists := templates[towerType]

	if !exists {
		return false
	}

	if g.Money < template.Cost {
		return false
	}

	tower := entities.NewTower(g.CursorX, g.CursorY, towerType)
	g.Towers = append(g.Towers, tower)
	g.Money -= template.Cost

	return true
}

func (g *Game) GetTowerAt(x, y int) *entities.Tower {
	for _, tower := range g.Towers {
		if tower.X == x && tower.Y == y {
			return tower
		}
	}
	return nil
}
