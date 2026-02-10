package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

const (
	ConfigVersion  = 1
	AppConfigDir   = "terminal-td"
	ConfigFileName = "config.json"
	UpdatesDir     = "updates"
)

type Config struct {
	Version         int  `json:"config_version"`
	CheckForUpdates bool `json:"check_for_updates"`
}

func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, AppConfigDir)
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}
	return appDir, nil
}

func UpdatesPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	updatesDir := filepath.Join(dir, UpdatesDir)
	if err := os.MkdirAll(updatesDir, 0755); err != nil {
		return "", err
	}
	return updatesDir, nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		log.Printf("config: failed to parse, using defaults: %v", err)
		return Default(), nil
	}
	return migrate(&c), nil
}

func Default() *Config {
	return &Config{
		Version:         ConfigVersion,
		CheckForUpdates: true,
	}
}

func migrate(c *Config) *Config {
	// Future: bump ConfigVersion and migrate old fields here
	if c.Version < 1 {
		c.Version = ConfigVersion
		c.CheckForUpdates = true
	}
	return c
}

func Save(c *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
