package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandHelp(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"--help"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"configure", "download", "use", "list", "remove", "doctor", "env"} {
		if !strings.Contains(out.String(), name) {
			t.Errorf("`gvm --help` does not mention the %q command", name)
		}
	}
}

func TestEveryCommandValidatesArgs(t *testing.T) {
	for _, command := range rootCmd.Commands() {
		if command.Name() == "help" || command.Name() == "completion" {
			continue
		}
		if command.Args == nil && command.RunE != nil {
			t.Errorf("command %q does not validate its arguments", command.Name())
		}
	}
}

func TestUnknownVersionIsRejectedBeforeAnyWork(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GVM_ROOT", root)
	t.Setenv("XDG_CONFIG_HOME", root+"/config")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"use", "not-a-version"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err == nil {
		t.Fatal("`gvm use not-a-version` should fail")
	}
}
