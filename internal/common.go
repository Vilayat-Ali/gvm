package internal

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Utility function to execute shell command and retrieve the shell output.
func ExecShellCommand(cmd string) ([]byte, error) {
	cmdParts := strings.Split(cmd, " ")

	if len(cmdParts) == 0 {
		return nil, fmt.Errorf("Cmd Error: Invalid shell command provided: %s", cmd)
	}

	fmt.Println(cmdParts)

	var command *exec.Cmd

	if len(cmdParts) == 1 {
		command = exec.Command(cmdParts[0])
	} else {
		command = exec.Command(cmdParts[0], cmdParts[1:]...)
	}

	out, err := command.Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Utility function to validate golang versions as strings
func ValidateGoVersion(version string) bool {
	re := regexp.MustCompile(`^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)(?:\.(?P<patch>0|[1-9]\d*))?(?:(?P<rc>rc[1-9]\d*))?$`)
	return re.MatchString(version)
}
