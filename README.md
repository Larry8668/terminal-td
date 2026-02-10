# Terminal Tower Defense 👾

A terminal-based tower defense game built with Go and tcell.

![Preview](images/preview.png)

> 🚧 **macOS:** Downloaded binaries may be blocked by Gatekeeper ("cannot be opened" / "not verified").  
> The app is unsigned (code signing requires a paid Apple Developer account). You can:
>
> 1. On Terminal: `xattr -d com.apple.quarantine terminal-td-v0.1.0-darwin-arm64` (use the actual path and filename).
> 2. Or **build from source:** `./build.sh mac-arm` then run the binary from the `builds/` folder.

## Quick Start ⚡

```bash
go run cmd/game/main.go
```

Or build and run:

```bash
go build -o terminal-td cmd/game/main.go
./terminal-td
```

## Controls 🎮

**Movement:**
- Arrow Keys or `WASD` - Move cursor

**Building:**
- `B` - Toggle build mode
- `SPACE/ENTER` - Place tower / Select tower
- `ESC` - Cancel / Deselect

**Gameplay:**
- `P` - Pause / Unpause
- `+/-` - Increase / Decrease game speed
- `R` - Restart (when game over)

**Quit:**
- `ESC` - Quit game (with confirmation)

## Features 🪄

- Tower placement and management
- Enemy waves with increasing difficulty
- Projectile-based combat system
- Real-time range visualization
- Economy system (earn money from kills)
- Wave progression system

## Requirements 📝

- Go 1.25+
- Terminal with UTF-8 support

## Building 🔧

Build for your platform:

```bash
go build -o terminal-td cmd/game/main.go
```

Cross-compile for other platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o terminal-td.exe cmd/game/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o terminal-td cmd/game/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o terminal-td cmd/game/main.go
```

Or use the [build script](build.sh) to build for all platforms.

---