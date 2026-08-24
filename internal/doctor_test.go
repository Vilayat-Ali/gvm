package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSwitchesToolchain(t *testing.T) {
	cases := []struct {
		name              string
		active, mod, tool string
		want              bool
	}{
		{"module needs newer go", "go1.23.12", "go1.25.5", "auto", true},
		{"module needs older go", "go1.25.5", "go1.23.0", "auto", false},
		{"same version", "go1.25.5", "go1.25.5", "auto", false},
		{"toolchain pinned to local", "go1.23.12", "go1.25.5", "local", false},
		{"no active version", "", "go1.25.5", "auto", false},
		{"no module", "go1.23.12", "", "auto", false},
		{"minor only", "go1.23", "go1.24", "auto", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := switchesToolchain(c.active, c.mod, c.tool); got != c.want {
				t.Errorf("switchesToolchain(%q, %q, %q) = %v, want %v", c.active, c.mod, c.tool, got, c.want)
			}
		})
	}
}

func TestGoDirectiveOf(t *testing.T) {
	dir := t.TempDir()

	cases := map[string]string{
		"module example.com/x\n\ngo 1.25.5\n":                        "go1.25.5",
		"module example.com/x\ngo 1.24\n":                            "go1.24",
		"module example.com/x\ntoolchain go1.26.0\ngo 1.25.0\n":      "go1.25.0",
		"module example.com/x\n":                                     "",
		"module example.com/x\ngo not-a-version\n":                   "",
		"module example.com/x\nrequire (\n\tgo.uber.org/zap v1.0\n)": "",
	}
	for content, want := range cases {
		path := filepath.Join(dir, "go.mod")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := goDirectiveOf(path); got != want {
			t.Errorf("goDirectiveOf(%q) = %q, want %q", content, got, want)
		}
	}

	if got := goDirectiveOf(filepath.Join(dir, "missing.mod")); got != "" {
		t.Errorf("goDirectiveOf on a missing file = %q, want empty", got)
	}
}

func TestFindGoMod(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if found := findGoMod(nested); found != "" && filepath.Dir(found) == root {
		t.Fatal("found a go.mod before one was written")
	}

	modPath := filepath.Join(root, "go.mod")
	if err := os.WriteFile(modPath, []byte("module example.com/x\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if found := findGoMod(nested); found != modPath {
		t.Errorf("findGoMod(%q) = %q, want %q", nested, found, modPath)
	}
}
