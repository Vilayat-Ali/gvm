package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsProtectedPath(t *testing.T) {
	protected := []string{
		"/", "/usr", "/usr/local", "/usr/local/go", "/usr/local/bin", "/etc",
		"/bin", "/var", "/home", "/boot", "/usr/local/", "/usr/local/go/",
	}
	for _, p := range protected {
		if !IsProtectedPath(p) {
			t.Errorf("%q should be protected", p)
		}
	}

	if home, err := os.UserHomeDir(); err == nil && !IsProtectedPath(home) {
		t.Errorf("the home directory %q should be protected", home)
	}

	allowed := []string{"/usr/local/gvm/versions", "/opt/gvm", "/home/someone/.local/share/gvm"}
	for _, p := range allowed {
		if IsProtectedPath(p) {
			t.Errorf("%q should not be protected", p)
		}
	}
}

func TestWithinDir(t *testing.T) {
	cases := []struct {
		parent, child string
		want          bool
	}{
		{"/a/b", "/a/b/c", true},
		{"/a/b", "/a/b/c/d", true},
		{"/a/b", "/a/b", false},
		{"/a/b", "/a", false},
		{"/a/b", "/a/c", false},
		{"/a/b", "/a/b/../../etc", false},
		{"/a/b", "/a/bc", false},
	}
	for _, c := range cases {
		if got := WithinDir(c.parent, c.child); got != c.want {
			t.Errorf("WithinDir(%q, %q) = %v, want %v", c.parent, c.child, got, c.want)
		}
	}
}

func TestRootRejectsSystemDirectories(t *testing.T) {
	for _, p := range []string{"/", "/usr", "/usr/local", "/usr/local/go", "/etc", "/var"} {
		t.Setenv(EnvRoot, p)
		if _, err := Root(); err == nil {
			t.Errorf("Root() accepted the system directory %q", p)
		}
	}
}

func TestRemoveManagedRefusesOutsideRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRoot, root)

	outside := filepath.Join(t.TempDir(), "keep.txt")
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveManaged(outside); err == nil {
		t.Fatal("RemoveManaged deleted a path outside the gvm root")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("the outside file was removed: %v", err)
	}

	for _, p := range []string{"/usr/local/go", "/etc", "/", "relative/path"} {
		if err := RemoveManaged(p); err == nil {
			t.Errorf("RemoveManaged accepted %q", p)
		}
	}
}

func TestRemoveManagedInsideRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRoot, root)

	victim := filepath.Join(root, "versions", "go1.22.0")
	if err := os.MkdirAll(victim, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(victim, "file"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveManaged(victim); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(victim); !os.IsNotExist(err) {
		t.Fatal("the managed directory was not removed")
	}
	if err := RemoveManaged(victim); err != nil {
		t.Fatalf("removing a missing path should succeed: %v", err)
	}
}

func TestRemoveManagedFollowsNoSymlinks(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRoot, root)

	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "precious"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "current")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := RemoveManaged(link); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "precious")); err != nil {
		t.Fatal("RemoveManaged followed a symlink and deleted the target contents")
	}
}
