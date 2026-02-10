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

	zipPath := filepath.Join(updatesDir, "terminal-td-new.zip")
	if err := DownloadZip(release, zipPath); err != nil {
		return err
	}

	extractDir := filepath.Join(updatesDir, "extract")
	if err := os.RemoveAll(extractDir); err != nil {
		return fmt.Errorf("clear extract dir: %w", err)
	}
	if err := ExtractZip(zipPath, extractDir); err != nil {
		return err
	}

	newExePath, err := FindGameExeInDir(extractDir)
	if err != nil {
		return err
	}

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("resolve exe: %w", err)
	}
	currentDir := filepath.Dir(currentExe)
	updaterPath := filepath.Join(currentDir, updaterName())

	newUpdaterPath := filepath.Join(filepath.Dir(newExePath), updaterName())
	if data, err := os.ReadFile(newUpdaterPath); err == nil {
		dest := filepath.Join(currentDir, updaterName())
		if err := os.WriteFile(dest, data, 0755); err != nil {
			log.Printf("updater: copy new updater binary: %v", err)
		}
	}

	changelogPath, err := WriteChangelogFile(release.Body)
	if err != nil {
		log.Printf("updater: write changelog: %v", err)
		changelogPath = ""
	}

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
