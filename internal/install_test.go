package internal

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func fakeToolchainArchive(t *testing.T, version string) []byte {
	t.Helper()

	script := "#!/bin/sh\necho \"go version " + version + " linux/amd64\"\n"
	files := []struct {
		name string
		body string
		mode int64
	}{
		{"go/VERSION", version, 0o644},
		{"go/bin/go", script, 0o755},
		{"go/bin/gofmt", script, 0o755},
	}

	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	tw := tar.NewWriter(gz)
	for _, file := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name:     file.name,
			Typeflag: tar.TypeReg,
			Mode:     file.mode,
			Size:     int64(len(file.body)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func serveArchive(t *testing.T, payload []byte) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		w.Write(payload)
	}))
	t.Cleanup(server.Close)
	return server
}

func fakeRemote(t *testing.T, version string) (*RemoteVersion, []byte) {
	t.Helper()
	payload := fakeToolchainArchive(t, version)
	sum := sha256.Sum256(payload)
	server := serveArchive(t, payload)
	return &RemoteVersion{
		Version:      version,
		Stable:       true,
		Filename:     version + ".linux-amd64.tar.gz",
		DownloadLink: server.URL + "/" + version + ".tar.gz",
		SHA256:       hex.EncodeToString(sum[:]),
		Size:         int64(len(payload)),
	}, payload
}

func TestCachedTarballPathRejectsHostileNames(t *testing.T) {
	withTempEnv(t)

	cacheDir, _ := CacheDir()
	for _, name := range []string{"../../etc/passwd", "/etc/passwd", "..", "a/../../b", "/", "."} {
		path, err := CachedTarballPath(&RemoteVersion{Filename: name})
		if err != nil {
			continue
		}
		if !WithinDir(cacheDir, path) {
			t.Errorf("CachedTarballPath(%q) escaped the cache directory: %s", name, path)
		}
	}

	path, err := CachedTarballPath(&RemoteVersion{Filename: "go1.22.0.linux-amd64.tar.gz"})
	if err != nil {
		t.Fatal(err)
	}
	cache, _ := CacheDir()
	if !WithinDir(cache, path) {
		t.Errorf("%s is not inside %s", path, cache)
	}
}

func TestEnsureDownloadedVerifiesChecksum(t *testing.T) {
	withTempEnv(t)

	remote, payload := fakeRemote(t, "go1.22.0")
	path, err := EnsureDownloaded(context.Background(), remote, nil)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(payload) {
		t.Fatalf("downloaded %d bytes, expected %d", len(data), len(payload))
	}
}

func TestEnsureDownloadedRejectsBadChecksum(t *testing.T) {
	withTempEnv(t)

	remote, _ := fakeRemote(t, "go1.22.0")
	remote.SHA256 = "0000000000000000000000000000000000000000000000000000000000000000"

	if _, err := EnsureDownloaded(context.Background(), remote, nil); err == nil {
		t.Fatal("a corrupted download was accepted")
	}

	cache, _ := CacheDir()
	entries, err := os.ReadDir(cache)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("a failed download left %d file(s) behind", len(entries))
	}
}

func TestEnsureDownloadedRefusesMissingChecksum(t *testing.T) {
	withTempEnv(t)

	remote, _ := fakeRemote(t, "go1.22.0")
	remote.SHA256 = ""

	if _, err := EnsureDownloaded(context.Background(), remote, nil); err == nil {
		t.Fatal("a download without a published checksum was accepted")
	}
}

func TestEnsureDownloadedReplacesCorruptCache(t *testing.T) {
	withTempEnv(t)

	remote, payload := fakeRemote(t, "go1.22.0")
	target, err := CachedTarballPath(remote)
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureLayout(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, err := EnsureDownloaded(context.Background(), remote, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if len(data) != len(payload) {
		t.Fatal("a corrupt cached archive was not replaced")
	}
}

func TestInstallActivateAndRemove(t *testing.T) {
	root := withTempEnv(t)
	ctx := context.Background()

	first, _ := fakeRemote(t, "go1.22.0")
	second, _ := fakeRemote(t, "go1.23.0")

	for _, remote := range []*RemoteVersion{first, second} {
		if _, err := EnsureInstalled(ctx, remote, nil); err != nil {
			t.Fatalf("installing %s: %v", remote.Version, err)
		}
		if !IsInstalled(remote.Version) {
			t.Fatalf("%s is not reported as installed", remote.Version)
		}
	}

	installed, err := InstalledVersions()
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 2 {
		t.Fatalf("InstalledVersions() returned %d entries, want 2", len(installed))
	}
	if installed[0].Version != "go1.23.0" {
		t.Errorf("installed versions are not sorted newest first: %s", installed[0].Version)
	}

	if err := Activate("1.22.0"); err != nil {
		t.Fatal(err)
	}
	current, err := CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if current != "go1.22.0" {
		t.Fatalf("CurrentVersion() = %s, want go1.22.0", current)
	}

	shim := filepath.Join(root, "bin", "go")
	if goVersionOf(shim) != "go1.22.0" {
		t.Errorf("the shim reports %q, want go1.22.0", goVersionOf(shim))
	}

	if err := RemoveVersion("1.22.0", false); err == nil {
		t.Error("removing the active version should be refused")
	}

	if err := Activate("1.23.0"); err != nil {
		t.Fatal(err)
	}
	if goVersionOf(shim) != "go1.23.0" {
		t.Errorf("the shim did not follow the switch: %q", goVersionOf(shim))
	}

	if err := RemoveVersion("1.22.0", true); err != nil {
		t.Fatal(err)
	}
	if IsInstalled("1.22.0") {
		t.Error("go1.22.0 is still installed after removal")
	}

	cached, err := CachedArchives()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range cached {
		if filepath.Base(name) == "go1.22.0.linux-amd64.tar.gz" {
			t.Error("--purge did not delete the cached archive")
		}
	}

	if err := RemoveVersion("1.21.0", false); err == nil {
		t.Error("removing a version that is not installed should fail")
	}
}

func TestEnsureInstalledIsIdempotent(t *testing.T) {
	withTempEnv(t)
	ctx := context.Background()

	remote, _ := fakeRemote(t, "go1.22.0")
	first, err := EnsureInstalled(ctx, remote, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := EnsureInstalled(ctx, remote, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Errorf("EnsureInstalled returned %s then %s", first, second)
	}
}

func TestActivateRejectsMissingVersion(t *testing.T) {
	withTempEnv(t)

	if err := Activate("1.22.0"); err == nil {
		t.Fatal("activating a version that is not installed should fail")
	}
	if _, err := CurrentVersion(); err == nil {
		t.Fatal("CurrentVersion() should fail when nothing is active")
	}
}

func TestCurrentVersionRejectsForeignLink(t *testing.T) {
	root := withTempEnv(t)
	if err := EnsureLayout(); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/usr/local/go", filepath.Join(root, "current")); err != nil {
		t.Fatal(err)
	}
	if _, err := CurrentVersion(); err == nil {
		t.Fatal("a link pointing outside the versions directory should be rejected")
	}
}

func TestResolveInstalled(t *testing.T) {
	withTempEnv(t)
	ctx := context.Background()

	if _, err := ResolveInstalled("1.22.0"); err == nil {
		t.Fatal("ResolveInstalled should fail when nothing is installed")
	}

	for _, version := range []string{"go1.22.0", "go1.23.4", "go1.23.9"} {
		remote, _ := fakeRemote(t, version)
		if _, err := EnsureInstalled(ctx, remote, nil); err != nil {
			t.Fatal(err)
		}
	}

	cases := map[string]string{
		"1.22.0":   "go1.22.0",
		"go1.22.0": "go1.22.0",
		"1.23":     "go1.23.9",
		"1.23.4":   "go1.23.4",
		"latest":   "go1.23.9",
	}
	for input, want := range cases {
		got, err := ResolveInstalled(input)
		if err != nil {
			t.Fatalf("ResolveInstalled(%q): %v", input, err)
		}
		if got != want {
			t.Errorf("ResolveInstalled(%q) = %s, want %s", input, got, want)
		}
	}

	if _, err := ResolveInstalled("1.21"); err == nil {
		t.Error("ResolveInstalled should fail for a version that is not installed")
	}
}

func TestRemoveResolvesPartialVersion(t *testing.T) {
	withTempEnv(t)
	ctx := context.Background()

	remote, _ := fakeRemote(t, "go1.23.9")
	if _, err := EnsureInstalled(ctx, remote, nil); err != nil {
		t.Fatal(err)
	}

	resolved, err := ResolveInstalled("1.23")
	if err != nil {
		t.Fatal(err)
	}
	if err := RemoveVersion(resolved, true); err != nil {
		t.Fatal(err)
	}
	if IsInstalled("go1.23.9") {
		t.Error("go1.23.9 was not removed")
	}
}
