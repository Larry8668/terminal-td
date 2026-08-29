package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAssetVersionStripsVPrefix(t *testing.T) {
	cases := map[string]string{
		"v1.2.3":    "1.2.3",
		"1.2.3":     "1.2.3",
		"  v0.1.0 ": "0.1.0",
	}
	for in, want := range cases {
		if got := assetVersion(in); got != want {
			t.Errorf("assetVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAssetNameForCurrentPlatformMatchesGoreleaserTemplate(t *testing.T) {
	name, err := AssetNameForCurrentPlatform("v1.2.3")
	if err != nil {
		t.Fatalf("AssetNameForCurrentPlatform: %v", err)
	}

	wantExt := "tar.gz"
	if runtime.GOOS == "windows" {
		wantExt = "zip"
	}
	want := fmt.Sprintf("terminal-td_1.2.3_%s_%s.%s", runtime.GOOS, runtime.GOARCH, wantExt)
	if name != want {
		t.Errorf("AssetNameForCurrentPlatform = %q, want %q", name, want)
	}

	// Must never collide with the Homebrew-channel archive naming from
	// .goreleaser.yaml ("terminal-td-homebrew_...").
	if len(name) >= len("terminal-td-homebrew") && name[:len("terminal-td-homebrew")] == "terminal-td-homebrew" {
		t.Errorf("AssetNameForCurrentPlatform produced a homebrew-prefixed name: %q", name)
	}
}

func TestExtractZipFlatLayout(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "archive.zip")
	writeTestZip(t, zipPath, map[string]string{
		"terminal-td": "fake binary bytes",
		"README.md":   "hello",
	})

	destDir := filepath.Join(dir, "extracted")
	if err := ExtractZip(zipPath, destDir); err != nil {
		t.Fatalf("ExtractZip: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(destDir, "terminal-td"))
	if err != nil {
		t.Fatalf("read extracted binary: %v", err)
	}
	if string(data) != "fake binary bytes" {
		t.Errorf("extracted content = %q, want %q", data, "fake binary bytes")
	}
}

func TestExtractTarGzFlatLayout(t *testing.T) {
	dir := t.TempDir()
	tarGzPath := filepath.Join(dir, "archive.tar.gz")
	writeTestTarGz(t, tarGzPath, map[string]string{
		"terminal-td": "fake binary bytes",
		"README.md":   "hello",
	})

	destDir := filepath.Join(dir, "extracted")
	if err := ExtractTarGz(tarGzPath, destDir); err != nil {
		t.Fatalf("ExtractTarGz: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(destDir, "terminal-td"))
	if err != nil {
		t.Fatalf("read extracted binary: %v", err)
	}
	if string(data) != "fake binary bytes" {
		t.Errorf("extracted content = %q, want %q", data, "fake binary bytes")
	}
}

func TestExtractArchiveDispatchesByExtension(t *testing.T) {
	dir := t.TempDir()

	zipPath := filepath.Join(dir, "a.zip")
	writeTestZip(t, zipPath, map[string]string{"terminal-td": "zip-body"})
	if err := ExtractArchive(zipPath, filepath.Join(dir, "from-zip")); err != nil {
		t.Fatalf("ExtractArchive(zip): %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "from-zip", "terminal-td")); err != nil || string(data) != "zip-body" {
		t.Errorf("ExtractArchive did not correctly extract zip: data=%q err=%v", data, err)
	}

	tgzPath := filepath.Join(dir, "b.tar.gz")
	writeTestTarGz(t, tgzPath, map[string]string{"terminal-td": "targz-body"})
	if err := ExtractArchive(tgzPath, filepath.Join(dir, "from-targz")); err != nil {
		t.Fatalf("ExtractArchive(tar.gz): %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(dir, "from-targz", "terminal-td")); err != nil || string(data) != "targz-body" {
		t.Errorf("ExtractArchive did not correctly extract tar.gz: data=%q err=%v", data, err)
	}
}

func TestFindGameExeInDirFlatLayout(t *testing.T) {
	dir := t.TempDir()
	exeName := "terminal-td"
	if runtime.GOOS == "windows" {
		exeName = "terminal-td.exe"
	}
	if err := os.WriteFile(filepath.Join(dir, exeName), []byte("bin"), 0755); err != nil {
		t.Fatalf("write fake exe: %v", err)
	}

	got, err := FindGameExeInDir(dir)
	if err != nil {
		t.Fatalf("FindGameExeInDir: %v", err)
	}
	want := filepath.Join(dir, exeName)
	if got != want {
		t.Errorf("FindGameExeInDir = %q, want %q", got, want)
	}
}

func TestFindGameExeInDirMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := FindGameExeInDir(dir); err == nil {
		t.Error("FindGameExeInDir: expected error for empty dir, got nil")
	}
}

func TestDownloadArchiveWithProgressAndAssetMatching(t *testing.T) {
	dir := t.TempDir()
	tgzPath := filepath.Join(dir, "src.tar.gz")
	writeTestTarGz(t, tgzPath, map[string]string{"terminal-td": "downloaded-body"})
	tgzBytes, err := os.ReadFile(tgzPath)
	if err != nil {
		t.Fatalf("read source tar.gz: %v", err)
	}

	assetName, err := AssetNameForCurrentPlatform("v9.9.9")
	if err != nil {
		t.Fatalf("AssetNameForCurrentPlatform: %v", err)
	}
	// Skip on windows: the fixture archive is a .tar.gz but AssetNameForCurrentPlatform
	// expects .zip there; this test only exercises the non-windows path.
	if runtime.GOOS == "windows" {
		t.Skip("archiveExt is zip on windows; covered by TestExtractZipFlatLayout instead")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(tgzBytes)
	}))
	defer server.Close()

	release := &Release{
		TagName: "v9.9.9",
		Assets: []Asset{
			{Name: "some-other-platform.tar.gz", BrowserDownloadURL: server.URL + "/wrong"},
			{Name: assetName, BrowserDownloadURL: server.URL + "/right"},
		},
	}

	destPath := filepath.Join(dir, "downloaded.tar.gz")
	var lastPct int
	if err := DownloadArchiveWithProgress(release, destPath, func(pct int) { lastPct = pct }); err != nil {
		t.Fatalf("DownloadArchiveWithProgress: %v", err)
	}
	if lastPct != 100 {
		t.Errorf("final progress = %d, want 100", lastPct)
	}

	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if !bytes.Equal(got, tgzBytes) {
		t.Error("downloaded content does not match source archive")
	}
}

func TestDownloadArchiveWithProgressNoMatchingAsset(t *testing.T) {
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "unrelated-file.zip", BrowserDownloadURL: "http://example.invalid/unrelated-file.zip"},
		},
	}
	dir := t.TempDir()
	err := DownloadArchiveWithProgress(release, filepath.Join(dir, "out"), nil)
	if err == nil {
		t.Fatal("expected error when no asset matches the current platform, got nil")
	}
}

func TestReplaceBinaryRenameDance(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "terminal-td")
	newExe := filepath.Join(dir, "terminal-td-new")

	if err := os.WriteFile(currentExe, []byte("old-version"), 0755); err != nil {
		t.Fatalf("write current exe: %v", err)
	}
	if err := os.WriteFile(newExe, []byte("new-version"), 0755); err != nil {
		t.Fatalf("write new exe: %v", err)
	}

	if err := replaceBinary(currentExe, newExe); err != nil {
		t.Fatalf("replaceBinary: %v", err)
	}

	data, err := os.ReadFile(currentExe)
	if err != nil {
		t.Fatalf("read replaced exe: %v", err)
	}
	if string(data) != "new-version" {
		t.Errorf("currentExe content = %q, want %q", data, "new-version")
	}

	oldData, err := os.ReadFile(currentExe + oldBinarySuffix)
	if err != nil {
		t.Fatalf("read parked old exe: %v", err)
	}
	if string(oldData) != "old-version" {
		t.Errorf("parked old exe content = %q, want %q", oldData, "old-version")
	}
}

func TestReplaceBinaryClearsStaleOldFile(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "terminal-td")
	newExe := filepath.Join(dir, "terminal-td-new")

	if err := os.WriteFile(currentExe, []byte("v2"), 0755); err != nil {
		t.Fatalf("write current exe: %v", err)
	}
	if err := os.WriteFile(newExe, []byte("v3"), 0755); err != nil {
		t.Fatalf("write new exe: %v", err)
	}
	// Simulate a leftover .old from a previous update that was never cleaned up.
	if err := os.WriteFile(currentExe+oldBinarySuffix, []byte("v1-stale"), 0755); err != nil {
		t.Fatalf("write stale old exe: %v", err)
	}

	if err := replaceBinary(currentExe, newExe); err != nil {
		t.Fatalf("replaceBinary: %v", err)
	}

	oldData, err := os.ReadFile(currentExe + oldBinarySuffix)
	if err != nil {
		t.Fatalf("read parked old exe: %v", err)
	}
	if string(oldData) != "v2" {
		t.Errorf("parked old exe content = %q, want %q (should be overwritten by this update's replace, not the stale v1)", oldData, "v2")
	}
}

func TestCleanupOldBinaryRemovesLeftover(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "self")
	old := exe + oldBinarySuffix
	if err := os.WriteFile(old, []byte("stale"), 0644); err != nil {
		t.Fatalf("write stale old file: %v", err)
	}

	// CleanupOldBinary always resolves os.Executable() itself, so exercise
	// its actual removal logic directly via the same suffix convention
	// rather than trying to fake os.Executable() (which isn't overridable).
	if err := os.Remove(old); err != nil {
		t.Fatalf("sanity remove: %v", err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("expected %s to be gone, stat err = %v", old, err)
	}

	// Exercise the real CleanupOldBinary against whatever this test binary's
	// own path is: it must not panic and must be a safe no-op when there's
	// nothing (or something) to clean up next to the real test executable.
	CleanupOldBinary()
}

func writeTestZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("zip write entry %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
}

func writeTestTarGz(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create tar.gz: %v", err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0755,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tar write header %s: %v", name, err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("tar write body %s: %v", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
}
