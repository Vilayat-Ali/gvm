package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withTempEnv(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(EnvRoot, root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	return root
}

func sampleCatalog() []RemoteVersion {
	return []RemoteVersion{
		{Version: "go1.24.0", Stable: true, Filename: "go1.24.0.linux-amd64.tar.gz", SHA256: "aa"},
		{Version: "go1.24rc1", Stable: false, Filename: "go1.24rc1.linux-amd64.tar.gz", SHA256: "bb"},
		{Version: "go1.23.4", Stable: true, Filename: "go1.23.4.linux-amd64.tar.gz", SHA256: "cc"},
		{Version: "go1.23.1", Stable: true, Filename: "go1.23.1.linux-amd64.tar.gz", SHA256: "dd"},
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	withTempEnv(t)

	config, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.AvailableVersions = sampleCatalog()
	config.LastRemoteFetch = time.Now().UnixMilli()

	if err := config.Save(); err != nil {
		t.Fatal(err)
	}
	if !ConfigExists() {
		t.Fatal("ConfigExists() is false after Save()")
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.AvailableVersions) != len(config.AvailableVersions) {
		t.Fatalf("loaded %d versions, saved %d", len(loaded.AvailableVersions), len(config.AvailableVersions))
	}
	if loaded.AvailableVersions[0].Version != "go1.24.0" {
		t.Errorf("unexpected first version %q", loaded.AvailableVersions[0].Version)
	}

	path, _ := ConfigFilePath()
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "config.json" {
			t.Errorf("Save() left a stray file behind: %s", entry.Name())
		}
	}
}

func TestLoadConfigRejectsCorruptFile(t *testing.T) {
	withTempEnv(t)

	path, err := ConfigFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() accepted a corrupt file")
	}
}

func TestLoadConfigDiscardsOldSchema(t *testing.T) {
	withTempEnv(t)

	path, _ := ConfigFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := `{"version":"v2.0.0","download_path":"/usr/local/gvm/go-versions","available_versions":[{"version":"go1.20.0"}]}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(config.AvailableVersions) != 0 {
		t.Error("a legacy catalog should be discarded, not reused")
	}
	if !config.IsStale() {
		t.Error("a migrated config should be considered stale")
	}
}

func TestIsStale(t *testing.T) {
	withTempEnv(t)

	config, _ := NewConfig()
	if !config.IsStale() {
		t.Error("an empty config should be stale")
	}

	config.AvailableVersions = sampleCatalog()
	config.LastRemoteFetch = time.Now().UnixMilli()
	if config.IsStale() {
		t.Error("a freshly fetched config should not be stale")
	}

	config.LastRemoteFetch = time.Now().Add(-2 * catalogTTL).UnixMilli()
	if !config.IsStale() {
		t.Error("a config older than the TTL should be stale")
	}
}

func TestLatestStableSkipsPrereleases(t *testing.T) {
	withTempEnv(t)

	config, _ := NewConfig()
	config.AvailableVersions = []RemoteVersion{
		{Version: "go1.25rc1", Stable: false},
		{Version: "go1.24.0", Stable: true},
	}

	latest, err := config.LatestStable()
	if err != nil {
		t.Fatal(err)
	}
	if latest.Version != "go1.24.0" {
		t.Errorf("LatestStable() = %s, want go1.24.0", latest.Version)
	}
}

func TestResolveVersion(t *testing.T) {
	withTempEnv(t)

	config, _ := NewConfig()
	config.AvailableVersions = sampleCatalog()
	config.LastRemoteFetch = time.Now().UnixMilli()

	ctx := context.Background()

	cases := map[string]string{
		"1.24.0":   "go1.24.0",
		"go1.24.0": "go1.24.0",
		"v1.24.0":  "go1.24.0",
		"1.23":     "go1.23.4",
		"1.24rc1":  "go1.24rc1",
		"latest":   "go1.24.0",
	}
	for input, want := range cases {
		resolved, err := config.ResolveVersion(ctx, input)
		if err != nil {
			t.Fatalf("ResolveVersion(%q) failed: %v", input, err)
		}
		if resolved.Version != want {
			t.Errorf("ResolveVersion(%q) = %s, want %s", input, resolved.Version, want)
		}
	}

	for _, input := range []string{"not-a-version", "../etc/passwd", "1.2.3.4"} {
		if _, err := config.ResolveVersion(ctx, input); err == nil {
			t.Errorf("ResolveVersion(%q) accepted invalid input", input)
		}
	}
}
