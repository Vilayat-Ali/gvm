package internal

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Represents the metadata for downloaded and locally
// saved versions of golang of gvm
type DownloadVersion struct {
	// version of golang
	Version string `json:"version"`
	// path to downloaded golang tarball.
	TarPath string `json:"tar_path"`
}

func (dv *DownloadVersion) GetDecompressedDirName() string {
	filename := strings.Replace(filepath.Base(dv.TarPath), ".tar.gz", "", 1)
	return fmt.Sprintf("go-%s", filename)
}

func DownloadGoVersion(version string, path string) error {
	return nil
}

func ValidateDownloadCheckSum(version string, path string) (bool, error) {
	return true, nil
}
