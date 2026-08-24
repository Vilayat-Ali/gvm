package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	configSchemaVersion = 3
	catalogTTL          = 24 * time.Hour
	maxStoredVersions   = 60
)

type Config struct {
	SchemaVersion     int             `json:"schema_version"`
	Root              string          `json:"root"`
	LastRemoteFetch   int64           `json:"last_remote_fetch"`
	AvailableVersions []RemoteVersion `json:"available_versions"`
}

func ConfigExists() bool {
	path, err := ConfigFilePath()
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func LoadConfig() (*Config, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("gvm is not configured yet. Run `gvm configure` first")
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%s is corrupted, run `gvm configure --force` to rebuild it: %w", path, err)
	}

	if config.SchemaVersion != configSchemaVersion {
		config.SchemaVersion = configSchemaVersion
		config.AvailableVersions = nil
		config.LastRemoteFetch = 0
	}
	return &config, nil
}

func NewConfig() (*Config, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}
	return &Config{SchemaVersion: configSchemaVersion, Root: root}, nil
}

func LoadOrInitConfig() (*Config, error) {
	if ConfigExists() {
		return LoadConfig()
	}
	return NewConfig()
}

func (c *Config) Save() error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w", dir, err)
	}

	root, err := Root()
	if err != nil {
		return err
	}
	c.SchemaVersion = configSchemaVersion
	c.Root = root

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".config-*.json")
	if err != nil {
		return fmt.Errorf("cannot create a temporary config file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot write config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot flush config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close config: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("cannot set config permissions: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("cannot install config at %s: %w", path, err)
	}
	return nil
}

func (c *Config) IsStale() bool {
	if len(c.AvailableVersions) == 0 || c.LastRemoteFetch <= 0 {
		return true
	}
	fetched := time.UnixMilli(c.LastRemoteFetch)
	return time.Since(fetched) > catalogTTL
}

func (c *Config) Refresh(ctx context.Context) error {
	versions, err := FetchRemoteVersions(ctx)
	if err != nil {
		return err
	}
	if len(versions) > maxStoredVersions {
		versions = versions[:maxStoredVersions]
	}
	c.AvailableVersions = versions
	c.LastRemoteFetch = time.Now().UnixMilli()
	return c.Save()
}

func (c *Config) LatestStable() (*RemoteVersion, error) {
	for i := range c.AvailableVersions {
		parsed, err := ParseVersion(c.AvailableVersions[i].Version)
		if err != nil {
			continue
		}
		if c.AvailableVersions[i].Stable && !parsed.IsPrerelease() {
			return &c.AvailableVersions[i], nil
		}
	}
	return nil, fmt.Errorf("no stable Go release found. Run `gvm list update`")
}

func lookupIn(versions []RemoteVersion, input string) *RemoteVersion {
	canonical, err := CanonicalVersion(input)
	if err != nil {
		return nil
	}
	for i := range versions {
		if versions[i].Version == canonical {
			return &versions[i]
		}
	}

	requested, err := ParseVersion(canonical)
	if err != nil || requested.Patch != 0 || requested.IsPrerelease() {
		return nil
	}

	var best *RemoteVersion
	var bestParsed ParsedVersion
	for i := range versions {
		candidate, err := ParseVersion(versions[i].Version)
		if err != nil || candidate.IsPrerelease() {
			continue
		}
		if candidate.Major != requested.Major || candidate.Minor != requested.Minor {
			continue
		}
		if best == nil || CompareVersions(candidate, bestParsed) > 0 {
			best = &versions[i]
			bestParsed = candidate
		}
	}
	return best
}

func (c *Config) ResolveVersion(ctx context.Context, input string) (*RemoteVersion, error) {
	if input == LatestAlias {
		if c.IsStale() {
			if err := c.Refresh(ctx); err != nil && len(c.AvailableVersions) == 0 {
				return nil, err
			}
		}
		return c.LatestStable()
	}

	if _, err := CanonicalVersion(input); err != nil {
		return nil, err
	}

	if found := lookupIn(c.AvailableVersions, input); found != nil {
		return found, nil
	}

	// The stored catalog only keeps recent releases, so fall back to the
	// complete published index before declaring a version unavailable.
	all, err := FetchRemoteVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("Go %s is not in the local catalog and go.dev could not be reached: %w", DisplayVersion(input), err)
	}

	stored := all
	if len(stored) > maxStoredVersions {
		stored = stored[:maxStoredVersions]
	}
	c.AvailableVersions = stored
	c.LastRemoteFetch = time.Now().UnixMilli()
	_ = c.Save()

	if found := lookupIn(all, input); found != nil {
		return found, nil
	}

	goos, goarch := PlatformOSArch()
	return nil, fmt.Errorf("Go %s is not published for %s/%s", DisplayVersion(input), goos, goarch)
}
