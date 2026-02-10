package updater

import (
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

func AssetForCurrentPlatform(assets []Asset) (downloadURL string, err error) {
	var want string
	switch runtime.GOOS {
	case "windows":
		want = "terminal-td-windows.exe"
	case "linux":
		want = "terminal-td-linux"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			want = "terminal-td-mac-arm"
		} else {
			want = "terminal-td-mac-intel"
		}
	default:
		return "", fmt.Errorf("unsupported GOOS: %s", runtime.GOOS)
	}
	for _, a := range assets {
		if a.Name == want {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no asset named %q found in release", want)
}

func DownloadUpdate(release *Release, destExePath string) error {
	url, err := AssetForCurrentPlatform(release.Assets)
	if err != nil {
		return err
	}
	log.Printf("updater: downloading %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: %s", resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(destExePath), 0755); err != nil {
		return err
	}
	f, err := os.Create(destExePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		os.Remove(destExePath)
		return err
	}
	if err := f.Chmod(0755); err != nil && runtime.GOOS != "windows" {
		log.Printf("updater: chmod 0755: %v", err)
	}
	return nil
}

func WriteChangelogFile(body string) (path string, err error) {
	updatesDir, err := config.UpdatesPath()
	if err != nil {
		return "", err
	}
	path = filepath.Join(updatesDir, "changelog.txt")
	return path, os.WriteFile(path, []byte(body), 0644)
}
