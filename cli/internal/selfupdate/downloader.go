package selfupdate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

const downloadTimeout = 10 * time.Minute

// AssetName returns the expected release asset filename for the current platform.
func AssetName() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("ely-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)
}

// FindAssetURL searches a release's assets for the binary matching the current platform
// and returns its download URL.
func FindAssetURL(release *Release) (string, error) {
	name := AssetName()
	for _, asset := range release.Assets {
		if asset.Name == name {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
}

// DownloadBinary downloads the binary from the given HTTPS URL to a temporary file.
// The caller is responsible for removing the temp file after use.
func DownloadBinary(url string) (string, error) {
	if len(url) < 8 || url[:8] != "https://" {
		return "", fmt.Errorf("refusing to download over non-HTTPS URL: %s", url)
	}

	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "ely-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write downloaded binary: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	return tmpFile.Name(), nil
}
