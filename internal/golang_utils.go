package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
)

// goVersionRegex parses the output of "go version" command.
var goVersionRegex = regexp.MustCompile(`go version go(\S+)`)

// GetCurrentGolangVersion returns the currently active Go version.
// It first tries the system Go in PATH, then falls back to /usr/local/go/bin/go.
func GetCurrentGolangVersion() (*string, error) {
	var res []byte
	var err error

	res, err = exec.Command("go", "version").Output()
	if err == nil {
		return parseGoVersionOutput(string(res))
	}

	goBinary := filepath.Join("/usr", "local", "go", "bin", "go")
	if _, statErr := os.Stat(goBinary); statErr == nil {
		res, err = exec.Command(goBinary, "version").Output()
		if err == nil {
			return parseGoVersionOutput(string(res))
		}
	}

	return nil, fmt.Errorf("go is not installed or not found in PATH")
}

// parseGoVersionOutput extracts the version string from "go version" output.
func parseGoVersionOutput(output string) (*string, error) {
	matches := goVersionRegex.FindStringSubmatch(output)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to parse go version output")
	}

	version := matches[1]
	return &version, nil
}

// IsGoInstalledAtSystemPath checks if Go is installed at /usr/local/go.
func IsGoInstalledAtSystemPath() bool {
	goBinary := filepath.Join("/usr", "local", "go", "bin", "go")
	if _, err := os.Stat(goBinary); err == nil {
		return true
	}
	return false
}

// PurgeCurrentGolangInstallation removes the Go installation from /usr/local/go.
// This is called before installing a new version to ensure a clean state.
func PurgeCurrentGolangInstallation() error {
	pathToDelete := path.Join("/usr", "local", "go")

	entries, err := os.ReadDir(pathToDelete)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	version, err := GetCurrentGolangVersion()
	if err != nil {
		fmt.Printf("Warning: could not determine current Go version: %v\n", err)
	} else {
		fmt.Printf("Removing Go version %s\n", *version)
	}

	if err := os.RemoveAll(pathToDelete); err != nil {
		return fmt.Errorf("failed to purge current golang binary at path %s: %w", pathToDelete, err)
	}

	return nil
}
