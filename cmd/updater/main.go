package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	currentPath := flag.String("current", "", "Path to current game executable (will be replaced)")
	newPath := flag.String("new", "", "Path to new game executable (downloaded)")
	changelogPath := flag.String("changelog", "", "Path to changelog file to pass to new instance")
	flag.Parse()

	if *currentPath == "" || *newPath == "" {
		log.Fatal("usage: updater -current <path> -new <path> [-changelog <path>]")
	}

	// Allow the game process to exit and release file locks
	time.Sleep(500 * time.Millisecond)

	// Copy new binary over current (replace)
	if err := copyFile(*newPath, *currentPath); err != nil {
		log.Fatalf("replace executable: %v", err)
	}

	// Build args for new process: --just-updated and optionally --changelog
	args := []string{"--just-updated"}
	if *changelogPath != "" {
		args = append(args, "--changelog", *changelogPath)
	}

	cmd := exec.Command(*currentPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("start new process: %v", err)
	}
	os.Exit(0)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read new binary: %w", err)
	}
	if err := os.WriteFile(dst, data, 0755); err != nil {
		return fmt.Errorf("write over current: %w", err)
	}
	_ = os.Remove(src)
	return nil
}
