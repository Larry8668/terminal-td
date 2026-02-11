package updater

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"terminal-td/internal/config"
)

func RunUpdateWithProgress(release *Release, progress *Progress) {
	defer func() {
		progress.Done = true
		if progress.Err != nil {
			log.Printf("updater: failed: %v", progress.Err)
		}
	}()

	progress.Step = "Pinging server..."
	progress.Percent = 0
	updatesDir, err := config.UpdatesPath()
	if err != nil {
		progress.Err = err
		return
	}

	progress.Step = "Downloading..."
	progress.Percent = 5
	zipPath := filepath.Join(updatesDir, "terminal-td-new.zip")
	err = DownloadZipWithProgress(release, zipPath, func(pct int) {
		progress.Percent = 5 + (pct*70)/100
	})
	if err != nil {
		progress.Err = err
		return
	}

	progress.Step = "Extracting..."
	progress.Percent = 75
	extractDir := filepath.Join(updatesDir, "extract")
	if err := os.RemoveAll(extractDir); err != nil {
		progress.Err = fmt.Errorf("clear extract dir: %w", err)
		return
	}
	if err := ExtractZip(zipPath, extractDir); err != nil {
		progress.Err = err
		return
	}
	progress.Percent = 85

	newExePath, err := FindGameExeInDir(extractDir, release.TagName)
	if err != nil {
		progress.Err = err
		return
	}

	progress.Step = "Replacing..."
	progress.Percent = 90
	currentExe, err := os.Executable()
	if err != nil {
		progress.Err = fmt.Errorf("executable path: %w", err)
		return
	}
	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		progress.Err = fmt.Errorf("resolve exe: %w", err)
		return
	}
	data, err := os.ReadFile(newExePath)
	if err != nil {
		progress.Err = fmt.Errorf("read new binary: %w", err)
		return
	}
	if err := os.WriteFile(currentExe, data, 0755); err != nil {
		progress.Err = fmt.Errorf("replace exe (close and run again if on Windows): %w", err)
		return
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(currentExe, 0755)
	}
	log.Printf("updater: replaced %s with new binary", currentExe)
	progress.Percent = 100
}
