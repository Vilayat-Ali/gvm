package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CreateSystemDirectory attempts to create directory in /usr/local/
// with fallback to user directory if permission denied
func CreateSystemDirectory(dirName string) (string, error) {
	systemPath := filepath.Join("/usr/local", dirName)

	// Try to create with current user permissions
	err := os.MkdirAll(systemPath, 0755)
	if err == nil {
		return systemPath, nil
	}

	// If permission denied, try sudo
	if os.IsPermission(err) {
		fmt.Printf("Permission denied for %s\n", systemPath)
		fmt.Println("Attempting with sudo...")

		// Try sudo mkdir
		cmd := exec.Command("sudo", "mkdir", "-p", systemPath)
		if sudoErr := cmd.Run(); sudoErr != nil {
			// If sudo fails, use user directory
			return CreateUserDirectory(dirName)
		}

		// Set ownership to current user
		user := os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME")
		}
		if user != "" {
			cmd = exec.Command("sudo", "chown", "-R", user+":"+user, systemPath)
			cmd.Run() // Ignore error, not critical
		}

		return systemPath, nil
	}

	// Other error
	return "", fmt.Errorf("failed to create %s: %w", systemPath, err)
}

// CreateUserDirectory creates directory in user's home
func CreateUserDirectory(dirName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	userPath := filepath.Join(homeDir, ".local", dirName)
	if err := os.MkdirAll(userPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create %s: %w", userPath, err)
	}

	fmt.Printf("Using user directory: %s\n", userPath)
	return userPath, nil
}
