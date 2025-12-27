package internal

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"time"
)

// Represents the layout of gvm config file
type Config struct {
	// version of gvm cli
	Version string `json:"version"`
	// configured path to store downloaded golang version artifacts
	DownloadPath string `json:"download_path"`
	// last time when list of available versions was fetched from official github
	// in millis unix time format
	LastRemoteFetch int64 `json:"last_remote_fetch"`
	// Ordered list of remote versions available for download
	AvailableVersions []RemoteVersion `json:"available_versions"`
	// version of golang downloaded and available to use and switch in local
	DownloadedVersions map[string]DownloadVersion `json:"downloaded_versions"`
}

// Fetches config from config file. Throws error if not found
func GetConfigOrThrow(configFilePath string) (*Config, error) {
	file, err := os.ReadFile(configFilePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var config Config

	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Marks an available version as download in the config file
func (c *Config) MarkAvailableVersionAsDownload(idx uint, binaryPath string) error {
	availableVersion := c.AvailableVersions[idx]

	if _, ok := c.DownloadedVersions[availableVersion.Version]; ok {
		return nil
	}

	downloadEntry := DownloadVersion{
		Version: availableVersion.Version,
		BinPath: binaryPath,
	}

	c.DownloadedVersions[availableVersion.Version] = downloadEntry
	return nil
}

// Fetches latest releases and diffs with local changes and then writes the latest ones
func (c *Config) UpdateAvailableVersions() error {
	newlyFetchedVersions, err := FetchGoVersionsFromGoGithubRelease()
	if err != nil {
		return err
	}

	var latestVersions []RemoteVersion = make([]RemoteVersion, 10)

	for idx := range 10 {
		if !slices.Contains(c.AvailableVersions, newlyFetchedVersions[idx]) {
			latestVersions = append(latestVersions, newlyFetchedVersions[idx])
		}
	}

	newVersionCount := len(latestVersions)

	for idx := range 10 - newVersionCount {
		latestVersions = append(latestVersions, c.AvailableVersions[idx])
	}

	c.LastRemoteFetch = time.Now().UnixMilli()
	c.AvailableVersions = latestVersions
	return nil
}

// Writes changes to config object to the config file at configFilePath
func (c *Config) SaveConfig(configFilePath string) error {
	configContentJson, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFilePath, configContentJson, 0655); err != nil {
		return err
	}

	return nil
}
