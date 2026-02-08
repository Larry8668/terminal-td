package game

import "time"

type WaveManager struct {
	CurrentWave int
	TotalWaves  int

	EnemiesPerWave int
	EnemiesSpawned int
	EnemiesAlive   int

	SpawnInterval time.Duration
	SpawnTimer    time.Duration

	WaveCooldown time.Duration

	Spawning      bool
	SpawnFinished bool
}
