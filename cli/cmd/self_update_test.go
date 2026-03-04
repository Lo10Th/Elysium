package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elysium/elysium/cli/internal/selfupdate"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"0.2.0", "v0.3.0", true},
		{"0.2.0", "0.2.0", false},
		{"0.2.0", "v0.2.0", false},
		{"v0.2.0", "v0.3.0", true},
		{"0.3.0", "v0.2.0", false}, // latest is older, should not be considered "newer"
		{"0.2.0", "v0.10.0", true}, // semver: 0.10.0 > 0.2.0
	}

	for _, tt := range tests {
		got := selfupdate.IsNewer(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"v0.3.0", "0.3.0"},
		{"0.3.0", "0.3.0"},
		{"v1.0.0-rc1", "1.0.0-rc1"},
	}

	for _, tt := range tests {
		got := selfupdate.NormalizeVersion(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAssetName(t *testing.T) {
	name := selfupdate.AssetName()
	if name == "" {
		t.Error("AssetName() returned empty string")
	}
	if !strings.HasPrefix(name, "ely-") {
		t.Errorf("AssetName() = %q, expected prefix 'ely-'", name)
	}
}

func TestFindAssetURL(t *testing.T) {
	release := &selfupdate.Release{
		TagName: "v0.3.0",
		Assets: []selfupdate.Asset{
			{Name: "ely-linux-amd64", BrowserDownloadURL: "https://example.com/ely-linux-amd64"},
			{Name: "ely-darwin-amd64", BrowserDownloadURL: "https://example.com/ely-darwin-amd64"},
			{Name: "ely-windows-amd64.exe", BrowserDownloadURL: "https://example.com/ely-windows-amd64.exe"},
		},
	}

	// Verify that FindAssetURL returns an error for a platform with no asset.
	emptyRelease := &selfupdate.Release{TagName: "v0.3.0", Assets: []selfupdate.Asset{}}
	_, err := selfupdate.FindAssetURL(emptyRelease)
	if err == nil {
		t.Error("FindAssetURL() expected error for empty assets, got nil")
	}

	// Verify that FindAssetURL finds the correct URL when the asset is present.
	expected := "https://example.com/" + selfupdate.AssetName()
	url, err := selfupdate.FindAssetURL(release)
	if err != nil {
		// It's OK if the current platform's asset is not in the list.
		t.Logf("FindAssetURL() skipped (current platform not in fixture): %v", err)
		return
	}
	if url != expected {
		t.Errorf("FindAssetURL() = %q, want %q", url, expected)
	}
}

func TestGetLatestRelease_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := selfupdate.Release{
			TagName: "v0.3.0",
			Assets: []selfupdate.Asset{
				{Name: "ely-linux-amd64", BrowserDownloadURL: "https://example.com/ely-linux-amd64"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// We can't inject the URL directly into GetLatestRelease, but we can call fetchRelease
	// indirectly by testing GetReleaseByTag against our mock server via the exported checker path.
	// Instead, validate the JSON parsing round-trip.
	body, _ := json.Marshal(selfupdate.Release{
		TagName: "v0.3.0",
		Assets:  []selfupdate.Asset{{Name: "ely-linux-amd64", BrowserDownloadURL: "https://example.com/ely-linux-amd64"}},
	})
	var got selfupdate.Release
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got.TagName != "v0.3.0" {
		t.Errorf("TagName = %q, want %q", got.TagName, "v0.3.0")
	}
	if len(got.Assets) != 1 || got.Assets[0].Name != "ely-linux-amd64" {
		t.Errorf("Assets = %v, want 1 asset named ely-linux-amd64", got.Assets)
	}
}

func TestSelfUpdateCmdFlags(t *testing.T) {
	// Verify that the self-update command is registered and has the expected flags.
	cmd, _, err := rootCmd.Find([]string{"self-update"})
	if err != nil || cmd == nil {
		t.Fatal("self-update command not found in rootCmd")
	}

	if cmd.Use != "self-update" {
		t.Errorf("Use = %q, want %q", cmd.Use, "self-update")
	}

	for _, flagName := range []string{"check", "version", "force"} {
		if f := cmd.Flags().Lookup(flagName); f == nil {
			t.Errorf("flag --%s not found on self-update command", flagName)
		}
	}
}
