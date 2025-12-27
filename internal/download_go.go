package internal

// Represents the metadata for downloaded and locally
// saved versions of golang of gvm
type DownloadVersion struct {
	// version of golang
	Version string `json:"version"`
	// path to golang version binary. Path to /go/bin
	BinPath string `json:"bin_path"`
}

func DownloadGoVersion(version string, path string) error {
	return nil
}

func ValidateDownloadCheckSum(version string, path string) (bool, error) {
	return true, nil
}
