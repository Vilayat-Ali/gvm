package internal

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	AppName   = "gvm"
	EnvRoot   = "GVM_ROOT"
	EnvNoRoot = "GVM_ALLOW_ROOT"
)

var (
	AppVersion = "dev"
	BuildTime  = "unknown"
	GitCommit  = "unknown"
)

func HomeDir() (string, error) {
	if os.Geteuid() == 0 {
		if name := os.Getenv("SUDO_USER"); name != "" && name != "root" {
			if u, err := user.Lookup(name); err == nil && u.HomeDir != "" {
				return u.HomeDir, nil
			}
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return home, nil
}

func Root() (string, error) {
	if custom := strings.TrimSpace(os.Getenv(EnvRoot)); custom != "" {
		abs, err := filepath.Abs(custom)
		if err != nil {
			return "", fmt.Errorf("invalid %s: %w", EnvRoot, err)
		}
		if err := guardRootPath(abs); err != nil {
			return "", err
		}
		return abs, nil
	}

	if data := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); data != "" && filepath.IsAbs(data) {
		return filepath.Join(data, AppName), nil
	}

	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", AppName), nil
}

func VersionsDir() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "versions"), nil
}

func CacheDir() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "cache"), nil
}

func ShimDir() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "bin"), nil
}

func CurrentLink() (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "current"), nil
}

func VersionDir(version string) (string, error) {
	canonical, err := CanonicalVersion(version)
	if err != nil {
		return "", err
	}
	versions, err := VersionsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(versions, canonical), nil
}

func ConfigDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); dir != "" && filepath.IsAbs(dir) {
		return filepath.Join(dir, AppName), nil
	}
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", AppName), nil
}

func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func EnsureLayout() error {
	root, err := Root()
	if err != nil {
		return err
	}
	if err := guardRootPath(root); err != nil {
		return err
	}

	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	dirs := []string{root, filepath.Join(root, "versions"), filepath.Join(root, "cache"), filepath.Join(root, "bin"), configDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}
	return nil
}

func RunningAsRoot() bool {
	return runtime.GOOS != "windows" && os.Geteuid() == 0
}

func RootWarning() string {
	if !RunningAsRoot() || os.Getenv(EnvNoRoot) == "1" {
		return ""
	}
	return "gvm does not need root; it installs into your home directory."
}

func GuardRoot() error {
	return guardRootFor(os.Geteuid(), os.Getenv("SUDO_USER"), os.Getenv(EnvNoRoot))
}

func guardRootFor(euid int, sudoUser, override string) error {
	if runtime.GOOS == "windows" || euid != 0 || override == "1" || sudoUser == "" {
		return nil
	}
	return fmt.Errorf("refusing to run under sudo: this would leave root-owned files in %s's home directory.\n"+
		"       Run `gvm %s` without sudo, or set %s=1 to override",
		sudoUser, strings.Join(os.Args[1:], " "), EnvNoRoot)
}
