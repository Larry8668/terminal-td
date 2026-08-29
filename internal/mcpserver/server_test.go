package mcpserver

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"terminal-td/internal/buildinfo"
	"terminal-td/internal/game"
	mapdata "terminal-td/internal/map"
	"terminal-td/internal/sim"
	waves "terminal-td/internal/waves"
)

// isolateConfigDir points the OS user config dir at a fresh temp dir for the
// duration of the test, mirroring internal/content's own test helper, so
// create_map/create_waves/delete_map never touch the real config dir.
func isolateConfigDir(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
}

// connectTestClient wires a real client to a real server over an in-memory
// transport pair — the same connection mechanics a stdio client/server use,
// minus the process boundary. This exercises the actual MCP protocol
// (schema validation, JSON-RPC framing, structured content) rather than
// calling the tool handler Go functions directly.
func connectTestClient(t *testing.T) *mcp.ClientSession {
	t.Helper()
	server := NewServer()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

// callTool invokes a tool and decodes its structured content into Out. It
// never fails the test on a tool-level error (IsError=true) — callers that
// expect success should assert !res.IsError themselves; this lets tests also
// cover the expected-failure paths (e.g. an invalid create_map).
func callTool[Out any](t *testing.T, session *mcp.ClientSession, name string, args any) (Out, *mcp.CallToolResult) {
	t.Helper()
	var out Out
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s): protocol-level error: %v", name, err)
	}
	if res.StructuredContent != nil {
		raw, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatalf("CallTool(%s): remarshal structured content: %v", name, err)
		}
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("CallTool(%s): unmarshal structured content into %T: %v", name, out, err)
		}
	}
	return out, res
}

func testMapDef(id string) mapdata.MapDef {
	return mapdata.MapDef{
		ID:   id,
		Name: "MCP Test Map",
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

func TestGetGameRules(t *testing.T) {
	session := connectTestClient(t)
	out, res := callTool[GetGameRulesOutput](t, session, "get_game_rules", GetGameRulesInput{})
	if res.IsError {
		t.Fatalf("get_game_rules returned an error result: %+v", res.Content)
	}
	if out.GameVersion != buildinfo.Version {
		t.Fatalf("game_version = %q, want %q", out.GameVersion, buildinfo.Version)
	}
	if len(out.Towers) == 0 {
		t.Fatal("expected at least one tower rule")
	}
	if len(out.EnemyTypes) == 0 {
		t.Fatal("expected at least one enemy type")
	}
	if out.Walls.LinkRadiusManhattan != game.MaxWallLinkDist {
		t.Fatalf("wall link radius = %d, want %d", out.Walls.LinkRadiusManhattan, game.MaxWallLinkDist)
	}
	if out.MapSchema.Example == nil || out.WaveSchema.Example == nil {
		t.Fatal("expected non-nil map/wave schema examples")
	}
}

func TestValidateMapAcceptsAndRejects(t *testing.T) {
	session := connectTestClient(t)

	good, res := callTool[map[string]any](t, session, "validate_map", ValidateMapInput{MapDef: testMapDef("validate-good")})
	if res.IsError {
		t.Fatalf("validate_map (good) returned a tool error: %+v", res.Content)
	}
	if valid, _ := good["valid"].(bool); !valid {
		t.Fatalf("expected valid=true, got %+v", good)
	}

	badDef := testMapDef("validate-bad")
	badDef.Base.HP = 0
	bad, res := callTool[map[string]any](t, session, "validate_map", ValidateMapInput{MapDef: badDef})
	if res.IsError {
		t.Fatalf("validate_map (bad) should report invalid via its output, not a tool error: %+v", res.Content)
	}
	if valid, _ := bad["valid"].(bool); valid {
		t.Fatalf("expected valid=false for a map with base hp <= 0, got %+v", bad)
	}
}

func TestFullMapLifecycleOverMCP(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)
	def := testMapDef("mcp-lifecycle")

	createOut, res := callTool[CreateMapOutput](t, session, "create_map", CreateMapInput{MapDef: def})
	if res.IsError {
		t.Fatalf("create_map: %+v", res.Content)
	}
	if createOut.ID != def.ID || createOut.SavedPath == "" {
		t.Fatalf("unexpected create_map output: %+v", createOut)
	}

	listOut, res := callTool[ListMapsOutput](t, session, "list_maps", ListMapsInput{})
	if res.IsError {
		t.Fatalf("list_maps: %+v", res.Content)
	}
	found := false
	for _, m := range listOut.Maps {
		if m.ID == def.ID {
			found = true
			if m.Source != mapdata.SourceUser {
				t.Fatalf("expected source %q, got %q", mapdata.SourceUser, m.Source)
			}
			if m.SpawnCount != 1 {
				t.Fatalf("expected spawn_count 1, got %d", m.SpawnCount)
			}
		}
	}
	if !found {
		t.Fatalf("list_maps did not include %q", def.ID)
	}

	getOut, res := callTool[GetMapOutput](t, session, "get_map", GetMapInput{ID: def.ID})
	if res.IsError {
		t.Fatalf("get_map: %+v", res.Content)
	}
	if getOut.Map.ID != def.ID || getOut.Source != mapdata.SourceUser {
		t.Fatalf("unexpected get_map output: %+v", getOut)
	}
	if len(getOut.Waves) != 0 {
		t.Fatalf("expected no waves before create_waves, got %d", len(getOut.Waves))
	}

	wavesIn := []waves.WaveDef{
		{Wave: 1, Groups: []waves.SpawnGroupDef{
			{SpawnID: "s1", EnemyType: "basic", Count: 3, Interval: 1.0},
		}},
	}
	wavesOut, res := callTool[CreateWavesOutput](t, session, "create_waves", CreateWavesInput{MapID: def.ID, Waves: wavesIn})
	if res.IsError {
		t.Fatalf("create_waves: %+v", res.Content)
	}
	if wavesOut.SavedPath == "" {
		t.Fatal("expected non-empty saved_path from create_waves")
	}

	getOut2, res := callTool[GetMapOutput](t, session, "get_map", GetMapInput{ID: def.ID})
	if res.IsError {
		t.Fatalf("get_map (after waves): %+v", res.Content)
	}
	if len(getOut2.Waves) != 1 {
		t.Fatalf("expected 1 wave after create_waves, got %d", len(getOut2.Waves))
	}

	deleteOut, res := callTool[DeleteMapOutput](t, session, "delete_map", DeleteMapInput{ID: def.ID})
	if res.IsError {
		t.Fatalf("delete_map: %+v", res.Content)
	}
	if !deleteOut.Deleted {
		t.Fatal("expected deleted=true")
	}

	_, res = callTool[GetMapOutput](t, session, "get_map", GetMapInput{ID: def.ID})
	if !res.IsError {
		t.Fatal("expected get_map to error for a deleted map id")
	}
}

func TestCreateMapRejectsInvalidWithoutWriting(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)

	badDef := testMapDef("create-invalid")
	badDef.Base.HP = 0
	_, res := callTool[CreateMapOutput](t, session, "create_map", CreateMapInput{MapDef: badDef})
	if !res.IsError {
		t.Fatal("expected create_map to report a tool error for an invalid map")
	}
}

func TestDeleteMapRefusesBuiltinOverMCP(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)

	_, res := callTool[DeleteMapOutput](t, session, "delete_map", DeleteMapInput{ID: "classic"})
	if !res.IsError {
		t.Fatal("expected delete_map to refuse deleting a built-in map")
	}
}

func TestSimulateRunNoneBotLoses(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)

	out, res := callTool[sim.Result](t, session, "simulate_run", SimulateRunInput{MapID: "forks", Bot: "none"})
	if res.IsError {
		t.Fatalf("simulate_run: %+v", res.Content)
	}
	if out.Outcome != "lost" {
		t.Fatalf("expected outcome=lost with bot=none (zero defense), got %q: %+v", out.Outcome, out)
	}
	if out.WavesTotal == 0 {
		t.Fatal("expected a non-zero waves_total")
	}
	if len(out.PathLengths) == 0 {
		t.Fatal("expected non-empty path_lengths")
	}
}

func TestSimulateRunGreedyBotBeatsNoneBot(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)

	none, res := callTool[sim.Result](t, session, "simulate_run", SimulateRunInput{MapID: "forks", Bot: "none"})
	if res.IsError {
		t.Fatalf("simulate_run(none): %+v", res.Content)
	}
	greedy, res := callTool[sim.Result](t, session, "simulate_run", SimulateRunInput{MapID: "forks", Bot: "greedy"})
	if res.IsError {
		t.Fatalf("simulate_run(greedy): %+v", res.Content)
	}
	if greedy.WavesSurvived < none.WavesSurvived {
		t.Fatalf("expected greedy bot to survive at least as many waves as none bot; none=%d greedy=%d", none.WavesSurvived, greedy.WavesSurvived)
	}
	if greedy.TowersPlaced == 0 {
		t.Fatal("expected greedy bot to place at least one tower")
	}
}

func TestSimulateRunDeterministic(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)

	first, res := callTool[sim.Result](t, session, "simulate_run", SimulateRunInput{MapID: "forks", Bot: "greedy"})
	if res.IsError {
		t.Fatalf("simulate_run (first): %+v", res.Content)
	}
	second, res := callTool[sim.Result](t, session, "simulate_run", SimulateRunInput{MapID: "forks", Bot: "greedy"})
	if res.IsError {
		t.Fatalf("simulate_run (second): %+v", res.Content)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected identical results for identical inputs, got:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

func TestSimulateRunRejectsBadBotName(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)

	_, res := callTool[sim.Result](t, session, "simulate_run", SimulateRunInput{MapID: "forks", Bot: "nonsense"})
	if !res.IsError {
		t.Fatal("expected simulate_run to reject an unknown bot strategy")
	}
}

func TestSimulateRunRejectsUnknownMap(t *testing.T) {
	isolateConfigDir(t)
	session := connectTestClient(t)

	_, res := callTool[sim.Result](t, session, "simulate_run", SimulateRunInput{MapID: "does-not-exist", Bot: "none"})
	if !res.IsError {
		t.Fatal("expected simulate_run to reject an unknown map id")
	}
}
