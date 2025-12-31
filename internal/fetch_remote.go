package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"

	progressbar "github.com/schollz/progressbar/v3"
)

// Version metadata for remote versions for download
type RemoteVersion struct {
	Version      string `json:"version"`
	DownloadLink string `json:"download_link"`
}

func (rv *RemoteVersion) Download() (*string, error) {
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

	filePath := filepath.Join(downloadDirPath, fmt.Sprintf("%s.tar.gz", rv.Version))

	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	totalSize := resp.ContentLength
	progress := progressbar.DefaultBytes(totalSize, "downloading")

	if _, err := io.Copy(io.MultiWriter(out, progress), resp.Body); err != nil {
		color.Red(err.Error())
	}

	return &filePath, nil
}

// Golang release url
const GO_GITHUB_RELEASE_URL = "https://github.com/golang/go/tags"

// This function downloads the HTML of official Golang release page on
// their github @ "https://github.com/golang/go/tags"
// Then it uses goquery to parse HTML and fetch 10 most recent golang
// versions
func FetchGoVersionsFromGoGithubRelease() ([]RemoteVersion, error) {
	response, err := http.Get(GO_GITHUB_RELEASE_URL)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to fetch go version from github releases. Status Code: %d. Status: %s", response.StatusCode, response.Status)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	releaseTableRows := doc.Find("div.Box-body").Children()
	if releaseTableRows == nil {
		return nil, fmt.Errorf("failed to fetch releases from official github releases. Check UI for diff.")
	}

	releases := make([]RemoteVersion, 10)

	for idx, release_row := range releaseTableRows.EachIter() {
		version_name_selection := release_row.Find("a.Link--primary")
		if version_name_selection == nil {
			return nil, fmt.Errorf("failed to parse selection for version name from github. Diff UI")
		}

		version_download_link := release_row.Find("a[href*='.tar.gz']")
		if version_download_link == nil {
			return nil, fmt.Errorf("failed to parse selection for version download link from github. Diff UI")
		}

		version := version_name_selection.Text()
		downloadLink, hrefExists := version_download_link.Attr("href")

		if !hrefExists {
			return nil, fmt.Errorf("failed to parse download link from github. Diff UI")
		}

		if version == "" || downloadLink == "" {
			return nil, fmt.Errorf("failed to parse values for version or download link from github. Diff UI")
		}

		releases[idx] = RemoteVersion{
			Version:      version,
			DownloadLink: fmt.Sprintf("https://github.com%s", downloadLink),
		}
	}

	return releases, nil
}
