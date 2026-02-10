package updater

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"terminal-td/internal/config"
)

const (
	UpdaterNameUnix    = "terminal-td-updater"
	UpdaterNameWindows = "terminal-td-updater.exe"
)

func RunUpdate(release *Release) error {
	updatesDir, err := config.UpdatesPath()
	if err != nil {
		return err
	}

	var newExeName string
	if runtime.GOOS == "windows" {
		newExeName = "terminal-td-new.exe"
	} else {
		newExeName = "terminal-td-new"
	}
	newExePath := filepath.Join(updatesDir, newExeName)

	if err := DownloadUpdate(release, newExePath); err != nil {
		return err
	}

	changelogPath, err := WriteChangelogFile(release.Body)
	if err != nil {
		log.Printf("updater: write changelog: %v", err)
		changelogPath = ""
	}

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("resolve exe: %w", err)
	}

	updaterPath := filepath.Join(filepath.Dir(currentExe), updaterName())
	cmd := exec.Command(updaterPath,
		"-current", currentExe,
		"-new", newExePath,
		"-changelog", changelogPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start updater: %w", err)
	}
	os.Exit(0)
	return nil
}

func updaterName() string {
	if runtime.GOOS == "windows" {
		return UpdaterNameWindows
	}
	return UpdaterNameUnix
}
