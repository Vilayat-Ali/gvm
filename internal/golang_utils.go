package internal

import (
	"fmt"
	"os/exec"
	"strings"
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
