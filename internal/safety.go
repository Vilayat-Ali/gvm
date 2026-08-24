package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var protectedPaths = map[string]bool{
	"/": true, "/bin": true, "/boot": true, "/dev": true, "/etc": true,
	"/home": true, "/lib": true, "/lib32": true, "/lib64": true, "/media": true,
	"/mnt": true, "/opt": true, "/proc": true, "/root": true, "/run": true,
	"/sbin": true, "/srv": true, "/sys": true, "/tmp": true, "/usr": true,
	"/usr/bin": true, "/usr/lib": true, "/usr/lib64": true, "/usr/local": true,
	"/usr/local/bin": true, "/usr/local/go": true, "/usr/local/lib": true,
	"/usr/local/sbin": true, "/usr/local/share": true, "/usr/sbin": true,
	"/usr/share": true, "/var": true, "/Applications": true, "/Library": true,
	"/System": true, "/Users": true,
}

func IsProtectedPath(p string) bool {
	clean := filepath.Clean(p)
	if protectedPaths[clean] {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil && filepath.Clean(home) == clean {
		return true
	}
	return false
}

func WithinDir(parent, child string) bool {
	rel, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	if err != nil {
		return false
	}
	if rel == "." || rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func guardRootPath(p string) error {
	clean := filepath.Clean(p)
	if !filepath.IsAbs(clean) {
		return fmt.Errorf("gvm root must be an absolute path, got %q", p)
	}
	if IsProtectedPath(clean) {
		return fmt.Errorf("refusing to use %q as the gvm root: it is a system directory", clean)
	}
	if strings.Count(strings.Trim(clean, "/"), "/") < 1 {
		return fmt.Errorf("refusing to use %q as the gvm root: path is too close to the filesystem root", clean)
	}
	return nil
}

func guardManagedPath(target string) (string, error) {
	clean := filepath.Clean(target)
	if !filepath.IsAbs(clean) {
		return "", fmt.Errorf("refusing to touch relative path %q", target)
	}
	if IsProtectedPath(clean) {
		return "", fmt.Errorf("refusing to touch %q: it is a system directory", clean)
	}
	root, err := Root()
	if err != nil {
		return "", err
	}
	if !WithinDir(root, clean) {
		return "", fmt.Errorf("refusing to touch %q: it lives outside the gvm root %q", clean, root)
	}
	return clean, nil
}

func RemoveManaged(target string) error {
	clean, err := guardManagedPath(target)
	if err != nil {
		return err
	}
	info, err := os.Lstat(clean)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return os.Remove(clean)
	}
	return os.RemoveAll(clean)
}
