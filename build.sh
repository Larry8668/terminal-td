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

build_windows() {
	echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/terminal-td-${VERSION}-windows-amd64.exe" cmd/game/main.go
}

build_linux() {
	echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/terminal-td-${VERSION}-linux-amd64" cmd/game/main.go
}

build_mac_intel() {
	echo "Building for macOS Intel (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/terminal-td-${VERSION}-darwin-amd64" cmd/game/main.go
}

build_mac_arm() {
	echo "Building for macOS ARM (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o "$BUILD_DIR/terminal-td-${VERSION}-darwin-arm64" cmd/game/main.go
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
echo "Done! Check the $BUILD_DIR/ directory"
ls -lh "$BUILD_DIR/" 2>/dev/null || true
