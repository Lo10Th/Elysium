package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/elysium/elysium/cli/internal/httpclient"
)

const downloadTimeout = 10 * time.Minute

// AssetName returns the expected release asset filename for the current platform.
func AssetName() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("ely-%s-%s%s.tar.gz", runtime.GOOS, runtime.GOARCH, ext)
}

// BinaryName returns the expected binary name inside the tarball.
func BinaryName() string {
	if runtime.GOOS == "windows" {
		return "ely.exe"
	}
	return "ely"
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

// DownloadBinary downloads the tarball from the given HTTPS URL, extracts the binary,
// and saves it to a temporary file. The caller is responsible for removing the temp file.
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

	resp, err := httpclient.DefaultClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file for extracted binary
	tmpFile, err := os.CreateTemp("", "ely-update-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	binaryName := BinaryName()

	// Extract binary from tar.gz
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	found := false

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		// Look for the binary file (may be in a subdirectory)
		if strings.HasSuffix(hdr.Name, binaryName) && !hdr.FileInfo().IsDir() {
			if _, err := io.Copy(tmpFile, tr); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return "", fmt.Errorf("failed to write extracted binary: %w", err)
			}
			found = true
			break
		}
	}

	if !found {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("binary %s not found in archive", binaryName)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	return tmpFile.Name(), nil
}
