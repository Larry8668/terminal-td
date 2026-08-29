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
	archiveName, err := AssetNameForCurrentPlatform(release.TagName)
	if err != nil {
		progress.Err = err
		return
	}
	archivePath := filepath.Join(updatesDir, "terminal-td-new-"+archiveName)
	err = DownloadArchiveWithProgress(release, archivePath, func(pct int) {
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
	if err := ExtractArchive(archivePath, extractDir); err != nil {
		progress.Err = err
		return
	}
	progress.Percent = 85

	newExePath, err := FindGameExeInDir(extractDir)
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
	if err := replaceBinary(currentExe, newExePath); err != nil {
		progress.Err = err
		return
	}
	progress.Percent = 100
}

// replaceBinary swaps newExePath's contents into currentExe's path.
//
// It renames the running binary aside rather than overwriting it in place.
// This is required on Windows -- the OS refuses to write over or delete the
// file backing a running process's executable image, but it does allow
// renaming it, since the process keeps its open handle regardless of what
// path now points at that file. A fresh file can then be created at the
// original path. The same rename-then-write sequence is used on every OS
// for one consistent, well-tested code path (it's also just as correct on
// macOS/Linux). The renamed-aside file (currentExe + oldBinarySuffix) is
// left behind for CleanupOldBinary to remove on the *next* startup, since
// removing it right now would race against this still-running process.
func replaceBinary(currentExe, newExePath string) error {
	data, err := os.ReadFile(newExePath)
	if err != nil {
		return fmt.Errorf("read new binary: %w", err)
	}

	oldExe := currentExe + oldBinarySuffix
	_ = os.Remove(oldExe) // drop any leftover from a previous, unfinished update
	if err := os.Rename(currentExe, oldExe); err != nil {
		return fmt.Errorf("move current binary aside: %w", err)
	}
	if err := os.WriteFile(currentExe, data, 0755); err != nil {
		_ = os.Rename(oldExe, currentExe) // best-effort restore
		return fmt.Errorf("write new binary: %w", err)
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(currentExe, 0755)
	}
	log.Printf("updater: replaced %s with new binary (old binary parked at %s, cleaned up on next launch)", currentExe, oldExe)
	return nil
}
