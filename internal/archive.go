package internal

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	maxArchiveBytes = 6 << 30
	maxArchiveFiles = 500000
	copyBufferSize  = 1 << 20
)

func sanitizeEntryName(name string, strip int) (string, bool) {
	cleaned := path.Clean(strings.TrimPrefix(name, "./"))
	if cleaned == "." || cleaned == "" || path.IsAbs(cleaned) || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", false
	}
	if strings.ContainsRune(cleaned, '\x00') {
		return "", false
	}
	parts := strings.Split(cleaned, "/")
	for _, part := range parts {
		if part == ".." {
			return "", false
		}
	}
	if len(parts) <= strip {
		return "", false
	}
	return filepath.Join(parts[strip:]...), true
}

func resolvesInside(dest, target string) bool {
	clean := filepath.Clean(target)
	return clean == filepath.Clean(dest) || WithinDir(dest, clean)
}

func ExtractTarGz(archivePath, dest string, onProgress func(int64)) error {
	if _, err := guardManagedPath(dest); err != nil {
		return err
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", archivePath, err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("%s is not a valid gzip archive: %w", filepath.Base(archivePath), err)
	}
	defer func() { _ = gz.Close() }()

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w", dest, err)
	}

	reader := tar.NewReader(gz)
	buffer := make([]byte, copyBufferSize)
	dirModes := make(map[string]os.FileMode)
	var written int64
	var entries int

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("corrupted archive %s: %w", filepath.Base(archivePath), err)
		}

		entries++
		if entries > maxArchiveFiles {
			return fmt.Errorf("archive %s has too many entries", filepath.Base(archivePath))
		}

		relative, ok := sanitizeEntryName(header.Name, 1)
		if !ok {
			continue
		}
		target := filepath.Join(dest, relative)
		if !resolvesInside(dest, target) {
			return fmt.Errorf("archive %s tried to write outside the install directory", filepath.Base(archivePath))
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			dirModes[target] = header.FileInfo().Mode().Perm()

		case tar.TypeReg:
			if written+header.Size > maxArchiveBytes {
				return fmt.Errorf("archive %s is larger than the %d GiB safety limit", filepath.Base(archivePath), maxArchiveBytes>>30)
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode().Perm())
			if err != nil {
				return err
			}
			copied, err := io.CopyBuffer(out, io.LimitReader(reader, header.Size), buffer)
			if closeErr := out.Close(); err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("cannot write %s: %w", relative, err)
			}
			written += copied
			if onProgress != nil {
				onProgress(copied)
			}

		case tar.TypeSymlink:
			if path.IsAbs(header.Linkname) {
				continue
			}
			resolved := filepath.Join(filepath.Dir(target), filepath.FromSlash(header.Linkname))
			if !resolvesInside(dest, resolved) {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			os.Remove(target)
			if err := os.Symlink(filepath.FromSlash(header.Linkname), target); err != nil {
				return err
			}

		case tar.TypeLink:
			source, ok := sanitizeEntryName(header.Linkname, 1)
			if !ok {
				continue
			}
			resolved := filepath.Join(dest, source)
			if !resolvesInside(dest, resolved) {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			os.Remove(target)
			if err := os.Link(resolved, target); err != nil {
				return err
			}

		default:
			continue
		}
	}

	for dir, mode := range dirModes {
		if err := os.Chmod(dir, mode); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
