package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var shimNames = []string{"go", "gofmt"}

type InstalledVersion struct {
	Version string
	Path    string
	Size    int64
}

func GoBinary(rootDir string) string {
	name := "go"
	if isWindows() {
		name = "go.exe"
	}
	return filepath.Join(rootDir, "bin", name)
}

func isWindows() bool {
	goos, _ := PlatformOSArch()
	return goos == "windows"
}

func CachedTarballPath(rv *RemoteVersion) (string, error) {
	cache, err := CacheDir()
	if err != nil {
		return "", err
	}
	base := filepath.Base(filepath.Clean(rv.Filename))
	if base == "." || base == string(os.PathSeparator) || strings.Contains(rv.Filename, "..") {
		return "", fmt.Errorf("refusing to use suspicious archive name %q", rv.Filename)
	}
	return filepath.Join(cache, base), nil
}

func fileMatchesChecksum(path, expected string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.CopyBuffer(hash, file, make([]byte, copyBufferSize)); err != nil {
		return false, err
	}
	return strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expected), nil
}

func EnsureDownloaded(ctx context.Context, rv *RemoteVersion, progress io.Writer) (string, error) {
	if rv.SHA256 == "" {
		return "", fmt.Errorf("no published checksum for %s, refusing to install", rv.Version)
	}
	if err := EnsureLayout(); err != nil {
		return "", err
	}

	target, err := CachedTarballPath(rv)
	if err != nil {
		return "", err
	}

	if info, statErr := os.Stat(target); statErr == nil && info.Mode().IsRegular() {
		ok, err := fileMatchesChecksum(target, rv.SHA256)
		if err == nil && ok {
			return target, nil
		}
		if err := RemoveManaged(target); err != nil {
			return "", err
		}
	}

	resp, err := httpGet(ctx, downloadClient, rv.DownloadLink)
	if err != nil {
		return "", fmt.Errorf("cannot download %s: %w", rv.Version, err)
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp(filepath.Dir(target), ".download-*.part")
	if err != nil {
		return "", fmt.Errorf("cannot create a temporary download file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	hash := sha256.New()
	writers := []io.Writer{tmp, hash}
	if progress != nil {
		writers = append(writers, progress)
	}

	written, err := io.CopyBuffer(io.MultiWriter(writers...), resp.Body, make([]byte, copyBufferSize))
	if err != nil {
		return "", fmt.Errorf("download of %s failed: %w", rv.Version, err)
	}
	if rv.Size > 0 && written != rv.Size {
		return "", fmt.Errorf("download of %s is incomplete (%d of %d bytes)", rv.Version, written, rv.Size)
	}
	if sum := hex.EncodeToString(hash.Sum(nil)); !strings.EqualFold(sum, rv.SHA256) {
		return "", fmt.Errorf("checksum mismatch for %s: expected %s, got %s", rv.Version, rv.SHA256, sum)
	}
	if err := tmp.Sync(); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, target); err != nil {
		return "", fmt.Errorf("cannot store the archive at %s: %w", target, err)
	}
	return target, nil
}

func IsInstalled(version string) bool {
	dir, err := VersionDir(version)
	if err != nil {
		return false
	}
	info, err := os.Stat(GoBinary(dir))
	return err == nil && info.Mode().IsRegular()
}

func verifyGoRoot(dir string) error {
	binary := GoBinary(dir)
	info, err := os.Stat(binary)
	if err != nil {
		return fmt.Errorf("the extracted archive has no go binary at %s", binary)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", binary)
	}
	if err := os.Chmod(binary, info.Mode().Perm()|0o111); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := probeCommand(ctx, binary, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("the extracted Go toolchain does not run (%s): %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// GOTOOLCHAIN=local stops Go from downloading and re-execing a different
// toolchain when the working directory has a go.mod requiring a newer one,
// which would otherwise make this probe hang and report a false failure.
func probeCommand(ctx context.Context, binary, dir string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binary, "version")
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOFLAGS=")
	if dir != "" {
		cmd.Env = append(cmd.Env, "GOROOT="+dir)
		cmd.Dir = dir
	} else {
		cmd.Dir = os.TempDir()
	}
	return cmd
}

func EnsureInstalled(ctx context.Context, rv *RemoteVersion, progress io.Writer) (string, error) {
	dir, err := VersionDir(rv.Version)
	if err != nil {
		return "", err
	}
	if IsInstalled(rv.Version) {
		return dir, nil
	}

	archive, err := EnsureDownloaded(ctx, rv, progress)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(archive, ".tar.gz") {
		return "", fmt.Errorf("%s is not a supported archive format on this platform", filepath.Base(archive))
	}

	staging := dir + fmt.Sprintf(".staging-%d", os.Getpid())
	if err := RemoveManaged(staging); err != nil {
		return "", err
	}
	defer func() { _ = RemoveManaged(staging) }()

	if err := ExtractTarGz(archive, staging, nil); err != nil {
		return "", err
	}
	if err := verifyGoRoot(staging); err != nil {
		return "", err
	}
	if err := RemoveManaged(dir); err != nil {
		return "", err
	}
	if err := os.Rename(staging, dir); err != nil {
		return "", fmt.Errorf("cannot install %s at %s: %w", rv.Version, dir, err)
	}
	return dir, nil
}

func Activate(version string) error {
	canonical, err := CanonicalVersion(version)
	if err != nil {
		return err
	}
	dir, err := VersionDir(canonical)
	if err != nil {
		return err
	}
	if err := verifyGoRoot(dir); err != nil {
		return fmt.Errorf("Go %s is not usable: %w", DisplayVersion(canonical), err)
	}

	link, err := CurrentLink()
	if err != nil {
		return err
	}
	tmpLink := link + fmt.Sprintf(".tmp-%d", os.Getpid())
	os.Remove(tmpLink)
	if err := os.Symlink(dir, tmpLink); err != nil {
		return fmt.Errorf("cannot create the version link: %w", err)
	}
	if err := os.Rename(tmpLink, link); err != nil {
		os.Remove(tmpLink)
		return fmt.Errorf("cannot activate %s: %w", DisplayVersion(canonical), err)
	}
	return refreshShims()
}

func refreshShims() error {
	shims, err := ShimDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(shims, 0o755); err != nil {
		return err
	}
	for _, name := range shimNames {
		target := filepath.Join(shims, name)
		relative := filepath.Join("..", "current", "bin", name)
		tmp := target + fmt.Sprintf(".tmp-%d", os.Getpid())
		os.Remove(tmp)
		if err := os.Symlink(relative, tmp); err != nil {
			return err
		}
		if err := os.Rename(tmp, target); err != nil {
			os.Remove(tmp)
			return err
		}
	}
	return nil
}

func CurrentVersion() (string, error) {
	link, err := CurrentLink()
	if err != nil {
		return "", err
	}
	resolved, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	canonical, err := CanonicalVersion(filepath.Base(resolved))
	if err != nil {
		return "", fmt.Errorf("the active version link points at an unexpected path %q", resolved)
	}
	if !IsInstalled(canonical) {
		return "", fmt.Errorf("the active version %s is missing from disk", DisplayVersion(canonical))
	}
	return canonical, nil
}

func InstalledVersions() ([]InstalledVersion, error) {
	dir, err := VersionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	installed := make([]InstalledVersion, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		canonical, err := CanonicalVersion(entry.Name())
		if err != nil || canonical != entry.Name() {
			continue
		}
		versionPath := filepath.Join(dir, entry.Name())
		if _, err := os.Stat(GoBinary(versionPath)); err != nil {
			continue
		}
		installed = append(installed, InstalledVersion{Version: canonical, Path: versionPath})
	}

	sort.SliceStable(installed, func(i, j int) bool {
		a, _ := ParseVersion(installed[i].Version)
		b, _ := ParseVersion(installed[j].Version)
		return CompareVersions(a, b) > 0
	})
	return installed, nil
}

func RemoveVersion(version string, purgeCache bool) error {
	canonical, err := CanonicalVersion(version)
	if err != nil {
		return err
	}
	dir, err := VersionDir(canonical)
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(dir); statErr != nil {
		return fmt.Errorf("Go %s is not installed", DisplayVersion(canonical))
	}
	if current, err := CurrentVersion(); err == nil && current == canonical {
		return fmt.Errorf("Go %s is currently active. Switch with `gvm use <other-version>` first", DisplayVersion(canonical))
	}
	if err := RemoveManaged(dir); err != nil {
		return err
	}
	if !purgeCache {
		return nil
	}
	return purgeCachedArchives(canonical)
}

func purgeCachedArchives(canonical string) error {
	cache, err := CacheDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(cache)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), canonical+".") {
			continue
		}
		if err := RemoveManaged(filepath.Join(cache, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func CachedArchives() ([]string, error) {
	cache, err := CacheDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(cache)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tar.gz") {
			names = append(names, filepath.Join(cache, entry.Name()))
		}
	}
	return names, nil
}

func ResolveInstalled(input string) (string, error) {
	installed, err := InstalledVersions()
	if err != nil {
		return "", err
	}
	if len(installed) == 0 {
		return "", fmt.Errorf("no Go versions are installed")
	}

	if input == LatestAlias {
		return installed[0].Version, nil
	}

	candidates := make([]RemoteVersion, 0, len(installed))
	for _, version := range installed {
		candidates = append(candidates, RemoteVersion{Version: version.Version})
	}
	if found := lookupIn(candidates, input); found != nil {
		return found.Version, nil
	}

	names := make([]string, 0, len(installed))
	for _, version := range installed {
		names = append(names, DisplayVersion(version.Version))
	}
	return "", fmt.Errorf("Go %s is not installed. Installed: %s", DisplayVersion(input), strings.Join(names, ", "))
}
