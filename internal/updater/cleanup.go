package updater

import (
	"log"
	"os"
	"path/filepath"
)

// oldBinarySuffix is appended to the running executable's path during an
// update to park the previous binary out of the way (see RunUpdateWithProgress).
const oldBinarySuffix = ".old"

// CleanupOldBinary removes a leftover "<exe>.old" file from a previous
// update, if one exists. It must be called from a *freshly started*
// process (e.g. once near the top of main), never from the process that
// just performed the rename -- that process still holds the old file open
// on some platforms, and deleting it out from under itself is exactly what
// the rename-then-write dance in RunUpdateWithProgress avoids doing.
//
// This is safe to call unconditionally on every startup: it's a no-op when
// there's nothing to clean up, and any error is logged, not fatal.
func CleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return
	}
	old := exe + oldBinarySuffix
	if _, err := os.Stat(old); err != nil {
		return // nothing to clean up
	}
	if err := os.Remove(old); err != nil {
		log.Printf("updater: cleanup old binary %s: %v", old, err)
		return
	}
	log.Printf("updater: cleaned up old binary %s", old)
}
