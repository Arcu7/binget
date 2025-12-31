package binary

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const (
	GithubHost   = "github.com"
	GithubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"
)

func DownloadRelease(request Request) error {
	url, err := checkRequest(request)
	if err != nil {
		return err
	}

	owner, repoName, err := getOwnerAndRepoName(url.Path)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf(GithubAPIURL, owner, repoName)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	assets, err := getLatestReleaseAssets(client, apiURL, request.AuthToken)
	if err != nil {
		return err
	}
	fmt.Printf("assets: %#v\n", assets)

	// Get correct link based on OS and Architecture
	os := runtime.GOOS
	arch := runtime.GOARCH
	fmt.Printf("Looking for OS: %s, ARCH: %s\n", os, arch)

	var link string
	for _, asset := range assets {
		lasset := strings.ToLower(asset.BrowserDownloadURL)
		if strings.Contains(lasset, os) && strings.Contains(lasset, arch) {
			link = asset.BrowserDownloadURL
		}
	}

	fmt.Printf("Download link: %s\n", link)

	return nil
}

func checkRequest(request Request) (*url.URL, error) {
	if request.RepositoryURL == "" {
		return nil, errors.New("repository URL is required")
	}

	url, err := url.ParseRequestURI(request.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}
	if url.Scheme != "https" && url.Scheme != "http" {
		return nil, errors.New("repository URL must start with http or https")
	}
	if url.Host != GithubHost {
		return nil, errors.New("repository URL must be from github.com")
	}

	return url, nil
}

func getOwnerAndRepoName(path string) (owner string, repoName string, err error) {
	// Trim the leading slash from url.Path
	trimmedPath := strings.TrimLeft(path, "/")

	parts := strings.Split(trimmedPath, "/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid repository path")
	}

	owner = parts[0]
	repoName = parts[1]

	return owner, repoName, nil
}

func getLatestReleaseAssets(client *http.Client, apiURL, authToken string) ([]GithubReleaseAssetsResponse, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request to %s", apiURL)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "binget-cli/1.0")
	if authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request to %s", apiURL)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response from %s: %d", apiURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s", apiURL)
	}

	var releaseResp GithubReleaseResponse
	err = json.Unmarshal(body, &releaseResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from %s: %w", apiURL, err)
	}

	return releaseResp.Assets, nil
}
