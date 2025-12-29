package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AppVersion    = "v1.0.0"
	AppName       = "gvm"
	ConfigDirName = "gvm"
	ConfigFile    = "config.json"
	GoVersionsDir = "go-versions"
)

type Config struct {
	Version            string                     `json:"version"`
	DownloadPath       string                     `json:"download_path"`
	LastRemoteFetch    int64                      `json:"last_remote_fetch"`
	AvailableVersions  []RemoteVersion            `json:"available_versions"`
	DownloadedVersions map[string]DownloadVersion `json:"downloaded_versions"`
}

// Path management functions
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", ConfigDirName), nil
}

func ConfigFilePath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFile), nil
}

func GoDownloadDir() (string, error) {
	// Try system location first
	systemPath := filepath.Join("/usr/local", AppName, GoVersionsDir)
	if err := os.MkdirAll(systemPath, 0755); err == nil {
		return systemPath, nil
	}

	// Fall back to user directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".local", AppName, GoVersionsDir), nil
}

// Setup functions
func ensureDirectories() error {
	// Create config directory
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create go versions directory
	goDir, err := GoDownloadDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(goDir, 0755); err != nil {
		return fmt.Errorf("failed to create go versions directory: %w", err)
	}

	return nil
}

func SetupConfig() error {
	if err := ensureDirectories(); err != nil {
		return err
	}

	remoteVersions, err := FetchGoVersionsFromGoGithubRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch remote versions: %w", err)
	}

	goDir, err := GoDownloadDir()
	if err != nil {
		return err
	}

	config := &Config{
		Version:            AppVersion,
		DownloadPath:       goDir,
		LastRemoteFetch:    time.Now().UnixMilli(),
		AvailableVersions:  remoteVersions,
		DownloadedVersions: make(map[string]DownloadVersion),
	}

	return config.Save()
}

// Config file operations
func ConfigExists() bool {
	configPath, err := ConfigFilePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}

func LoadConfig() (*Config, error) {
	configPath, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *Config) GetLTSVersion() (*string, error) {
	for _, remote := range c.AvailableVersions {
		if !strings.Contains(remote.Version, "rc") {
			return &remote.Version, nil
		}
	}

	return nil, fmt.Errorf("Config Error: failed to fetch lts from config")
}

func (c *Config) Save() error {
	configPath, err := ConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Config operations
func (c *Config) MarkVersionAsDownloaded(idx uint, binaryPath string) error {
	if int(idx) >= len(c.AvailableVersions) {
		return fmt.Errorf("invalid version index: %d", idx)
	}

	version := c.AvailableVersions[idx]
	if _, exists := c.DownloadedVersions[version.Version]; exists {
		return nil // Already downloaded
	}

	c.DownloadedVersions[version.Version] = DownloadVersion{
		Version: version.Version,
		BinPath: binaryPath,
	}

	return nil
}

func (c *Config) UpdateAvailableVersions() error {
	newVersions, err := FetchGoVersionsFromGoGithubRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch new versions: %w", err)
	}

	// Keep only top 10 versions
	limit := 10
	if len(newVersions) > limit {
		newVersions = newVersions[:limit]
	}

	// Create a set of existing versions for quick lookup
	existingSet := make(map[string]bool)
	for _, v := range c.AvailableVersions {
		existingSet[v.Version] = true
	}

	// Add only new versions
	var latestVersions []RemoteVersion
	for _, v := range newVersions {
		if !existingSet[v.Version] {
			latestVersions = append(latestVersions, v)
		}
	}

	// Add existing versions to fill up to limit
	for _, v := range c.AvailableVersions {
		if len(latestVersions) >= limit {
			break
		}
		latestVersions = append(latestVersions, v)
	}

	c.LastRemoteFetch = time.Now().UnixMilli()
	c.AvailableVersions = latestVersions

	return c.Save()
}
