// Package sim runs a headless, deterministic playthrough of a map+waves pair
// so a map can be difficulty-tested without a human or a terminal. It drives
// internal/game.Game directly with a fixed timestep — no tcell, no wall
// clock, no randomness — the same map+waves+bot combination always produces
// the same Result, which is what makes simulate_run useful for an AI agent
// iterating on map difficulty.
package sim

import (
	"fmt"

	"terminal-td/internal/entities"
	"terminal-td/internal/flow"
	"terminal-td/internal/game"
	mapdata "terminal-td/internal/map"
	"terminal-td/internal/waves"
)

const (
	// DefaultDt is the fixed simulation timestep in seconds. 1/30s mirrors a
	// typical frame rate closely enough to match the real game's behavior
	// (spawn intervals, projectile travel) without being so coarse that
	// timing-sensitive waves behave differently than in real play.
	DefaultDt = 1.0 / 30.0

	// DefaultMaxSimTime caps total simulated game-time (not wall-clock time)
	// at 30 minutes, guaranteeing the loop always terminates even for a
	// pathological map+waves combination that never resolves to a win or
	// loss (e.g. a bot that can't keep up but also never quite loses).
	DefaultMaxSimTime = 30 * 60.0

	// DefaultBudget is the game's normal starting money, used whenever a
	// Config doesn't specify a budget override.
	DefaultBudget = 500
)

// Config describes one simulation run.
type Config struct {
	// Map is the map to simulate. Required.
	Map *mapdata.GameMap

	// Waves overrides the map's own waves (embedded or user) when non-empty.
	// Validated against Map's spawn ids before the run starts.
	Waves []waves.WaveDef

	// Bot decides what to build before each wave. Defaults to NoneBot.
	Bot Bot

	// Budget overrides the starting money. 0 means DefaultBudget.
	Budget int

	// Dt overrides the fixed timestep. 0 means DefaultDt.
	Dt float64

	// MaxSimTime overrides the termination cap (seconds of simulated
	// game-time). 0 means DefaultMaxSimTime.
	MaxSimTime float64
}

// Result is the outcome of a simulation run. Field names/JSON tags match the
// contract already committed to by the simulate_run MCP tool and the
// `simulate` CLI subcommand (see docs/agent/PLAN.md Phase C3).
type Result struct {
	Outcome       string         `json:"outcome"` // "won", "lost", or "timeout"
	WavesTotal    int            `json:"waves_total"`
	WavesSurvived int            `json:"waves_survived"`
	BaseHPStart   int            `json:"base_hp_start"`
	BaseHPEnd     int            `json:"base_hp_end"`
	LeaksPerWave  []int          `json:"leaks_per_wave"`
	KillsPerWave  []int          `json:"kills_per_wave"`
	MoneyEnd      int            `json:"money_end"`
	TowersPlaced  int            `json:"towers_placed"`
	SimTimeS      float64        `json:"sim_time_s"`
	PathLengths   map[string]int `json:"path_lengths"`
}

// Run builds a Game from cfg and simulates it to completion (win, loss, or
// timeout), returning a Result. A non-nil error means the request itself was
// malformed (bad waves override) — an in-simulation loss is not an error,
// it's a normal Result with Outcome: "lost".
func Run(cfg Config) (*Result, error) {
	if cfg.Map == nil {
		return nil, fmt.Errorf("sim: Config.Map is required")
	}

	dt := cfg.Dt
	if dt <= 0 {
		dt = DefaultDt
	}
	maxSimTime := cfg.MaxSimTime
	if maxSimTime <= 0 {
		maxSimTime = DefaultMaxSimTime
	}
	bot := cfg.Bot
	if bot == nil {
		bot = NoneBot{}
	}

	g := game.NewGameFromMap(cfg.Map)

	if len(cfg.Waves) > 0 {
		spawnIDs := make(map[string]bool, len(cfg.Map.Spawns))
		for _, s := range cfg.Map.Spawns {
			spawnIDs[s.ID] = true
		}
		if err := waves.ValidateWavesAgainstMap(cfg.Waves, spawnIDs); err != nil {
			return nil, fmt.Errorf("sim: invalid waves override: %w", err)
		}
		g.Wave = waves.NewWaveManager(cfg.Waves)
		g.Manager.TotalWaves = len(cfg.Waves)
	}

	budget := cfg.Budget
	if budget <= 0 {
		budget = DefaultBudget
	}
	g.Money = budget

	totalWaves := g.GetTotalWaves()
	baseHPStart := g.Base.HP
	pathLens := pathLengths(g)

	killsPerWave := make([]int, totalWaves)
	leaksPerWave := make([]int, totalWaves)

	g.OnEnemyKilled = func(_ *entities.Enemy) {
		if idx := g.Manager.CurrentWave - 1; idx >= 0 && idx < len(killsPerWave) {
			killsPerWave[idx]++
		}
	}
	g.OnEnemyLeaked = func(_ *entities.Enemy) {
		if idx := g.Manager.CurrentWave - 1; idx >= 0 && idx < len(leaksPerWave) {
			leaksPerWave[idx]++
		}
	}

	// Real play waits out a 5s pre-wave countdown (see GameManager's
	// InterWaveDelay) purely for human pacing; that's meaningless for a
	// headless sim, so each pre-wave phase gives the bot one chance to act
	// and then starts the wave immediately. lastState tracks the previous
	// tick's state so the bot is invoked exactly once per wave boundary
	// (edge-triggered), not once per tick spent in StatePreWave.
	g.Manager.State = game.StatePreWave
	lastState := game.GameState(-1)

	var simTime float64
	outcome := "timeout"
	for simTime < maxSimTime {
		if g.Manager.State == game.StatePreWave && lastState != game.StatePreWave {
			bot.PreWave(g)
			g.Manager.StartWave()
		}
		lastState = g.Manager.State

		g.Manager.Update(dt)
		if g.Manager.IsSimulationRunning() {
			g.Update(dt)
		}
		simTime += dt

		if g.Manager.State == game.StateWon {
			outcome = "won"
			break
		}
		if g.Manager.State == game.StateLost {
			outcome = "lost"
			break
		}
	}

	return &Result{
		Outcome:       outcome,
		WavesTotal:    totalWaves,
		WavesSurvived: g.Score.WavesCleared,
		BaseHPStart:   baseHPStart,
		BaseHPEnd:     g.Base.HP,
		LeaksPerWave:  leaksPerWave,
		KillsPerWave:  killsPerWave,
		MoneyEnd:      g.Money,
		TowersPlaced:  len(g.Towers),
		SimTimeS:      simTime,
		PathLengths:   pathLens,
	}, nil
}

// pathLengths reports each spawn's flow-field distance to base, computed
// once up front (before any bot action). Towers never block path tiles —
// only walls do, and no v1 bot builds walls — so these lengths stay valid
// for the whole run, matching how validate_map/mapcheck reports them.
func pathLengths(g *game.Game) map[string]int {
	out := map[string]int{}
	if g.Map == nil || g.FlowField == nil {
		return out
	}
	for _, spawn := range g.Map.Spawns {
		dist, _ := g.FlowField.At(spawn.X, spawn.Y)
		if dist >= flow.Inf {
			continue
		}
		out[spawn.ID] = int(dist)
	}
	return out
}
