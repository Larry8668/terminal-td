#!/bin/bash

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o builds/terminal-td-windows.exe cmd/game/main.go

echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o builds/terminal-td-linux cmd/game/main.go

echo "Building for macOS Intel..."
GOOS=darwin GOARCH=amd64 go build -o builds/terminal-td-mac-intel cmd/game/main.go

echo "Building for macOS ARM..."
GOOS=darwin GOARCH=arm64 go build -o builds/terminal-td-mac-arm cmd/game/main.go

echo "Done! Check the builds/ directory"
