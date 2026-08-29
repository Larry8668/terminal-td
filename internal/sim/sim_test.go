package sim

import (
	"reflect"
	"testing"

	"terminal-td/internal/content"
	mapdata "terminal-td/internal/map"
	"terminal-td/internal/waves"
)

// isolateConfigDir points the OS user config dir at a fresh temp dir for the
// duration of the test, so content.LoadMapByID never touches the real
// config dir (mirrors internal/content and internal/mcpserver's own helper
// of the same name).
func isolateConfigDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
}

func loadForks(t *testing.T) *mapdata.GameMap {
	t.Helper()
	m, err := content.LoadMapByID("forks")
	if err != nil {
		t.Fatalf("load builtin map %q: %v", "forks", err)
	}
	return m
}

func TestRunRequiresMap(t *testing.T) {
	if _, err := Run(Config{}); err == nil {
		t.Fatal("expected an error when Config.Map is nil")
	}
}

func TestRunNoneBotLoses(t *testing.T) {
	isolateConfigDir(t)
	m := loadForks(t)

	res, err := Run(Config{Map: m, Bot: NoneBot{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome != "lost" {
		t.Fatalf("expected a map with real waves and zero defense to lose, got outcome=%q: %+v", res.Outcome, res)
	}
	if res.TowersPlaced != 0 {
		t.Fatalf("expected NoneBot to place zero towers, got %d", res.TowersPlaced)
	}
	if res.BaseHPEnd != 0 {
		t.Fatalf("expected base_hp_end=0 on a loss, got %d", res.BaseHPEnd)
	}
	if res.WavesSurvived >= res.WavesTotal {
		t.Fatalf("expected fewer waves survived than total on a loss: survived=%d total=%d", res.WavesSurvived, res.WavesTotal)
	}
	if len(res.PathLengths) == 0 {
		t.Fatal("expected non-empty path_lengths")
	}
	total := 0
	for _, n := range res.LeaksPerWave {
		total += n
	}
	if total == 0 {
		t.Fatal("expected at least one leak with zero defense")
	}
}

func TestRunGreedyBotOutperformsNoneBot(t *testing.T) {
	isolateConfigDir(t)
	m := loadForks(t)

	none, err := Run(Config{Map: m, Bot: NoneBot{}})
	if err != nil {
		t.Fatalf("Run(none): %v", err)
	}
	greedy, err := Run(Config{Map: m, Bot: GreedyBot{}})
	if err != nil {
		t.Fatalf("Run(greedy): %v", err)
	}

	if greedy.WavesSurvived < none.WavesSurvived {
		t.Fatalf("expected greedy to survive at least as many waves as none: none=%d greedy=%d", none.WavesSurvived, greedy.WavesSurvived)
	}
	if greedy.TowersPlaced == 0 {
		t.Fatal("expected GreedyBot to place at least one tower")
	}
	if greedy.MoneyEnd >= DefaultBudget && greedy.TowersPlaced > 0 {
		t.Fatalf("expected greedy to have spent some of its starting budget, money_end=%d towers_placed=%d", greedy.MoneyEnd, greedy.TowersPlaced)
	}
}

func TestRunDeterministic(t *testing.T) {
	isolateConfigDir(t)
	m := loadForks(t)

	first, err := Run(Config{Map: m, Bot: GreedyBot{}})
	if err != nil {
		t.Fatalf("Run (first): %v", err)
	}
	second, err := Run(Config{Map: m, Bot: GreedyBot{}})
	if err != nil {
		t.Fatalf("Run (second): %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected identical results for identical inputs:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

func TestRunBudgetOverrideAffectsTowerCount(t *testing.T) {
	isolateConfigDir(t)
	m := loadForks(t)

	low, err := Run(Config{Map: m, Bot: GreedyBot{}, Budget: 100})
	if err != nil {
		t.Fatalf("Run(budget=100): %v", err)
	}
	high, err := Run(Config{Map: m, Bot: GreedyBot{}, Budget: 1000})
	if err != nil {
		t.Fatalf("Run(budget=1000): %v", err)
	}
	if low.TowersPlaced >= high.TowersPlaced {
		t.Fatalf("expected a bigger starting budget to place more towers: budget=100 -> %d, budget=1000 -> %d", low.TowersPlaced, high.TowersPlaced)
	}
}

func TestRunTimeoutWhenCapReachedBeforeResolution(t *testing.T) {
	isolateConfigDir(t)
	m := loadForks(t)

	res, err := Run(Config{Map: m, Bot: NoneBot{}, MaxSimTime: 1.0})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Outcome != "timeout" {
		t.Fatalf("expected outcome=timeout when the map hasn't resolved by MaxSimTime, got %q", res.Outcome)
	}
	if res.SimTimeS < 1.0 {
		t.Fatalf("expected sim_time_s >= MaxSimTime, got %f", res.SimTimeS)
	}
}

func TestRunRejectsWavesOverrideWithUnknownSpawnID(t *testing.T) {
	isolateConfigDir(t)
	m := loadForks(t)

	badWaves := []waves.WaveDef{
		{Wave: 1, Groups: []waves.SpawnGroupDef{
			{SpawnID: "does-not-exist", EnemyType: "basic", Count: 1, Interval: 1},
		}},
	}
	if _, err := Run(Config{Map: m, Waves: badWaves, Bot: NoneBot{}}); err == nil {
		t.Fatal("expected an error for a waves override referencing an unknown spawn id")
	}
}

func TestNewBotRejectsUnknownName(t *testing.T) {
	if _, err := NewBot("totally-not-a-bot"); err == nil {
		t.Fatal("expected NewBot to reject an unrecognized strategy name")
	}
}

func TestNewBotKnownNames(t *testing.T) {
	for _, name := range []string{"", "none", "greedy"} {
		if _, err := NewBot(name); err != nil {
			t.Fatalf("NewBot(%q): unexpected error: %v", name, err)
		}
	}
}
