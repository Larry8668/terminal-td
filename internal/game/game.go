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

	Wave WaveManager
	Base Base
}

func NewGame() *Game {
	grid := mapdata.NewGrid(40, 20)
	path := mapdata.DefaultPath()
	mapdata.ApplyPath(grid, path)

	// enemy := entities.NewEnemy(path)

	g := &Game{
		Grid: grid,
		Path: path,
	}

	g.Base = Base{
		HP: 10,
	}

	g.Wave = WaveManager{
		CurrentWave:    1,
		TotalWaves:     5,
		EnemiesPerWave: 5,

		SpawnInterval: 1 * time.Second,
		WaveCooldown:  5 * time.Second,
		CooldownTimer: 5 * time.Second,
	}

	return g
}

func (g *Game) spawnEnemy() {
	enemy := entities.NewEnemy(g.Path)
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
		"Cooldown:", w.CooldownTimer,
	)

	if !w.Spawning {
		if w.EnemiesAlive > 0 {
			return
		}

		if w.CurrentWave > w.TotalWaves {
			return
		}

		w.CooldownTimer -= delta

		if w.CooldownTimer <= 0 {
			log.Println("Wave spawning started")

			w.Spawning = true
			w.EnemiesSpawned = 0
			w.SpawnTimer = 0
		}
		return
	}

	w.SpawnTimer += delta

	if w.SpawnTimer >= w.SpawnInterval && w.EnemiesSpawned < w.EnemiesPerWave {
		w.SpawnTimer = 0
		g.spawnEnemy()
	}

	if w.EnemiesSpawned == w.EnemiesPerWave {
		log.Println("Wave spawn finished")
		w.Spawning = false
	}
}

func (g *Game) updateEnemies(dt float64) {
	alive := []*entities.Enemy{}

	for _, e := range g.Enemies {
		e.Update(dt)

		if e.ReachedBase {
			g.Base.HP--
			g.Wave.EnemiesAlive--
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

		if w.CurrentWave < w.TotalWaves {
			w.CurrentWave++
			w.EnemiesPerWave += 2
			w.CooldownTimer = w.WaveCooldown
		}
	}
}

func (g *Game) Update(dt float64) {
	g.updateSpawning(dt)
	g.updateEnemies(dt)
	g.updateWaveState()
}
