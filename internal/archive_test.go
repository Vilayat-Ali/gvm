package internal

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

type entry struct {
	name     string
	typeflag byte
	linkname string
	body     string
	mode     int64
}

func buildArchive(t *testing.T, path string, entries []entry) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)

	for _, e := range entries {
		mode := e.mode
		if mode == 0 {
			mode = 0o644
		}
		header := &tar.Header{
			Name:     e.name,
			Typeflag: e.typeflag,
			Linkname: e.linkname,
			Mode:     mode,
			Size:     int64(len(e.body)),
		}
		if e.typeflag == tar.TypeDir {
			header.Size = 0
			header.Mode = 0o755
		}
		if e.typeflag != tar.TypeReg {
			header.Size = 0
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if e.typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

func setupExtract(t *testing.T, entries []entry) (string, string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv(EnvRoot, root)

	archive := filepath.Join(root, "archive.tar.gz")
	buildArchive(t, archive, entries)
	return archive, filepath.Join(root, "versions", "go1.22.0")
}

func TestExtractStripsLeadingComponent(t *testing.T) {
	archive, dest := setupExtract(t, []entry{
		{name: "go/", typeflag: tar.TypeDir},
		{name: "go/bin/", typeflag: tar.TypeDir},
		{name: "go/bin/go", typeflag: tar.TypeReg, body: "binary", mode: 0o755},
		{name: "go/VERSION", typeflag: tar.TypeReg, body: "go1.22.0"},
	})

	if err := ExtractTarGz(archive, dest, nil); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dest, "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "go1.22.0" {
		t.Errorf("VERSION = %q", data)
	}

	info, err := os.Stat(filepath.Join(dest, "bin", "go"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Error("the go binary is not executable after extraction")
	}
}

func TestExtractRejectsPathTraversal(t *testing.T) {
	archive, dest := setupExtract(t, []entry{
		{name: "go/", typeflag: tar.TypeDir},
		{name: "go/../../../../../../tmp/gvm-escape", typeflag: tar.TypeReg, body: "pwned"},
		{name: "go/VERSION", typeflag: tar.TypeReg, body: "go1.22.0"},
	})

	if err := ExtractTarGz(archive, dest, nil); err != nil {
		t.Fatalf("extraction should skip the hostile entry, not fail: %v", err)
	}
	if _, err := os.Stat("/tmp/gvm-escape"); err == nil {
		os.Remove("/tmp/gvm-escape")
		t.Fatal("a traversal entry escaped the destination directory")
	}
	if _, err := os.Stat(filepath.Join(dest, "VERSION")); err != nil {
		t.Fatal("the safe entries were not extracted")
	}
}

func TestExtractRejectsAbsolutePaths(t *testing.T) {
	archive, dest := setupExtract(t, []entry{
		{name: "/etc/gvm-escape", typeflag: tar.TypeReg, body: "pwned"},
		{name: "go/VERSION", typeflag: tar.TypeReg, body: "go1.22.0"},
	})

	if err := ExtractTarGz(archive, dest, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("/etc/gvm-escape"); err == nil {
		t.Fatal("an absolute path entry was written outside the destination")
	}
}

func TestExtractRejectsEscapingSymlinks(t *testing.T) {
	archive, dest := setupExtract(t, []entry{
		{name: "go/", typeflag: tar.TypeDir},
		{name: "go/evil", typeflag: tar.TypeSymlink, linkname: "../../../../etc"},
		{name: "go/absolute", typeflag: tar.TypeSymlink, linkname: "/etc/passwd"},
		{name: "go/safe", typeflag: tar.TypeSymlink, linkname: "VERSION"},
		{name: "go/VERSION", typeflag: tar.TypeReg, body: "go1.22.0"},
	})

	if err := ExtractTarGz(archive, dest, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(dest, "evil")); err == nil {
		t.Error("an escaping symlink was created")
	}
	if _, err := os.Lstat(filepath.Join(dest, "absolute")); err == nil {
		t.Error("an absolute symlink was created")
	}
	if _, err := os.Lstat(filepath.Join(dest, "safe")); err != nil {
		t.Error("a safe relative symlink was skipped")
	}
}

func TestExtractSkipsSpecialFiles(t *testing.T) {
	archive, dest := setupExtract(t, []entry{
		{name: "go/", typeflag: tar.TypeDir},
		{name: "go/device", typeflag: tar.TypeChar},
		{name: "go/pipe", typeflag: tar.TypeFifo},
		{name: "go/VERSION", typeflag: tar.TypeReg, body: "go1.22.0"},
	})

	if err := ExtractTarGz(archive, dest, nil); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"device", "pipe"} {
		if _, err := os.Lstat(filepath.Join(dest, name)); err == nil {
			t.Errorf("%s should have been skipped", name)
		}
	}
}

func TestExtractRefusesDestinationOutsideRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRoot, root)

	archive := filepath.Join(root, "a.tar.gz")
	buildArchive(t, archive, []entry{{name: "go/VERSION", typeflag: tar.TypeReg, body: "x"}})

	outside := filepath.Join(t.TempDir(), "dest")
	if err := ExtractTarGz(archive, outside, nil); err == nil {
		t.Fatal("extraction was allowed outside the gvm root")
	}
	if err := ExtractTarGz(archive, "/usr/local/go", nil); err == nil {
		t.Fatal("extraction into /usr/local/go was allowed")
	}
}

func TestExtractRejectsCorruptArchive(t *testing.T) {
	root := t.TempDir()
	t.Setenv(EnvRoot, root)

	archive := filepath.Join(root, "broken.tar.gz")
	if err := os.WriteFile(archive, []byte("this is not a gzip stream"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ExtractTarGz(archive, filepath.Join(root, "versions", "go1.22.0"), nil); err == nil {
		t.Fatal("a corrupt archive was accepted")
	}
}
