package selfupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	githubAPIBase = "https://api.github.com/repos/Lo10Th/Elysium/releases"
	httpTimeout   = 15 * time.Second
)

// Release represents a GitHub release.
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a downloadable file attached to a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetLatestRelease fetches the latest release from GitHub.
func GetLatestRelease() (*Release, error) {
	return fetchRelease(githubAPIBase + "/latest")
}

// GetReleaseByTag fetches a specific release by tag name (e.g. "v0.3.0").
func GetReleaseByTag(tag string) (*Release, error) {
	return fetchRelease(fmt.Sprintf("%s/tags/%s", githubAPIBase, tag))
}

func fetchRelease(url string) (*Release, error) {
	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

// NormalizeVersion strips a leading "v" prefix for comparison.
func NormalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// IsNewer returns true when latestTag is a strictly higher semantic version than
// currentVersion. Both values are normalised (leading "v" stripped) before comparison.
// If either value cannot be parsed as X.Y.Z, a plain string inequality test is used
// as a fallback so that the function never silently reports "up to date" for unusual
// version strings.
func IsNewer(currentVersion, latestTag string) bool {
	cur := NormalizeVersion(currentVersion)
	lat := NormalizeVersion(latestTag)
	if cur == lat {
		return false
	}

	cv, cOK := parseSemver(cur)
	lv, lOK := parseSemver(lat)
	if cOK && lOK {
		for i := range cv {
			if lv[i] > cv[i] {
				return true
			}
			if lv[i] < cv[i] {
				return false
			}
		}
		return false
	}

	// Fallback: different strings are treated as "newer" to avoid silently
	// skipping non-standard version strings.
	return cur != lat
}

// parseSemver parses a "X.Y.Z" string into a [3]int array.
func parseSemver(v string) ([3]int, bool) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var nums [3]int
	for i, p := range parts {
		// Strip any pre-release suffix (e.g. "1-rc1" → "1").
		p = strings.SplitN(p, "-", 2)[0]
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, false
		}
		nums[i] = n
	}
	return nums, true
}
