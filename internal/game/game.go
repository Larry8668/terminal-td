package game

import (
	"fmt"
	"log"
	"terminal-td/internal/enemies"
	"terminal-td/internal/entities"
	"terminal-td/internal/flow"
	mapdata "terminal-td/internal/map"
	waves "terminal-td/internal/waves"
	"time"
)

// Wall links two towers and blocks path tiles on the segment between them.
type Wall struct {
	Ax, Ay, Bx, By int
}

type Game struct {
	Map         *mapdata.GameMap
	Grid        *mapdata.Grid
	Path        mapdata.Path
	Enemies     []*entities.Enemy
	Towers      []*entities.Tower
	Projectiles []*entities.Projectile
	Walls       []Wall

	Wave       *waves.WaveManager
	EnemyDB    *enemies.EnemyDatabase
	LegacyWave WaveManager // kept for backward compat during transition
	Base       Base
	Speed      float64

	FlowField *flow.Field
	Walkable  [][]bool

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

	walkable := flow.BuildWalkability(grid)
	flowField := flow.Compute(grid.Width, grid.Height, walkable, m.Base.X, m.Base.Y)

	g := &Game{
		Map:         m,
		Grid:        grid,
		Path:        path,
		Towers:      []*entities.Tower{},
		Projectiles: []*entities.Projectile{},
		Walls:       nil,

		Wave:      waveMgr,
		EnemyDB:   enemyDB,
		FlowField: flowField,
		Walkable:  walkable,

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

	walkable := flow.BuildWalkability(grid)
	flowField := flow.Compute(grid.Width, grid.Height, walkable, 79, 22)

	g := &Game{
		Map:         nil,
		Grid:        grid,
		Path:        path,
		Towers:      []*entities.Tower{},
		Projectiles: []*entities.Projectile{},
		Walls:       nil,

		FlowField: flowField,
		Walkable:  walkable,

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
	log.Printf("DEBUG: Enemy spawned (type=%s spawn=%s pos=(%.1f,%.1f) Alive: %d)", enemyTypeID, spawnID, enemy.X, enemy.Y, g.Wave.EnemiesAlive)
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

const flowReachedBaseDist = 0.5

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

		if g.FlowField != nil {
			dist, dir := g.FlowField.AtFloat(e.X, e.Y)
			if dist >= flow.Inf {
				log.Printf("DEBUG: flow unreachable at (%.1f,%.1f) dist=Inf → marking reached base", e.X, e.Y)
				e.ReachedBase = true
			} else if dist < flowReachedBaseDist {
				e.ReachedBase = true
			} else {
				e.UpdateFlow(dt, dir.X, dir.Y)
			}
		} else {
			e.Update(dt)
		}

		if e.ReachedBase {
			g.Base.HP--
			if g.Wave != nil {
				g.Wave.EnemiesAlive--
			} else {
				g.LegacyWave.EnemiesAlive--
			}
			log.Printf("DEBUG: Enemy reached base at (%.1f,%.1f) Base HP: %d Enemies alive: %d", e.X, e.Y, g.Base.HP, g.GetEnemiesAlive())
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

// FlowDebugString returns a short debug line when flow field is active and there are enemies (for on-screen debug).
func (g *Game) FlowDebugString() string {
	if g.FlowField == nil || len(g.Enemies) == 0 {
		return ""
	}
	e := g.Enemies[0]
	dist, _ := g.FlowField.AtFloat(e.X, e.Y)
	if dist >= flow.Inf {
		return "Flow: unreachable"
	}
	return fmt.Sprintf("Flow dist: %.1f pos:(%.0f,%.0f)", dist, e.X, e.Y)
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

// TracePathsForNextWave returns flow-field paths from each active spawn to base for pre-wave preview.
// Each path is a slice of (x,y) tiles. Only valid when FlowField and Map are set.
func (g *Game) TracePathsForNextWave() [][]flow.Tile {
	if g.FlowField == nil || g.Map == nil {
		return nil
	}
	spawnIDs := g.GetNextWaveSpawnIDs()
	if len(spawnIDs) == 0 {
		return nil
	}
	var paths [][]flow.Tile
	for _, spawn := range g.Map.Spawns {
		if !spawnIDs[spawn.ID] {
			continue
		}
		path := g.FlowField.TracePath(spawn.X, spawn.Y, g.Base.X, g.Base.Y)
		if len(path) > 0 {
			paths = append(paths, path)
		}
	}
	return paths
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
	g.Walls = nil

	g.RecomputeFlow()

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

// ComputeBlockedTiles returns path tiles that lie on any wall segment (spawn/base stay walkable).
func (g *Game) ComputeBlockedTiles() [][2]int {
	seen := make(map[[2]int]bool)
	var out [][2]int
	for _, w := range g.Walls {
		for _, p := range mapdata.TilesOnSegment(w.Ax, w.Ay, w.Bx, w.By) {
			x, y := p[0], p[1]
			if seen[[2]int{x, y}] {
				continue
			}
			if y < 0 || y >= g.Grid.Height || x < 0 || x >= g.Grid.Width {
				continue
			}
			if g.Grid.Tiles[y][x] != mapdata.PathTile {
				continue
			}
			seen[[2]int{x, y}] = true
			out = append(out, [2]int{x, y})
		}
	}
	return out
}

// RecomputeFlow rebuilds walkability (including wall blocks) and flow field.
func (g *Game) RecomputeFlow() {
	blocked := g.ComputeBlockedTiles()
	g.Walkable = flow.BuildWalkabilityWithBlocked(g.Grid, blocked)
	g.FlowField = flow.Compute(g.Grid.Width, g.Grid.Height, g.Walkable, g.Base.X, g.Base.Y)
	log.Printf("DEBUG: Flow recomputed (blocked tiles: %d)", len(blocked))
}

const maxWallLinkDist = 4

// GetLinkableTowers returns positions (x,y) of towers that can form a wall with the tower at (ax,ay). Order is stable for HUD numbering.
func (g *Game) GetLinkableTowers(ax, ay int) [][2]int {
	var out [][2]int
	for _, t := range g.Towers {
		bx, by := t.X, t.Y
		if ax == bx && ay == by {
			continue
		}
		dx := ax - bx
		if dx < 0 {
			dx = -dx
		}
		dy := ay - by
		if dy < 0 {
			dy = -dy
		}
		if dx+dy > maxWallLinkDist {
			continue
		}
		exists := false
		for _, w := range g.Walls {
			if (w.Ax == ax && w.Ay == ay && w.Bx == bx && w.By == by) || (w.Ax == bx && w.Ay == by && w.Bx == ax && w.By == ay) {
				exists = true
				break
			}
		}
		if !exists && !g.WouldDisconnectSpawnsFromBase(ax, ay, bx, by) {
			out = append(out, [2]int{bx, by})
		}
	}
	return out
}

// GetWallsForTower returns positions (x,y) of towers connected by a wall to the tower at (ax,ay).
func (g *Game) GetWallsForTower(ax, ay int) [][2]int {
	var out [][2]int
	for _, w := range g.Walls {
		if w.Ax == ax && w.Ay == ay {
			out = append(out, [2]int{w.Bx, w.By})
		} else if w.Bx == ax && w.By == ay {
			out = append(out, [2]int{w.Ax, w.Ay})
		}
	}
	return out
}

// RemoveWall removes the wall between (ax,ay) and (bx,by). Returns true if a wall was removed.
func (g *Game) RemoveWall(ax, ay, bx, by int) bool {
	for i, w := range g.Walls {
		if (w.Ax == ax && w.Ay == ay && w.Bx == bx && w.By == by) || (w.Ax == bx && w.Ay == by && w.Bx == ax && w.By == ay) {
			g.Walls = append(g.Walls[:i], g.Walls[i+1:]...)
			g.RecomputeFlow()
			log.Printf("DEBUG: Wall removed (%d,%d)-(%d,%d)", ax, ay, bx, by)
			return true
		}
	}
	return false
}

const sellRefundPercent = 50

// SellTower removes the tower at (x,y), refunds part of cost, and removes any walls using it. Returns true if sold.
func (g *Game) SellTower(x, y int) bool {
	tower := g.GetTowerAt(x, y)
	if tower == nil {
		return false
	}
	templates := GetTowerTemplates()
	template, ok := templates[tower.Type]
	if !ok {
		template = TowerTemplate{Cost: 50}
	}
	refund := (template.Cost * sellRefundPercent) / 100
	g.Money += refund

	var newWalls []Wall
	for _, w := range g.Walls {
		if (w.Ax == x && w.Ay == y) || (w.Bx == x && w.By == y) {
			continue
		}
		newWalls = append(newWalls, w)
	}
	g.Walls = newWalls

	for i := range g.Towers {
		if g.Towers[i].X == x && g.Towers[i].Y == y {
			g.Towers = append(g.Towers[:i], g.Towers[i+1:]...)
			break
		}
	}
	g.RecomputeFlow()
	log.Printf("DEBUG: Tower sold at (%d,%d), refund %d", x, y, refund)
	return true
}

// AddWall links two towers with a wall; path tiles on the segment become blocked. Returns false if invalid.
func (g *Game) AddWall(ax, ay, bx, by int) bool {
	if ax == bx && ay == by {
		return false
	}
	if g.GetTowerAt(ax, ay) == nil || g.GetTowerAt(bx, by) == nil {
		log.Printf("DEBUG: AddWall: both cells must have towers")
		return false
	}
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx+dy > maxWallLinkDist {
		log.Printf("DEBUG: AddWall: towers too far (max %d)", maxWallLinkDist)
		return false
	}
	for _, w := range g.Walls {
		if (w.Ax == ax && w.Ay == ay && w.Bx == bx && w.By == by) || (w.Ax == bx && w.Ay == by && w.Bx == ax && w.By == ay) {
			log.Printf("DEBUG: AddWall: wall already exists")
			return false
		}
	}
	if g.WouldDisconnectSpawnsFromBase(ax, ay, bx, by) {
		log.Printf("DEBUG: AddWall: would block only path to base, rejected")
		return false
	}
	g.Walls = append(g.Walls, Wall{Ax: ax, Ay: ay, Bx: bx, By: by})
	g.RecomputeFlow()
	log.Printf("DEBUG: Wall added (%d,%d)-(%d,%d)", ax, ay, bx, by)
	return true
}

// WouldDisconnectSpawnsFromBase returns true if adding a wall from (ax,ay) to (bx,by) would leave any spawn with no path to base.
func (g *Game) WouldDisconnectSpawnsFromBase(ax, ay, bx, by int) bool {
	if g.Map == nil || len(g.Map.Spawns) == 0 {
		return false
	}
	blocked := g.ComputeBlockedTiles()
	seen := make(map[[2]int]bool)
	for _, p := range blocked {
		seen[p] = true
	}
	for _, p := range mapdata.TilesOnSegment(ax, ay, bx, by) {
		x, y := p[0], p[1]
		if x < 0 || x >= g.Grid.Width || y < 0 || y >= g.Grid.Height {
			continue
		}
		if g.Grid.Tiles[y][x] != mapdata.PathTile {
			continue
		}
		if !seen[[2]int{x, y}] {
			seen[[2]int{x, y}] = true
			blocked = append(blocked, [2]int{x, y})
		}
	}
	walkable := flow.BuildWalkabilityWithBlocked(g.Grid, blocked)
	testField := flow.Compute(g.Grid.Width, g.Grid.Height, walkable, g.Base.X, g.Base.Y)
	for _, spawn := range g.Map.Spawns {
		dist, _ := testField.At(spawn.X, spawn.Y)
		if dist >= flow.Inf {
			return true
		}
	}
	return false
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
