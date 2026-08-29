package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/semver"
	"terminal-td/internal/config"
)

const (
	DefaultOwner   = "Larry8668"
	DefaultRepo    = "terminal-td"
	DefaultAPIBase = "https://api.github.com"
	EnvUpdateAPI   = "TERMINAL_TD_UPDATE_API_BASE"
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

type Progress struct {
	Step    string
	Percent int
	Done    bool
	Err     error
}

func apiBase() string {
	if b := os.Getenv(EnvUpdateAPI); b != "" {
		return strings.TrimSuffix(b, "/")
	}
	return DefaultAPIBase
}

func FetchLatest(owner, repo string) (*Release, error) {
	base := apiBase()
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", base, owner, repo)
	log.Printf("updater: fetch latest release from %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("updater: fetch failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("updater: fetch returned %s", resp.Status)
		return nil, fmt.Errorf("releases/latest: %s", resp.Status)
	}
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Printf("updater: decode response failed: %v", err)
		return nil, err
	}
	log.Printf("updater: got release %s with %d assets", release.TagName, len(release.Assets))
	return &release, nil
}

func NormalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v != "" && v[0] != 'v' {
		return "v" + v
	}
	return v
}

func IsNewer(current, latest string) bool {
	c := NormalizeVersion(current)
	l := NormalizeVersion(latest)
	if !semver.IsValid(c) || !semver.IsValid(l) {
		return false
	}
	return semver.Compare(c, l) < 0
}

// platformSuffix returns the "<goos>_<goarch>" suffix GoReleaser uses in its
// archive name_template ("terminal-td_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
// in .goreleaser.yaml). Go's raw GOOS/GOARCH values are used verbatim there
// (no name replacements configured), so this must track that template.
func platformSuffix() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "windows_amd64", nil
	case "linux":
		return "linux_amd64", nil
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "darwin_arm64", nil
		}
		return "darwin_amd64", nil
	default:
		return "", fmt.Errorf("unsupported GOOS: %s", runtime.GOOS)
	}
}

// archiveExt returns the archive format GoReleaser produces for the current
// platform: zip for Windows, tar.gz everywhere else (see format_overrides in
// .goreleaser.yaml).
func archiveExt() string {
	if runtime.GOOS == "windows" {
		return "zip"
	}
	return "tar.gz"
}

// gameExeName returns the binary's filename as it sits inside the archive.
// GoReleaser archives are flat (no wrapping version/platform folder), so
// this is just the binary name itself.
func gameExeName() (string, error) {
	switch runtime.GOOS {
	case "windows", "linux", "darwin":
		if runtime.GOOS == "windows" {
			return "terminal-td.exe", nil
		}
		return "terminal-td", nil
	default:
		return "", fmt.Errorf("unsupported GOOS: %s", runtime.GOOS)
	}
}

// assetVersion strips the leading "v" GitHub release tags use (e.g. "v1.2.3")
// down to the bare semver GoReleaser's {{ .Version }} template variable
// produces (e.g. "1.2.3"), since that's what ends up in the archive filename.
func assetVersion(tag string) string {
	return strings.TrimPrefix(strings.TrimSpace(tag), "v")
}

// AssetNameForCurrentPlatform returns the GitHub release asset name for the
// running platform, matching the archives.name_template in .goreleaser.yaml.
// It intentionally can never match the "terminal-td-homebrew_..." assets
// built for the Homebrew cask (see .goreleaser.yaml) -- those are a distinct
// prefix and are never meant to be fetched by this in-app updater, which is
// disabled entirely for Homebrew installs (see buildinfo.IsHomebrew).
func AssetNameForCurrentPlatform(tag string) (string, error) {
	suffix, err := platformSuffix()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("terminal-td_%s_%s.%s", assetVersion(tag), suffix, archiveExt()), nil
}

func DownloadArchive(release *Release, destPath string) error {
	return DownloadArchiveWithProgress(release, destPath, nil)
}

func DownloadArchiveWithProgress(release *Release, destPath string, setPercent func(int)) error {
	name, err := AssetNameForCurrentPlatform(release.TagName)
	if err != nil {
		return err
	}
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == name {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		log.Printf("updater: no asset %q in release (have %d assets)", name, len(release.Assets))
		return fmt.Errorf("no asset named %q found in release", name)
	}
	log.Printf("updater: downloading archive from %s -> %s", downloadURL, destPath)
	resp, err := http.Get(downloadURL)
	if err != nil {
		log.Printf("updater: download request failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("updater: download returned %s", resp.Status)
		return fmt.Errorf("download: %s", resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()
	var n int64
	total := resp.ContentLength
	if total > 0 && setPercent != nil {
		buf := make([]byte, 32*1024)
		var lastPct int
		for {
			nr, er := resp.Body.Read(buf)
			if nr > 0 {
				nw, ew := f.Write(buf[:nr])
				n += int64(nw)
				if ew != nil {
					os.Remove(destPath)
					return ew
				}
				if nw != nr {
					os.Remove(destPath)
					return io.ErrShortWrite
				}
				pct := int(100 * n / total)
				if pct != lastPct && pct <= 100 {
					lastPct = pct
					setPercent(pct)
				}
			}
			if er != nil {
				if er != io.EOF {
					os.Remove(destPath)
					return er
				}
				break
			}
		}
		setPercent(100)
	} else {
		n, err = io.Copy(f, resp.Body)
		if err != nil {
			os.Remove(destPath)
			log.Printf("updater: download write failed: %v", err)
			return err
		}
		if setPercent != nil {
			setPercent(100)
		}
	}
	log.Printf("updater: downloaded %d bytes to %s", n, destPath)
	return nil
}

func ExtractZip(zipPath, destDir string) error {
	log.Printf("updater: extracting %s -> %s", zipPath, destDir)
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Printf("updater: open zip failed: %v", err)
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		join := filepath.Join(destDir, f.Name)
		if !filepath.IsLocal(f.Name) {
			return fmt.Errorf("zip slip: %s", f.Name)
		}
		abs, err := filepath.Abs(join)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(abs, destAbs+string(filepath.Separator)) {
			return fmt.Errorf("zip slip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(join, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(join), 0755); err != nil {
			return err
		}
		out, err := os.OpenFile(join, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
		if runtime.GOOS != "windows" {
			_ = os.Chmod(join, f.Mode())
		}
	}
	log.Printf("updater: extracted %d entries to %s", len(r.File), destDir)
	return nil
}

// ExtractTarGz extracts a .tar.gz archive as produced by GoReleaser for
// non-Windows platforms. Mirrors ExtractZip's path-traversal ("tar slip")
// protection.
func ExtractTarGz(tarGzPath, destDir string) error {
	log.Printf("updater: extracting %s -> %s", tarGzPath, destDir)
	f, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("open tar.gz: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("open gzip stream: %w", err)
	}
	defer gz.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gz)
	count := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}
		if !filepath.IsLocal(hdr.Name) {
			return fmt.Errorf("tar slip: %s", hdr.Name)
		}
		join := filepath.Join(destDir, hdr.Name)
		abs, err := filepath.Abs(join)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(abs, destAbs+string(filepath.Separator)) {
			return fmt.Errorf("tar slip: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(join, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(join), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(join, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			count++
		default:
			// Skip symlinks, hardlinks, etc. -- none are expected in our
			// release archives.
		}
	}
	log.Printf("updater: extracted %d entries to %s", count, destDir)
	return nil
}

// ExtractArchive dispatches to ExtractZip or ExtractTarGz based on the
// archive's extension.
func ExtractArchive(archivePath, destDir string) error {
	if strings.HasSuffix(archivePath, ".zip") {
		return ExtractZip(archivePath, destDir)
	}
	return ExtractTarGz(archivePath, destDir)
}

// FindGameExeInDir locates the game binary directly inside an extracted
// archive. GoReleaser archives are flat (no wrapping version/platform
// folder), so this is a simple lookup rather than a folder-name guess.
func FindGameExeInDir(dir string) (string, error) {
	exeName, err := gameExeName()
	if err != nil {
		return "", err
	}
	exePath := filepath.Join(dir, exeName)
	if info, err := os.Stat(exePath); err == nil && !info.IsDir() {
		log.Printf("updater: found game exe at %s", exePath)
		return exePath, nil
	}
	log.Printf("updater: expected %s not found in extracted archive", exeName)
	return "", fmt.Errorf("expected game binary %q not found in extracted archive", exeName)
}

func WriteChangelogFile(body string) (path string, err error) {
	updatesDir, err := config.UpdatesPath()
	if err != nil {
		return "", err
	}
	path = filepath.Join(updatesDir, "changelog.txt")
	return path, os.WriteFile(path, []byte(body), 0644)
}
