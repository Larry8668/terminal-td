package game

import (
	"log"
	// "math"
	"terminal-td/internal/entities"
	"terminal-td/internal/enemies"
	mapdata "terminal-td/internal/map"
	waves "terminal-td/internal/waves"
	"time"
)

type Game struct {
	Map         *mapdata.GameMap
	Grid        *mapdata.Grid
	Path        mapdata.Path
	Enemies     []*entities.Enemy
	Towers      []*entities.Tower
	Projectiles []*entities.Projectile

	Wave         *waves.WaveManager
	EnemyDB      *enemies.EnemyDatabase
	LegacyWave   WaveManager // kept for backward compat during transition
	Base         Base
	Speed        float64

	Money int

	Score      Score
	Difficulty Difficulty

	Manager *GameManager

	CursorX int
	CursorY int
}

// NewGame loads the default map and returns a Game. Use NewGameFromMap for custom maps.
func NewGame() *Game {
	m, err := mapdata.DefaultMap()
	if err != nil {
		log.Printf("load default map: %v", err)
		return newGameLegacy()
	}
	return NewGameFromMap(m)
}

// NewGameFromMap builds a Game from a data-driven map (spawns, paths, base from map).
func NewGameFromMap(m *mapdata.GameMap) *Game {
	grid := m.Grid
	path := m.PrimaryPath()

	enemyDB, err := enemies.DefaultEnemies()
	if err != nil {
		log.Printf("load enemies: %v, using fallback", err)
		enemyDB = &enemies.EnemyDatabase{Enemies: make(map[string]enemies.EnemyDef)}
	}

	waveDefs, err := waves.LoadWavesForMap(m.ID)
	if err != nil {
		log.Printf("load waves for map %q: %v, trying fallback", m.ID, err)
		waveDefs, err = waves.DefaultWaves()
		if err != nil {
			log.Printf("load default waves: %v, using empty waves", err)
			waveDefs = []waves.WaveDef{}
		}
	}

	spawnIDs := make(map[string]bool)
	for _, spawn := range m.Spawns {
		spawnIDs[spawn.ID] = true
	}
	if err := waves.ValidateWavesAgainstMap(waveDefs, spawnIDs); err != nil {
		log.Printf("WARN: wave validation failed for map %q: %v", m.ID, err)
	}

	waveMgr := waves.NewWaveManager(waveDefs)

	g := &Game{
		Map:         m,
		Grid:        grid,
		Path:        path,
		Towers:      []*entities.Tower{},
		Projectiles: []*entities.Projectile{},

		Wave:    waveMgr,
		EnemyDB: enemyDB,

		Money: 500,

		CursorX: grid.Width / 2,
		CursorY: grid.Height / 2,
	}

	g.Base = Base{
		X:  m.Base.X,
		Y:  m.Base.Y,
		HP: m.Base.HP,
	}

	g.Speed = 1.0
	g.Difficulty = Difficulty{
		SpeedMultiplier: 1.0,
		SpawnMultiplier: 1.0,
		CountBonus:      0,
	}

	totalWaves := len(waveDefs)
	if totalWaves == 0 {
		totalWaves = 5
	}
	g.Manager = NewGameManager(totalWaves, 5)

	return g
}

// newGameLegacy is fallback when default map fails to load (hardcoded grid/path/base).
func newGameLegacy() *Game {
	grid := mapdata.NewGrid(80, 25)
	path := mapdata.DefaultPath()
	mapdata.ApplyPath(grid, path)

	g := &Game{
		Map:         nil,
		Grid:        grid,
		Path:        path,
		Towers:      []*entities.Tower{},
		Projectiles: []*entities.Projectile{},

		Money: 500,

		CursorX: grid.Width / 2,
		CursorY: grid.Height / 2,
	}

	g.Base = Base{
		X:  79,
		Y:  22,
		HP: 10,
	}

	g.Speed = 1.0
	g.Difficulty = Difficulty{
		SpeedMultiplier: 1.0,
		SpawnMultiplier: 1.0,
		CountBonus:      0,
	}

	g.LegacyWave = WaveManager{
		CurrentWave:    1,
		TotalWaves:     5,
		EnemiesPerWave: 5,

		SpawnInterval: 1 * time.Second,
	}

	g.Manager = NewGameManager(g.LegacyWave.TotalWaves, 5)

	return g
}

func (g *Game) spawnEnemy(enemyTypeID string, spawnID string) {
	var path mapdata.Path
	if g.Map != nil {
		path = g.Map.Paths[spawnID]
		if len(path.Points) == 0 {
			path = g.Path
		}
	} else {
		path = g.Path
	}

	var enemy *entities.Enemy
	if g.EnemyDB != nil {
		def := g.EnemyDB.Get(enemyTypeID)
		if def != nil {
			enemy = entities.NewEnemyFromDef(def.HP, def.Speed, def.Reward, def.ID, path)
		} else {
			log.Printf("WARN: enemy type %q not found, using basic", enemyTypeID)
			enemy = entities.NewEnemy(path)
		}
	} else {
		enemy = entities.NewEnemy(path)
	}

	enemy.Speed *= g.Difficulty.SpeedMultiplier

	g.Enemies = append(g.Enemies, enemy)

	if g.Wave != nil {
		g.Wave.EnemiesAlive++
	}
	log.Printf("DEBUG: Enemy spawned (type=%s spawn=%s, Alive: %d)", enemyTypeID, spawnID, g.Wave.EnemiesAlive)
}

func (g *Game) updateSpawning(dt float64) {
	if !g.Manager.IsSimulationRunning() {
		return
	}

	if g.Wave == nil {
		g.updateSpawningLegacy(dt)
		return
	}

	if len(g.Wave.ActiveGroups) == 0 && g.Wave.CurrentWave < len(g.Wave.Waves) {
		log.Printf("DEBUG: Wave %d spawning started", g.Wave.CurrentWave+1)
		g.Wave.StartWave()
	}

	for i := range g.Wave.ActiveGroups {
		group := &g.Wave.ActiveGroups[i]
		if group.Completed {
			continue
		}

		if group.DelayTimer > 0 {
			group.DelayTimer -= dt
			continue
		}

		group.Timer += dt

		if group.Timer >= group.Def.Interval && group.Spawned < group.Def.Count {
			group.Timer = 0
			g.spawnEnemy(group.Def.EnemyType, group.Def.SpawnID)
			group.Spawned++

			if group.Spawned >= group.Def.Count {
				group.Completed = true
				log.Printf("DEBUG: Group completed (spawn=%s type=%s count=%d)", group.Def.SpawnID, group.Def.EnemyType, group.Spawned)
			}
		}
	}
}

func (g *Game) updateSpawningLegacy(dt float64) {
	w := &g.LegacyWave

	delta := time.Duration(dt * float64(time.Second))

	if !g.Manager.IsSimulationRunning() {
		return
	}

	if !w.Spawning && !w.SpawnFinished {
		log.Printf("DEBUG: Wave %d spawning started (legacy)", w.CurrentWave)
		w.Spawning = true
		w.EnemiesSpawned = 0
		w.SpawnTimer = 0
	}

	w.SpawnTimer += delta

	if w.SpawnTimer >= w.SpawnInterval && w.EnemiesSpawned < w.EnemiesPerWave {
		w.SpawnTimer = 0
		g.spawnEnemy("basic", "default")
	}

	if w.EnemiesSpawned == w.EnemiesPerWave {
		log.Printf("DEBUG: Wave %d spawn finished (%d enemies spawned)", w.CurrentWave, w.EnemiesSpawned)
		w.Spawning = false
		w.SpawnFinished = true
	}
}

func (g *Game) updateEnemies(dt float64) {
	alive := []*entities.Enemy{}

	for _, e := range g.Enemies {
		if e.HP <= 0 {
			if g.Wave != nil {
				g.Wave.EnemiesAlive--
			} else {
				g.LegacyWave.EnemiesAlive--
			}
			log.Printf("DEBUG: Dead enemy removed (Alive: %d)", g.GetEnemiesAlive())
			continue
		}

		e.Update(dt)

		if e.ReachedBase {
			g.Base.HP--
			if g.Wave != nil {
				g.Wave.EnemiesAlive--
			} else {
				g.LegacyWave.EnemiesAlive--
			}
			log.Printf("DEBUG: Enemy reached base (Base HP: %d, Enemies alive: %d)", g.Base.HP, g.GetEnemiesAlive())
			if g.Base.HP <= 0 {
				log.Println("DEBUG: Base destroyed - game lost")
				g.Manager.OnBaseDestroyed()
			}
			continue
		}

		alive = append(alive, e)
	}

	g.Enemies = alive
}

func (g *Game) GetEnemiesAlive() int {
	if g.Wave != nil {
		return g.Wave.EnemiesAlive
	}
	return g.LegacyWave.EnemiesAlive
}

func (g *Game) GetCurrentWave() int {
	if g.Wave != nil {
		return g.Wave.CurrentWave + 1
	}
	return g.LegacyWave.CurrentWave
}

func (g *Game) GetTotalWaves() int {
	if g.Wave != nil {
		return len(g.Wave.Waves)
	}
	return g.LegacyWave.TotalWaves
}

// GetNextWaveSpawnIDs returns spawn IDs that will be used in the next wave (for highlighting).
func (g *Game) GetNextWaveSpawnIDs() map[string]bool {
	spawnIDs := make(map[string]bool)
	if g.Wave == nil || g.Map == nil {
		return spawnIDs
	}
	nextWaveIndex := g.Wave.CurrentWave
	if nextWaveIndex >= 0 && nextWaveIndex < len(g.Wave.Waves) {
		wave := g.Wave.Waves[nextWaveIndex]
		for _, group := range wave.Groups {
			spawnIDs[group.SpawnID] = true
		}
	}
	return spawnIDs
}

func (g *Game) updateWaveState() {
	if g.Wave == nil {
		g.updateWaveStateLegacy()
		return
	}

	if g.Wave.IsWaveComplete() {
		waveNum := g.Wave.CurrentWave + 1
		log.Printf("DEBUG: Wave %d cleared! Score: %d (+100)", waveNum, g.Score.Points+100)

		g.Score.WavesCleared++
		g.Score.Points += 100

		g.Difficulty.SpeedMultiplier += 0.1
		g.Difficulty.SpawnMultiplier += 0.05
		g.Difficulty.CountBonus += 1

		g.Manager.EndWave()

		if g.Wave.NextWave() {
			log.Printf("DEBUG: Starting wave %d", g.Wave.CurrentWave+1)
		} else {
			log.Println("DEBUG: All waves completed!")
		}
	}
}

func (g *Game) updateWaveStateLegacy() {
	w := &g.LegacyWave

	if !w.Spawning && w.EnemiesAlive == 0 && w.EnemiesSpawned == w.EnemiesPerWave {
		log.Printf("DEBUG: Wave %d cleared! Score: %d (+100)", w.CurrentWave, g.Score.Points+100)

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
	g.updateTowers(scaled)
	g.updateProjectiles(scaled)
	g.updateEnemies(scaled)
	g.updateWaveState()
}

func (g *Game) Reset() {
	g.Enemies = nil
	g.Towers = []*entities.Tower{}
	g.Projectiles = []*entities.Projectile{}

	if g.Map != nil {
		g.Base.HP = g.Map.Base.HP
	} else {
		g.Base.HP = 10
	}
	g.Money = 500

	g.CursorX = g.Grid.Width / 2
	g.CursorY = g.Grid.Height / 2

	if g.Wave != nil {
		g.Wave.CurrentWave = 0
		g.Wave.ActiveGroups = nil
		g.Wave.EnemiesAlive = 0
	} else {
		g.LegacyWave = WaveManager{
			CurrentWave:    1,
			TotalWaves:     5,
			EnemiesPerWave: 5,
			SpawnInterval:  1 * time.Second,
		}
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
		log.Printf("DEBUG: Cannot place tower at (%d, %d) - invalid location", g.CursorX, g.CursorY)
		return false
	}

	templates := GetTowerTemplates()
	template, exists := templates[towerType]

	if !exists {
		log.Printf("DEBUG: Tower type %d does not exist", towerType)
		return false
	}

	if g.Money < template.Cost {
		log.Printf("DEBUG: Insufficient funds to place tower (Have: %d, Need: %d)", g.Money, template.Cost)
		return false
	}

	tower := entities.NewTower(g.CursorX, g.CursorY, towerType)
	g.Towers = append(g.Towers, tower)
	g.Money -= template.Cost
	log.Printf("DEBUG: Tower placed at (%d, %d), Money remaining: %d, Total towers: %d", g.CursorX, g.CursorY, g.Money, len(g.Towers))

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

func (g *Game) isEnemyInRange(tower *entities.Tower, enemy *entities.Enemy) bool {
	dist := tower.DistanceTo(enemy.X, enemy.Y)
	return dist <= tower.Range && enemy.HP > 0
}

func (g *Game) updateTowers(dt float64) {
	for _, tower := range g.Towers {
		if tower.Cooldown > 0 {
			tower.Cooldown = maxFloat(0, tower.Cooldown-dt)
		}

		if tower.Target == nil || !g.isEnemyInRange(tower, tower.Target) {
			oldTarget := tower.Target
			tower.Target = g.findClosestEnemyInRange(tower)
			if tower.Target != oldTarget && tower.Target != nil {
				log.Printf("DEBUG: Tower at (%d, %d) acquired new target", tower.X, tower.Y)
			}
		}

		if tower.Cooldown <= 0 && tower.Target != nil {
			g.fireTower(tower)
			tower.Cooldown = 1.0 / tower.FireRate
		}
	}
}

func (g *Game) findClosestEnemyInRange(tower *entities.Tower) *entities.Enemy {
	var closest *entities.Enemy
	closestDist := tower.Range + 1.0

	for _, enemy := range g.Enemies {
		if enemy.HP <= 0 {
			continue
		}

		dist := tower.DistanceTo(enemy.X, enemy.Y)
		if dist <= tower.Range && dist < closestDist {
			closest = enemy
			closestDist = dist
		}
	}

	return closest
}

func (g *Game) fireTower(tower *entities.Tower) {
	if tower.Target == nil {
		return
	}

	projectile := entities.NewProjectile(
		float64(tower.X),
		float64(tower.Y),
		tower.Target,
		20.0,
		tower.Damage,
	)

	g.Projectiles = append(g.Projectiles, projectile)
	log.Printf("DEBUG: Tower at (%d, %d) fired at enemy (HP: %.1f), Projectiles: %d", tower.X, tower.Y, tower.Target.HP, len(g.Projectiles))
}

func (g *Game) updateProjectiles(dt float64) {
	active := []*entities.Projectile{}

	for _, proj := range g.Projectiles {
		proj.Update(dt)

		if proj.HasHit {
				if proj.TargetEnemy != nil && proj.TargetEnemy.HP > 0 {
					proj.TargetEnemy.HP -= proj.Damage
					if proj.TargetEnemy.HP <= 0 {
						reward := proj.TargetEnemy.Reward
						if reward == 0 {
							reward = 10
						}
						g.Money += reward
						g.Score.Points += reward
					}
				}
			continue
		}
		if proj.TargetEnemy != nil && proj.TargetEnemy.HP > 0 {
			active = append(active, proj)
		}
	}
	g.Projectiles = active
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
