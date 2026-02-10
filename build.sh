#!/bin/bash

# Read version from internal/game/version.go (const Version = "x.y.z")
VERSION=$(grep 'Version = ' internal/game/version.go | cut -d'"' -f2)
if [ -z "$VERSION" ]; then
	echo "Could not read version from internal/game/version.go"
	exit 1
fi

BUILD_DIR="builds"
PLATFORM="${1:-all}"

mkdir -p "$BUILD_DIR"

write_readme() {
	local dir="$1"
	cat > "$dir/readme.txt" << 'EOF'
TERMINAL TOWER DEFENSE

Files in this folder:
- The game binary (terminal-td-*): run this to play.
- terminal-td-updater: helper used when you install an update from inside the game. Do not delete; keep it next to the game.

Mac users: macOS may block the app because it is unsigned. If you see "cannot be opened" / "not verified", you can follow what is mentioned in the README.md file of repo - https://github.com/Larry8668/terminal-td/blob/main/README.md.
EOF
}

build_windows() {
	local dir="$BUILD_DIR/terminal-td-${VERSION}-windows-amd64"
	mkdir -p "$dir"
	echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o "$dir/terminal-td-windows.exe" cmd/game/main.go
	GOOS=windows GOARCH=amd64 go build -o "$dir/terminal-td-updater.exe" cmd/updater/main.go
	write_readme "$dir"
}

build_linux() {
	local dir="$BUILD_DIR/terminal-td-${VERSION}-linux-amd64"
	mkdir -p "$dir"
	echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o "$dir/terminal-td-linux" cmd/game/main.go
	GOOS=linux GOARCH=amd64 go build -o "$dir/terminal-td-updater" cmd/updater/main.go
	write_readme "$dir"
}

build_mac_intel() {
	local dir="$BUILD_DIR/terminal-td-${VERSION}-darwin-amd64"
	mkdir -p "$dir"
	echo "Building for macOS Intel (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o "$dir/terminal-td-mac-intel" cmd/game/main.go
	GOOS=darwin GOARCH=amd64 go build -o "$dir/terminal-td-updater" cmd/updater/main.go
	write_readme "$dir"
}

build_mac_arm() {
	local dir="$BUILD_DIR/terminal-td-${VERSION}-darwin-arm64"
	mkdir -p "$dir"
	echo "Building for macOS ARM (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o "$dir/terminal-td-mac-arm" cmd/game/main.go
	GOOS=darwin GOARCH=arm64 go build -o "$dir/terminal-td-updater" cmd/updater/main.go
	write_readme "$dir"
}

build_all() {
	build_windows
	build_linux
	build_mac_intel
	build_mac_arm
}

case "$PLATFORM" in
	windows|win)
		build_windows
		;;
	linux)
		build_linux
		;;
	mac-intel|darwin-amd64)
		build_mac_intel
		;;
	mac-arm|darwin-arm64|mac)
		build_mac_arm
		;;
	all|"")
		build_all
		;;
	-h|--help)
		echo "Usage: $0 [platform]"
		echo ""
		echo "Platforms:"
		echo "  all (default)  - Build for Windows, Linux, macOS Intel, macOS ARM"
		echo "  windows, win  - Windows amd64"
		echo "  linux         - Linux amd64"
		echo "  mac-intel     - macOS Intel (amd64)"
		echo "  mac-arm, mac  - macOS Apple Silicon (arm64)"
		echo ""
		exit 0
		;;
	*)
		echo "Unknown platform: $PLATFORM"
		echo "Run '$0 --help' for usage"
		exit 1
		;;
esac

echo ""
echo "Done! Versioned folders in $BUILD_DIR/"
for d in "$BUILD_DIR"/*/; do [ -d "$d" ] && echo "  $d" && ls -lh "$d" 2>/dev/null; done