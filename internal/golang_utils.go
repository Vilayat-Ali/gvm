package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fatih/color"
)

// Fetches current golang version from CMD
// It uses `go version` command.
func GetCurrentGolangVersion() (*string, error) {
	res, err := exec.Command("go", "version").Output()
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to fetch go version")
	}

	return &strings.Split(string(res), " ")[2], nil
}

func PurgeCurrentGolangInstallation() {
	pathToDelete := path.Join("/usr", "local", "go")

	if _, err := os.ReadDir(pathToDelete); os.IsNotExist(err) {
		return
	}

	version, err := GetCurrentGolangVersion()
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}

	if err := os.RemoveAll(pathToDelete); err != nil {
		color.Red(fmt.Sprintf("IO Error: Failed to purge current golang binary at path: %s", pathToDelete))
		os.Exit(1)
	}

	color.Green(fmt.Sprintf("Successfully removed current golang version %s", *version))
}
