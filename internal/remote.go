package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	defaultAPIURL      = "https://go.dev/dl/?mode=json&include=all"
	defaultDownloadURL = "https://go.dev/dl/"
	maxAPIBodyBytes    = 16 << 20
)

type RemoteVersion struct {
	Version      string `json:"version"`
	Stable       bool   `json:"stable"`
	Filename     string `json:"filename"`
	DownloadLink string `json:"download_link"`
	SHA256       string `json:"sha256"`
	Size         int64  `json:"size"`
}

type goFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

type goRelease struct {
	Version string   `json:"version"`
	Stable  bool     `json:"stable"`
	Files   []goFile `json:"files"`
}

var apiClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          8,
		IdleConnTimeout:       30 * time.Second,
	},
}

var downloadClient = &http.Client{
	Transport: &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		IdleConnTimeout:       60 * time.Second,
	},
}

func apiURL() string {
	if custom := strings.TrimSpace(os.Getenv("GVM_API_URL")); custom != "" {
		return custom
	}
	return defaultAPIURL
}

func downloadBase() string {
	if custom := strings.TrimSpace(os.Getenv("GVM_DOWNLOAD_URL")); custom != "" {
		return strings.TrimSuffix(custom, "/") + "/"
	}
	return defaultDownloadURL
}

func PlatformOSArch() (string, string) {
	goos, goarch := runtime.GOOS, runtime.GOARCH
	if goos == "linux" && goarch == "arm" {
		goarch = "armv6l"
	}
	return goos, goarch
}

func httpGet(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	var lastErr error
	for attempt := range 3 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(1<<attempt) * 400 * time.Millisecond):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", AppName+"/"+AppVersion)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		lastErr = fmt.Errorf("%s returned %s: %s", url, resp.Status, strings.TrimSpace(string(body)))
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return nil, lastErr
		}
	}
	return nil, lastErr
}

func FetchRemoteVersions(ctx context.Context) ([]RemoteVersion, error) {
	resp, err := httpGet(ctx, apiClient, apiURL())
	if err != nil {
		return nil, fmt.Errorf("cannot reach the Go release index: %w", err)
	}
	defer resp.Body.Close()

	var releases []goRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxAPIBodyBytes)).Decode(&releases); err != nil {
		return nil, fmt.Errorf("cannot parse the Go release index: %w", err)
	}

	goos, goarch := PlatformOSArch()
	versions := make([]RemoteVersion, 0, len(releases))

	for _, release := range releases {
		if !IsValidVersion(release.Version) {
			continue
		}
		for _, file := range release.Files {
			if file.Kind != "archive" || file.OS != goos || file.Arch != goarch {
				continue
			}
			if !strings.HasSuffix(file.Filename, ".tar.gz") && !strings.HasSuffix(file.Filename, ".zip") {
				continue
			}
			if strings.ContainsAny(file.Filename, "/\\") || file.SHA256 == "" {
				continue
			}
			versions = append(versions, RemoteVersion{
				Version:      release.Version,
				Stable:       release.Stable,
				Filename:     file.Filename,
				DownloadLink: downloadBase() + file.Filename,
				SHA256:       file.SHA256,
				Size:         file.Size,
			})
			break
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no Go builds published for %s/%s", goos, goarch)
	}

	SortVersionsDesc(versions)
	return versions, nil
}

func SortVersionsDesc(versions []RemoteVersion) {
	parsedCache := make(map[string]ParsedVersion, len(versions))
	parse := func(v string) ParsedVersion {
		if p, ok := parsedCache[v]; ok {
			return p
		}
		p, _ := ParseVersion(v)
		parsedCache[v] = p
		return p
	}
	sort.SliceStable(versions, func(i, j int) bool {
		return CompareVersions(parse(versions[i].Version), parse(versions[j].Version)) > 0
	})
}
