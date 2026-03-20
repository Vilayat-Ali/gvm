package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"

	progressbar "github.com/schollz/progressbar/v3"
)

// RemoteVersion represents a Go version available for download from go.dev.
type RemoteVersion struct {
	Version      string `json:"version"`
	DownloadLink string `json:"download_link"`
}

// Download downloads the Go version tarball to the local storage directory.
// Shows progress, verifies checksum after download, and removes the file if verification fails.
func (rv *RemoteVersion) Download() (*string, error) {
	goos, goarch := getOSAndArch()

	faint := color.New(color.Faint)
	faint.Printf("  Platform: %s/%s\n", goos, goarch)
	faint.Println(fmt.Sprintf("  Source: %s", rv.DownloadLink))

	resp, err := http.Get(rv.DownloadLink)
	if err != nil {
		return nil, fmt.Errorf("download error (%s): %w", rv.Version, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed (%s): %s", rv.Version, resp.Status)
	}

	downloadDirPath, err := GoDownloadDir()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(*downloadDirPath, fmt.Sprintf("%s.tar.gz", rv.Version))

	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	totalSize := resp.ContentLength
	progress := progressbar.DefaultBytes(totalSize, "  downloading...")

	if _, err := io.Copy(io.MultiWriter(out, progress), resp.Body); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	faint.Println("  Verifying integrity...")
	valid, err := rv.VerifyChecksum(filePath)
	if err != nil {
		color.Yellow("  ⚠️  Warning: Could not verify checksum: %s", err.Error())
	} else if !valid {
		os.Remove(filePath)
		return nil, fmt.Errorf("checksum verification failed for %s", rv.Version)
	} else {
		color.Green("  ✓ Checksum verified")
	}

	return &filePath, nil
}

// GO_GITHUB_RELEASE_URL is the GitHub page to scrape for Go release information.
const GO_GITHUB_RELEASE_URL = "https://github.com/golang/go/tags"

// FetchGoVersionsFromGoGithubRelease scrapes the GitHub releases page and returns
// the 10 most recent Go versions available for download. Constructs download URLs
// based on the current platform (OS and architecture).
func FetchGoVersionsFromGoGithubRelease() ([]RemoteVersion, error) {
	response, err := http.Get(GO_GITHUB_RELEASE_URL)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch releases. HTTP %d: %s", response.StatusCode, response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	releaseTableRows := doc.Find("div.Box-body").Children()
	if releaseTableRows == nil {
		return nil, fmt.Errorf("failed to fetch releases from GitHub - page structure may have changed")
	}

	releases := make([]RemoteVersion, 10)
	goos, goarch := getOSAndArch()

	for idx, release_row := range releaseTableRows.EachIter() {
		version_name_selection := release_row.Find("a.Link--primary")
		if version_name_selection == nil {
			return nil, fmt.Errorf("failed to parse version name from GitHub")
		}

		version_download_link := release_row.Find("a[href*='.tar.gz']")
		if version_download_link == nil {
			return nil, fmt.Errorf("failed to parse download link from GitHub")
		}

		version := version_name_selection.Text()
		downloadLink, hrefExists := version_download_link.Attr("href")

		if !hrefExists {
			return nil, fmt.Errorf("failed to parse download link")
		}

		if version == "" || downloadLink == "" {
			return nil, fmt.Errorf("failed to parse version or download link")
		}

		scrappedDownloadLinkParts := strings.Split(downloadLink, "/")
		scrappedFileName := scrappedDownloadLinkParts[len(scrappedDownloadLinkParts)-1]
		baseVersion := strings.Replace(scrappedFileName, ".tar.gz", "", 1)
		downloadableBinaryName := fmt.Sprintf("%s.%s-%s.tar.gz", baseVersion, goos, goarch)

		releases[idx] = RemoteVersion{
			Version:      version,
			DownloadLink: fmt.Sprintf("https://go.dev/dl/%s", downloadableBinaryName),
		}
	}

	return releases, nil
}
