package content

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	mapdata "terminal-td/internal/map"
)

// userMapFile pairs a decoded user map with the file it was loaded from.
type userMapFile struct {
	path string
	m    *mapdata.GameMap
}

// scanUserMaps reads every *.json file in the user maps dir and decodes it as
// a map. Invalid files are logged and skipped rather than failing the whole
// scan, so one bad file doesn't hide the rest of the user's maps.
func scanUserMaps() ([]userMapFile, error) {
	dir, err := MapsDir()
	if err != nil {
		return nil, fmt.Errorf("user maps dir: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read user maps dir: %w", err)
	}
	var out []userMapFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("content: read user map %s: %v", e.Name(), err)
			continue
		}
		m, err := mapdata.LoadMapBytes(data)
		if err != nil {
			log.Printf("content: invalid user map %s: %v", e.Name(), err)
			continue
		}
		out = append(out, userMapFile{path: path, m: m})
	}
	return out, nil
}

// ListMaps returns built-in maps merged with user maps from the config dir.
// A user map whose id matches a built-in map shadows it (logged, per
// DECISIONS.md); it does not appear twice.
func ListMaps() ([]mapdata.MapInfo, error) {
	builtin, err := mapdata.ListMaps()
	if err != nil {
		return nil, err
	}

	result := make([]mapdata.MapInfo, len(builtin))
	copy(result, builtin)
	indexByID := make(map[string]int, len(result))
	for i, m := range result {
		indexByID[m.ID] = i
	}

	userMaps, err := scanUserMaps()
	if err != nil {
		log.Printf("content: %v (continuing with built-in maps only)", err)
		return result, nil
	}
	for _, um := range userMaps {
		info := mapdata.InfoFromMap(um.m, mapdata.SourceUser)
		if idx, ok := indexByID[um.m.ID]; ok {
			log.Printf("content: user map %q (%s) shadows built-in map with the same id", um.m.ID, um.path)
			result[idx] = info
		} else {
			indexByID[um.m.ID] = len(result)
			result = append(result, info)
		}
	}
	return result, nil
}

// LoadMapByID returns the map for id, preferring a user override over the
// built-in embedded map of the same id.
func LoadMapByID(id string) (*mapdata.GameMap, error) {
	if userMaps, err := scanUserMaps(); err != nil {
		log.Printf("content: %v (falling back to built-in maps)", err)
	} else {
		for _, um := range userMaps {
			if um.m.ID == id {
				log.Printf("content: loaded user map %q from %s", id, um.path)
				return um.m, nil
			}
		}
	}
	return mapdata.LoadMapByID(id)
}

// LoadMapDefByID returns the raw MapDef (preserving every authored path
// entry, including forks) for id, preferring a user override over the
// built-in embedded map of the same id. Use this instead of LoadMapByID when
// the caller needs to hand back the map's original JSON shape, e.g. the MCP
// get_map tool.
func LoadMapDefByID(id string) (*mapdata.MapDef, error) {
	dir, err := MapsDir()
	if err != nil {
		log.Printf("content: %v (falling back to built-in maps)", err)
		return mapdata.LoadMapDefByID(id)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("content: read user maps dir: %v (falling back to built-in maps)", err)
		return mapdata.LoadMapDefByID(id)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		def, err := mapdata.LoadMapDefBytes(data)
		if err != nil {
			continue
		}
		if def.ID == id {
			log.Printf("content: loaded user map def %q from %s", id, path)
			return def, nil
		}
	}
	return mapdata.LoadMapDefByID(id)
}

// SaveMap validates def (by round-tripping it through the same loader used
// for embedded/user maps) and writes it to "<maps dir>/<id>.json". It refuses
// to overwrite an existing user map unless overwrite is true. Returns the
// path written.
func SaveMap(def *mapdata.MapDef, overwrite bool) (string, error) {
	if def.ID == "" {
		return "", fmt.Errorf("map id is required")
	}
	if _, err := mapdata.LoadMapDef(def); err != nil {
		return "", fmt.Errorf("map %q failed validation: %w", def.ID, err)
	}
	data, err := json.MarshalIndent(def, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal map %q: %w", def.ID, err)
	}

	dir, err := MapsDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, def.ID+".json")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return "", fmt.Errorf("user map %q already exists (pass overwrite=true to replace it)", def.ID)
		}
	}

	if builtin, err := mapdata.LoadMapByID(def.ID); err == nil && builtin != nil {
		log.Printf("content: user map %q will shadow the built-in map of the same id", def.ID)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write map %q: %w", def.ID, err)
	}
	log.Printf("content: saved user map %q to %s", def.ID, path)
	return path, nil
}

// DeleteMap removes the user map with the given id, plus its waves override
// file if one exists. It never touches built-in maps: if id only exists as a
// built-in map (or doesn't exist at all), this returns an error.
func DeleteMap(id string) error {
	userMaps, err := scanUserMaps()
	if err != nil {
		return fmt.Errorf("user maps dir: %w", err)
	}
	for _, um := range userMaps {
		if um.m.ID != id {
			continue
		}
		if err := os.Remove(um.path); err != nil {
			return fmt.Errorf("delete map %q: %w", id, err)
		}
		log.Printf("content: deleted user map %q (%s)", id, um.path)

		if wavesPath, err := wavesPathForMap(id); err == nil {
			if _, statErr := os.Stat(wavesPath); statErr == nil {
				if rmErr := os.Remove(wavesPath); rmErr != nil {
					log.Printf("content: deleted map %q but failed to remove its waves override %s: %v", id, wavesPath, rmErr)
				} else {
					log.Printf("content: deleted waves override for map %q (%s)", id, wavesPath)
				}
			}
		}
		return nil
	}
	return fmt.Errorf("user map %q not found (built-in maps can't be deleted)", id)
}
