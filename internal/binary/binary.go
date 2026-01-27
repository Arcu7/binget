// Package binary provides functionality to find and download the latest release
// TODO: Refactor this whole thing lol
package binary

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/Arcu7/binget/internal/util/syslist"
	"github.com/Arcu7/binget/internal/util/transform"
)

type Finder struct {
	client *http.Client
	logger *slog.Logger
}

func NewFinder(logger *slog.Logger) *Finder {
	return &Finder{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (f *Finder) DownloadRelease(request Request) (err error) {
	f.logger.Info("Starting download process...")
	f.logger.Debug("Request details", "request", request)
	url, err := checkRequest(request)
	if err != nil {
		return err
	}

	owner, repoName, err := getOwnerAndRepoName(url.Path)
	if err != nil {
		return err
	}

	slog.Info("Fetching latest release assets...", slog.String("owner", owner), slog.String("repo", repoName))

	apiURL := fmt.Sprintf(GithubAPIURL, owner, repoName)
	assets, err := f.getLatestReleaseAssets(apiURL, request.AuthToken)
	if err != nil {
		return err
	}

	slog.Info("Fetched latest release assets", slog.Int("assetCount", len(assets)))

	// Get architecture aliases
	var goarch []string
	goarch = append(goarch, request.Arch)
	goarch = append(goarch, syslist.KnownArchAliases[runtime.GOARCH]...)

	slog.Info("Searching for suitable release asset", slog.String("os", request.OS), slog.Any("arch", goarch))

	// Filter download link based on OS, architecture, and file type
	assetsIter := slices.Values(assets)
	asset, found := transform.FindBy(assetsIter, getCorrectAssetsCondition(request.OS, goarch))
	if !found {
		return fmt.Errorf("no suitable release asset found")
	}

	fileType, err := getFileExtension(asset.Name)
	if err != nil {
		return err
	}

	f.logger.Info("Found suitable release asset", slog.String("downloadURL", asset.BrowserDownloadURL))

	release, err := f.downloadRelease(asset.BrowserDownloadURL)
	if err != nil {
		return err
	}

	f.logger.Info("Download process completed successfully")
	f.logger.Info("Starting extraction process...")

	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*", asset.Name))
	if err != nil {
		f.logger.Error(
			"Failed to create temporary file",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create temporary file")
	}
	defer f.cleanUpFile(tempFile, true)

	_, err = tempFile.Write(release)
	if err != nil {
		slog.Error(
			"Failed to write to temporary file",
			slog.String("tempFile", tempFile.Name()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to write to temporary file")
	}

	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		f.logger.Error(
			"Failed to seek temporary file",
			slog.String("tempFile", tempFile.Name()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to seek temporary file")
	}

	var binaryFile *os.File
	switch fileType {
	case FileTypeZIP:
		// TODO: implement zip extraction
	case FileTypeTarGz:
		binaryFile, err = f.extractBinaryFromTarGz(tempFile)
		if err != nil {
			return fmt.Errorf("failed to extract binary: %w", err)
		}

		f.logger.Info("Extracted binary from tar.gz archive", slog.String("binaryPath", binaryFile.Name()))
	}

	// Move the binary to the path specified in the request
	destinationPath := filepath.Join(request.PathEnv, binaryFile.Name())
	err = os.Rename(binaryFile.Name(), destinationPath)
	if err != nil {
		f.cleanUpFile(binaryFile, true)
		return fmt.Errorf("failed to move binary to destination: %w", err)
	}

	// Change permission to be executable
	info, err := binaryFile.Stat()
	if err != nil {
		f.cleanUpFile(binaryFile, true)
		return fmt.Errorf("failed to get binary file info: %w", err)
	}

	var mode os.FileMode
	if request.OS == syslist.OSLinux || request.OS == syslist.OSDarwin {
		chmod := info.Mode().Perm()
		f.logger.Debug("Current file permissions", slog.String("permissions", chmod.String()))
		mode = chmod | 0o111 // Set executable bits for user, group, others
		f.logger.Debug("New file permissions", slog.String("permissions", mode.String()))
	}
	err = binaryFile.Chmod(mode)
	if err != nil {
		f.cleanUpFile(binaryFile, true)
		return fmt.Errorf("failed to set executable permission on binary: %w", err)
	}

	f.cleanUpFile(binaryFile, false)

	slog.Info("Binary moved to destination successfully", slog.String("destination", destinationPath))
	slog.Info("Extraction process completed successfully")

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

func (f *Finder) getLatestReleaseAssets(apiURL, authToken string) ([]GithubReleaseAssetsResponse, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request to %s", apiURL)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "binget-cli/1.0")
	if authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	resp, err := f.client.Do(req)
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

// Get the correct asset condition based on OS and Architecture for the FindBy function
// Using a closure to capture os and arch variables
func getCorrectAssetsCondition(os string, arch []string) func(asset GithubReleaseAssetsResponse) bool {
	return func(asset GithubReleaseAssetsResponse) bool {
		if !strings.Contains(strings.ToLower(asset.BrowserDownloadURL), os) {
			return false
		}

		for _, a := range arch {
			if strings.Contains(strings.ToLower(asset.BrowserDownloadURL), a) {
				isTarFile := strings.HasSuffix(asset.Name, ".tar.gz")
				isZIPFile := strings.HasSuffix(asset.Name, ".zip")

				if isTarFile || isZIPFile {
					return true
				}
			}
		}

		return false
	}
}

func (f *Finder) downloadRelease(downloadLink string) ([]byte, error) {
	req, err := http.NewRequest("GET", downloadLink, nil)
	if err != nil {
		f.logger.Error(
			"Failed to create HTTP request",
			slog.String("downloadLink", downloadLink),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to create HTTP request")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "binget-cli/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		f.logger.Error(
			"Failed to perform HTTP request",
			slog.String("downloadLink", downloadLink),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to perform HTTP request")
	}
	if resp.StatusCode != http.StatusOK {
		f.logger.Error(
			"Failed to get valid response",
			slog.String("downloadLink", downloadLink),
			slog.Int("statusCode", resp.StatusCode),
		)
		return nil, fmt.Errorf("received non-200 response")
	}

	file, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body")
	}

	return file, nil
}

func getFileExtension(fileName string) (FileType, error) {
	if strings.HasSuffix(fileName, ".tar.gz") {
		return FileTypeTarGz, nil
	}

	ext := filepath.Ext(fileName)
	switch ext {
	case ".zip":
		return FileTypeZIP, nil
	default:
		return "", fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func (f *Finder) cleanUpFile(file *os.File, remove bool) {
	f.logger.Debug("Closing file", slog.String("file", file.Name()))
	closeErr := file.Close()
	if closeErr != nil {
		f.logger.Error(
			"Failed to close temporary file",
			slog.String("file", file.Name()),
			slog.String("error", closeErr.Error()),
		)
	}

	if remove {
		f.logger.Debug("Removing file", slog.String("file", file.Name()))
		removeErr := os.Remove(file.Name())
		if removeErr != nil {
			f.logger.Error(
				"Failed to remove temporary file",
				slog.String("file", file.Name()),
				slog.String("error", removeErr.Error()),
			)
		}
	}
}
