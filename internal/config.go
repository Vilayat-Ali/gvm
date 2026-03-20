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
	AppVersion    = "v2.0.0"
	AppName       = "gvm"
	ConfigDirName = "gvm"
	ConfigFile    = "config.json"
	GoVersionsDir = "go-versions"
)

// Config represents the gvm configuration stored in ~/.config/gvm/config.json.
// It tracks available and downloaded Go versions along with metadata.
type Config struct {
	Version            string                     `json:"version"`
	DownloadPath       string                     `json:"download_path"`
	LastRemoteFetch    int64                      `json:"last_remote_fetch"`
	AvailableVersions  []RemoteVersion            `json:"available_versions"`
	DownloadedVersions map[string]DownloadVersion `json:"downloaded_versions"`
}

// ConfigDir returns the configuration directory path (~/.config/gvm).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", ConfigDirName), nil
}

// ConfigFilePath returns the full path to the config.json file.
func ConfigFilePath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFile), nil
}

// GoDownloadDir returns the directory where Go version tarballs are stored.
// Default: /usr/local/gvm/go-versions/
func GoDownloadDir() (*string, error) {
	systemPath := filepath.Join("/usr/local", AppName, GoVersionsDir)
	if err := os.MkdirAll(systemPath, 0755); err == nil {
		return &systemPath, nil
	}

	return nil, fmt.Errorf("root user permission required. Run again with sudo prefix")
}

// ensureDirectories creates all necessary directories for gvm to function.
func ensureDirectories() error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	goDir, err := GoDownloadDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*goDir, 0755); err != nil {
		return fmt.Errorf("failed to create go versions directory: %w", err)
	}

	return nil
}

// SetupConfig initializes gvm's configuration for first-time use.
// Creates directories, fetches available versions, and saves the config.
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
		DownloadPath:       *goDir,
		LastRemoteFetch:    time.Now().UnixMilli(),
		AvailableVersions:  remoteVersions,
		DownloadedVersions: make(map[string]DownloadVersion),
	}

	return config.Save()
}

// ConfigExists checks if gvm has been configured (config file exists).
func ConfigExists() bool {
	configPath, err := ConfigFilePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}

// LoadConfig reads and parses the gvm configuration file.
func LoadConfig() (*Config, error) {
	configPath, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w. Please run `gvm configure` first", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetDownloadedVersions returns a list of versions that have been downloaded.
func (c *Config) GetDownloadedVersions() *[]DownloadVersion {
	downloadedVersions := make([]DownloadVersion, 0)

	for _, version := range c.AvailableVersions {
		if downloadedVersion, ok := c.DownloadedVersions[version.Version]; ok {
			downloadedVersions = append(downloadedVersions, downloadedVersion)
		}
	}

	return &downloadedVersions
}

// GetLTSVersion returns the first non-RC version as the LTS candidate.
func (c *Config) GetLTSVersion() (*string, error) {
	for _, remote := range c.AvailableVersions {
		if !strings.Contains(remote.Version, "rc") {
			return &remote.Version, nil
		}
	}

	return nil, fmt.Errorf("config error: failed to fetch LTS version")
}

// RemoveDownloadedVersion removes a downloaded Go version from the system.
// Deletes the tarball file and updates the config.
func (c *Config) RemoveDownloadedVersion(version string) error {
	versionFmtToBeDeleted := fmt.Sprintf("go%s", version)

	versionToBeDeleted, exists := c.DownloadedVersions[versionFmtToBeDeleted]
	if !exists {
		return fmt.Errorf("go version %s is not downloaded", versionFmtToBeDeleted)
	}

	if err := os.Remove(versionToBeDeleted.TarPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete the downloaded tar file of %s: %w", versionFmtToBeDeleted, err)
	}

	delete(c.DownloadedVersions, versionFmtToBeDeleted)

	return c.Save()
}

// Save writes the current config to ~/.config/gvm/config.json.
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

// MarkVersionAsDownloaded records that a version has been downloaded.
// Stores the tarball path in the config for future reference.
func (c *Config) MarkVersionAsDownloaded(remoteVersion *RemoteVersion, tarballPath string) error {
	if remoteVersion == nil {
		return fmt.Errorf("invalid remote version provided")
	}

	if _, exists := c.DownloadedVersions[remoteVersion.Version]; exists {
		return nil
	}

	c.DownloadedVersions[remoteVersion.Version] = DownloadVersion{
		Version: remoteVersion.Version,
		TarPath: tarballPath,
	}

	return c.Save()
}

// UpdateAvailableVersions fetches the latest Go versions from GitHub
// and updates the config with new releases.
func (c *Config) UpdateAvailableVersions() error {
	newVersions, err := FetchGoVersionsFromGoGithubRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch new versions: %w", err)
	}

	limit := 10
	if len(newVersions) > limit {
		newVersions = newVersions[:limit]
	}

	existingSet := make(map[string]bool)
	for _, v := range c.AvailableVersions {
		existingSet[v.Version] = true
	}

	var latestVersions []RemoteVersion
	for _, v := range newVersions {
		if !existingSet[v.Version] {
			latestVersions = append(latestVersions, v)
		}
	}

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
