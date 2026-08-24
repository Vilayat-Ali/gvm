package internal

import "testing"

func TestGuardRootFor(t *testing.T) {
	cases := []struct {
		name      string
		euid      int
		sudoUser  string
		override  string
		wantError bool
	}{
		{"normal user", 1000, "", "", false},
		{"normal user under sudo env leftovers", 1000, "someone", "", false},
		{"root in a container", 0, "", "", false},
		{"root via sudo", 0, "someone", "", true},
		{"root via sudo with override", 0, "someone", "1", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := guardRootFor(c.euid, c.sudoUser, c.override)
			if c.wantError && err == nil {
				t.Error("expected gvm to refuse to run")
			}
			if !c.wantError && err != nil {
				t.Errorf("expected gvm to run, got: %v", err)
			}
		})
	}
}

func TestLayoutStaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRoot, root)

	if err := EnsureLayout(); err != nil {
		t.Fatal(err)
	}

	for _, get := range []func() (string, error){VersionsDir, CacheDir, ShimDir, CurrentLink} {
		path, err := get()
		if err != nil {
			t.Fatal(err)
		}
		if !WithinDir(root, path) {
			t.Errorf("%s is not inside the gvm root %s", path, root)
		}
	}

	dir, err := VersionDir("1.22.0")
	if err != nil {
		t.Fatal(err)
	}
	if !WithinDir(root, dir) {
		t.Errorf("%s is not inside the gvm root", dir)
	}
	if _, err := VersionDir("../../etc"); err == nil {
		t.Error("VersionDir accepted a traversal attempt")
	}
}
