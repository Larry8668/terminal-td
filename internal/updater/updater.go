package updater

import (
	"archive/zip"
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
	DefaultOwner = "Larry8668"
	DefaultRepo  = "terminal-td"
	APIBase      = "https://api.github.com"
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

func FetchLatest(owner, repo string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", APIBase, owner, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("releases/latest: %s", resp.Status)
	}
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
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

func platformSuffix() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "windows-amd64", nil
	case "linux":
		return "linux-amd64", nil
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "darwin-arm64", nil
		}
		return "darwin-amd64", nil
	default:
		return "", fmt.Errorf("unsupported GOOS: %s", runtime.GOOS)
	}
}

func gameExeName() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "terminal-td-windows.exe", nil
	case "linux":
		return "terminal-td-linux", nil
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "terminal-td-mac-arm", nil
		}
		return "terminal-td-mac-intel", nil
	default:
		return "", fmt.Errorf("unsupported GOOS: %s", runtime.GOOS)
	}
}

func ZipAssetNameForCurrentPlatform(tag string) (string, error) {
	suffix, err := platformSuffix()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("terminal-td-%s-%s.zip", tag, suffix), nil
}

func DownloadZip(release *Release, destPath string) error {
	name, err := ZipAssetNameForCurrentPlatform(release.TagName)
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
		return fmt.Errorf("no asset named %q found in release", name)
	}
	log.Printf("updater: downloading %s", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
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
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		os.Remove(destPath)
		return err
	}
	return nil
}

func ExtractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
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
	return nil
}

func FindGameExeInDir(dir string) (string, error) {
	exeName, err := gameExeName()
	if err != nil {
		return "", err
	}
	atRoot := filepath.Join(dir, exeName)
	if info, err := os.Stat(atRoot); err == nil && !info.IsDir() {
		return atRoot, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var singleDir string
	for _, e := range entries {
		if e.IsDir() {
			if singleDir != "" {
				return "", fmt.Errorf("multiple top-level dirs in zip")
			}
			singleDir = e.Name()
		}
	}
	if singleDir == "" {
		return "", fmt.Errorf("game binary %q not found in zip", exeName)
	}
	inSub := filepath.Join(dir, singleDir, exeName)
	if info, err := os.Stat(inSub); err == nil && !info.IsDir() {
		return inSub, nil
	}
	return "", fmt.Errorf("game binary %q not found in zip", exeName)
}

func WriteChangelogFile(body string) (path string, err error) {
	updatesDir, err := config.UpdatesPath()
	if err != nil {
		return "", err
	}
	path = filepath.Join(updatesDir, "changelog.txt")
	return path, os.WriteFile(path, []byte(body), 0644)
}
