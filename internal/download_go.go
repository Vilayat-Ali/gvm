package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// getOSAndArch returns the current operating system and architecture.
// Used to download the correct Go binary for the user's platform.
func getOSAndArch() (goos, goarch string) {
	return runtime.GOOS, runtime.GOARCH
}

// DownloadVersion represents a downloaded Go version stored locally.
type DownloadVersion struct {
	Version string `json:"version"`
	TarPath string `json:"tar_path"`
}

// GetDecompressedDirName returns the directory name that would be created
// when the tarball is extracted (e.g., "go-go1.22.0").
func (dv *DownloadVersion) GetDecompressedDirName() string {
	filename := strings.Replace(filepath.Base(dv.TarPath), ".tar.gz", "", 1)
	return fmt.Sprintf("go-%s", filename)
}

// GoVersionInfo represents version metadata from go.dev's JSON API.
type GoVersionInfo struct {
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	Size     string `json:"size"`
	Sha256   string `json:"sha256"`
	Kind     string `json:"kind"`
}

// ValidateDownloadCheckSum verifies that a downloaded file matches its expected checksum.
// Returns true if the checksum matches, false otherwise.
func ValidateDownloadCheckSum(version string, tarPath string) (bool, error) {
	checksum, err := fetchChecksum(version)
	if err != nil {
		return false, fmt.Errorf("failed to fetch checksum: %w", err)
	}

	file, err := os.Open(tarPath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("failed to compute hash: %w", err)
	}

	computedChecksum := hex.EncodeToString(hash.Sum(nil))
	return computedChecksum == checksum, nil
}

// fetchChecksum retrieves the SHA256 checksum for a specific Go version
// from go.dev's JSON API, matching the current platform.
func fetchChecksum(version string) (string, error) {
	resp, err := http.Get("https://go.dev/dl/?mode=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var versions []GoVersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	goos, goarch := getOSAndArch()
	versionName := strings.TrimPrefix(version, "go")

	for _, v := range versions {
		if v.Version == versionName && v.OS == goos && v.Arch == goarch && strings.HasSuffix(v.Filename, ".tar.gz") {
			return v.Sha256, nil
		}
	}

	return "", fmt.Errorf("checksum not found for version %s", version)
}

// GetChecksum retrieves the checksum for a RemoteVersion.
func (rv *RemoteVersion) GetChecksum() (string, error) {
	versionName := strings.TrimPrefix(rv.Version, "go")
	checksum, err := fetchChecksum(versionName)
	if err != nil {
		return "", err
	}
	return checksum, nil
}

// VerifyChecksum verifies that a downloaded tarball matches the expected checksum.
func (rv *RemoteVersion) VerifyChecksum(tarPath string) (bool, error) {
	checksum, err := rv.GetChecksum()
	if err != nil {
		return false, err
	}

	file, err := os.Open(tarPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, err
	}

	computedChecksum := hex.EncodeToString(hash.Sum(nil))
	return computedChecksum == checksum, nil
}

// GetDownloadDir returns the directory where Go versions are stored.
func GetDownloadDir() (string, error) {
	goDir, err := GoDownloadDir()
	if err != nil {
		return "", err
	}
	return *goDir, nil
}

// GetTarballPath returns the expected path for a version's tarball.
func GetTarballPath(version string) string {
	downloadDir, _ := GetDownloadDir()
	return filepath.Join(downloadDir, fmt.Sprintf("%s.tar.gz", version))
}

type DownloadVersion struct {
	Version string `json:"version"`
	TarPath string `json:"tar_path"`
}

func (dv *DownloadVersion) GetDecompressedDirName() string {
	filename := strings.Replace(filepath.Base(dv.TarPath), ".tar.gz", "", 1)
	return fmt.Sprintf("go-%s", filename)
}

type GoVersionInfo struct {
	Version   string `json:"version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Filename  string `json:"filename"`
	Size      string `json:"size"`
	Sha256    string `json:"sha256"`
	Kind      string `json:"kind"`
	Checksum  string `json:"checksum,omitempty"`
}

func DownloadGoVersion(version string, path string) error {
	return nil
}

func ValidateDownloadCheckSum(version string, tarPath string) (bool, error) {
	checksum, err := fetchChecksum(version)
	if err != nil {
		return false, fmt.Errorf("failed to fetch checksum: %w", err)
	}

	file, err := os.Open(tarPath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("failed to compute hash: %w", err)
	}

	computedChecksum := hex.EncodeToString(hash.Sum(nil))
	return computedChecksum == checksum, nil
}

func fetchChecksum(version string) (string, error) {
	resp, err := http.Get("https://go.dev/dl/?mode=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var versions []GoVersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	userMachineOS, userMachineArch := getOSAndArch()

	versionName := strings.TrimPrefix(version, "go")
	for _, v := range versions {
		if v.Version == versionName && v.OS == userMachineOS && v.Arch == userMachineArch && strings.HasSuffix(v.Filename, ".tar.gz") {
			return v.Sha256, nil
		}
	}

	return "", fmt.Errorf("checksum not found for version %s", version)
}

func (rv *RemoteVersion) GetChecksum() (string, error) {
	versionName := strings.TrimPrefix(rv.Version, "go")
	checksum, err := fetchChecksum(versionName)
	if err != nil {
		return "", err
	}
	return checksum, nil
}

func (rv *RemoteVersion) VerifyChecksum(tarPath string) (bool, error) {
	checksum, err := rv.GetChecksum()
	if err != nil {
		return false, err
	}

	file, err := os.Open(tarPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, err
	}

	computedChecksum := hex.EncodeToString(hash.Sum(nil))
	return computedChecksum == checksum, nil
}

func GetDownloadDir() (string, error) {
	goDir, err := GoDownloadDir()
	if err != nil {
		return "", err
	}
	return *goDir, nil
}

func GetTarballPath(version string) string {
	downloadDir, _ := GetDownloadDir()
	return filepath.Join(downloadDir, fmt.Sprintf("%s.tar.gz", version))
}
