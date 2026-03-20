// Package internal provides core functionality for GVM (Go Version Manager).
// It handles configuration management, version fetching, downloading, and
// system utilities for managing multiple Go installations.
package internal

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// versionRegex validates Go version strings in formats like: 1.21.0, 1.22.0-rc1, v1.21.5
var versionRegex = regexp.MustCompile(`^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)(?:\.(?P<patch>0|[1-9]\d*))?(?:(?P<rc>rc[1-9]\d*))?$`)

// ExecShellCommand executes a shell command with the given arguments and returns the output.
// It provides a safe alternative to shell string parsing by accepting command and args separately.
//
// Example:
//
//	output, err := ExecShellCommand("tar", "-C", "/usr/local", "-xzf", "file.tar.gz")
func ExecShellCommand(cmd string, args ...string) ([]byte, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("Cmd Error: Invalid shell command provided")
	}

	command := exec.Command(cmd, args...)
	out, err := command.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

// ValidateGoVersion checks if a string represents a valid Go version.
// Supports formats: 1.21.0, v1.21.0, 1.22.0-rc1, etc.
//
// Returns true if the version format is valid, false otherwise.
func ValidateGoVersion(version string) bool {
	return versionRegex.MatchString(version)
}

// NormalizeVersion removes the "go" prefix from a version string if present.
// This is useful for comparing user input (e.g., "1.22.0") with stored versions
// (e.g., "go1.22.0").
//
// Example:
//
//	NormalizeVersion("go1.22.0") // returns "1.22.0"
//	NormalizeVersion("1.22.0")    // returns "1.22.0"
func NormalizeVersion(version string) string {
	return strings.TrimPrefix(version, "go")
}
