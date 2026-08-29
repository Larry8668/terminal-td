package content

import (
	"os"
	"testing"

	mapdata "terminal-td/internal/map"
	waves "terminal-td/internal/waves"
)

// isolateConfigDir points the OS user config dir at a fresh temp dir for the
// duration of the test, so we never touch the real ~/Library/Application
// Support/terminal-td (or ~/.config on Linux) during tests.
func isolateConfigDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
}

func testMapDef(id string) *mapdata.MapDef {
	return &mapdata.MapDef{
		ID:   id,
		Name: "Test Map",
		Grid: mapdata.GridDef{Width: 10, Height: 5},
		Spawns: []mapdata.SpawnDef{
			{ID: "s1", X: 0, Y: 2},
		},
		Paths: []mapdata.PathDef{
			{SpawnID: "s1", Points: []mapdata.PointDef{{X: 0, Y: 2}, {X: 9, Y: 2}}},
		},
		Base: mapdata.BaseDef{X: 9, Y: 2, HP: 10},
	}
}

func TestSaveMapThenListAndLoad(t *testing.T) {
	isolateConfigDir(t)

	def := testMapDef("smoke-test-map")
	path, err := SaveMap(def, false)
	if err != nil {
		t.Fatalf("SaveMap: %v", err)
	}
	if path == "" {
		t.Fatal("SaveMap returned empty path")
	}

	maps, err := ListMaps()
	if err != nil {
		t.Fatalf("ListMaps: %v", err)
	}
	var found *mapdata.MapInfo
	for i := range maps {
		if maps[i].ID == def.ID {
			found = &maps[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("ListMaps did not include saved user map %q", def.ID)
	}
	if found.Source != mapdata.SourceUser {
		t.Fatalf("expected source %q, got %q", mapdata.SourceUser, found.Source)
	}

	loaded, err := LoadMapByID(def.ID)
	if err != nil {
		t.Fatalf("LoadMapByID: %v", err)
	}
	if loaded.ID != def.ID {
		t.Fatalf("loaded map id = %q, want %q", loaded.ID, def.ID)
	}
}

func TestSaveMapRejectsInvalidWithoutWriting(t *testing.T) {
	isolateConfigDir(t)

	bad := testMapDef("smoke-bad-map")
	bad.Base.HP = 0 // invalid: base hp must be positive

	if _, err := SaveMap(bad, false); err == nil {
		t.Fatal("expected SaveMap to reject an invalid map, got nil error")
	}

	dir, err := MapsDir()
	if err != nil {
		t.Fatalf("MapsDir: %v", err)
	}
	entries, err := readDirNames(dir)
	if err != nil {
		t.Fatalf("read maps dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no files written for invalid map, found %v", entries)
	}
}

func TestSaveMapRefusesOverwriteUnlessRequested(t *testing.T) {
	isolateConfigDir(t)

	def := testMapDef("smoke-overwrite-map")
	if _, err := SaveMap(def, false); err != nil {
		t.Fatalf("initial SaveMap: %v", err)
	}
	if _, err := SaveMap(def, false); err == nil {
		t.Fatal("expected second SaveMap without overwrite to fail")
	}
	if _, err := SaveMap(def, true); err != nil {
		t.Fatalf("SaveMap with overwrite=true: %v", err)
	}
}

func TestSaveWavesValidatesAgainstMapSpawns(t *testing.T) {
	isolateConfigDir(t)

	def := testMapDef("smoke-waves-map")
	if _, err := SaveMap(def, false); err != nil {
		t.Fatalf("SaveMap: %v", err)
	}

	goodWaves := []waves.WaveDef{
		{Wave: 1, Groups: []waves.SpawnGroupDef{
			{SpawnID: "s1", EnemyType: "basic", Count: 3, Interval: 1.0},
		}},
	}
	path, err := SaveWaves(def.ID, goodWaves)
	if err != nil {
		t.Fatalf("SaveWaves: %v", err)
	}
	if path == "" {
		t.Fatal("SaveWaves returned empty path")
	}

	loaded, err := LoadWavesForMap(def.ID)
	if err != nil {
		t.Fatalf("LoadWavesForMap: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Groups[0].SpawnID != "s1" {
		t.Fatalf("unexpected loaded waves: %+v", loaded)
	}

	badWaves := []waves.WaveDef{
		{Wave: 1, Groups: []waves.SpawnGroupDef{
			{SpawnID: "does-not-exist", EnemyType: "basic", Count: 3, Interval: 1.0},
		}},
	}
	if _, err := SaveWaves(def.ID, badWaves); err == nil {
		t.Fatal("expected SaveWaves to reject a spawn_id not present on the map")
	}
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
