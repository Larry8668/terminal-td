package mapcheck

import (
	"testing"

	mapdata "terminal-td/internal/map"
)

func validDef() *mapdata.MapDef {
	return &mapdata.MapDef{
		ID:   "check-me",
		Name: "Check Me",
		Grid: mapdata.GridDef{Width: 20, Height: 10},
		Spawns: []mapdata.SpawnDef{
			{ID: "s1", X: 0, Y: 5},
		},
		Paths: []mapdata.PathDef{
			{SpawnID: "s1", Points: []mapdata.PointDef{{X: 0, Y: 5}, {X: 19, Y: 5}}},
		},
		Base: mapdata.BaseDef{X: 19, Y: 5, HP: 10},
	}
}

func TestValidateMapAccepted(t *testing.T) {
	res := ValidateMap(validDef())
	if !res.Valid {
		t.Fatalf("expected valid map, got errors: %v", res.Errors)
	}
	if got := res.PathLengths["s1"]; got != 19 {
		t.Fatalf("path length for s1 = %d, want 19", got)
	}
	if len(res.Warnings) != 0 {
		t.Fatalf("expected no warnings for a 19-tile lane, got %v", res.Warnings)
	}
}

func TestValidateMapStructuralError(t *testing.T) {
	def := validDef()
	def.Base.HP = 0 // invalid per the loader
	res := ValidateMap(def)
	if res.Valid {
		t.Fatal("expected invalid result for base hp <= 0")
	}
	if len(res.Errors) == 0 {
		t.Fatal("expected at least one error message")
	}
}

func TestValidateMapDisconnectedSpawn(t *testing.T) {
	def := &mapdata.MapDef{
		ID:   "disconnected",
		Name: "Disconnected",
		Grid: mapdata.GridDef{Width: 10, Height: 5},
		Spawns: []mapdata.SpawnDef{
			{ID: "reachable", X: 0, Y: 2},
			{ID: "island", X: 9, Y: 0}, // never connected by any path
		},
		Paths: []mapdata.PathDef{
			{SpawnID: "reachable", Points: []mapdata.PointDef{{X: 0, Y: 2}, {X: 5, Y: 2}}},
		},
		Base: mapdata.BaseDef{X: 5, Y: 2, HP: 10},
	}
	res := ValidateMap(def)
	if res.Valid {
		t.Fatal("expected invalid result: 'island' spawn has no path tiles at all")
	}
	found := false
	for _, e := range res.Errors {
		if e != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected a non-empty error describing the disconnected spawn")
	}
}

func TestValidateMapShortLaneWarning(t *testing.T) {
	def := &mapdata.MapDef{
		ID:   "short-lane",
		Name: "Short Lane",
		Grid: mapdata.GridDef{Width: 10, Height: 5},
		Spawns: []mapdata.SpawnDef{
			{ID: "s1", X: 0, Y: 2},
		},
		Paths: []mapdata.PathDef{
			{SpawnID: "s1", Points: []mapdata.PointDef{{X: 0, Y: 2}, {X: 2, Y: 2}}},
		},
		Base: mapdata.BaseDef{X: 2, Y: 2, HP: 10},
	}
	res := ValidateMap(def)
	if !res.Valid {
		t.Fatalf("expected valid map, got errors: %v", res.Errors)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("expected exactly one short-lane warning, got %v", res.Warnings)
	}
}
