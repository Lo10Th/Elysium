package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "creates default config directory",
			setup: func() string {
				dir := filepath.Join(os.TempDir(), "elysium-test-"+time.Now().Format("20060102150405"))
				os.Setenv("HOME", dir)
				return dir
			},
			wantErr: false,
		},
		{
			name: "loads existing config",
			setup: func() string {
				dir := filepath.Join(os.TempDir(), "elysium-test-existing-"+time.Now().Format("20060102150405"))
				os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
				configData := "registry: https://test.example.com\ntoken: test-token\ninstalled:\n  test-emblem: 1.0.0\n"
				os.WriteFile(filepath.Join(dir, ".elysium", "config.yaml"), []byte(configData), 0644)
				os.Setenv("HOME", dir)
				return dir
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup()
			defer os.RemoveAll(dir)

			// Reset config for each test
			config = nil
			configDir = ""

			err := Init()
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if config == nil {
					t.Error("Init() did not initialize config")
				}
				if configDir == "" {
					t.Error("Init() did not set configDir")
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg := Get()
	if cfg == nil {
		t.Error("Get() returned nil")
	}
}

func TestSave(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-save-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	config.Token = "test-token"
	err = Save()
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".elysium", "config.yaml"))
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(data) == "" {
		t.Error("Save() did not write config to file")
	}
}

func TestSetRegistry(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-registry-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetRegistry("https://custom.registry.com")
	if err != nil {
		t.Errorf("SetRegistry() error = %v", err)
	}

	if config.Registry != "https://custom.registry.com" {
		t.Errorf("SetRegistry() did not update config, got %v", config.Registry)
	}
}

func TestSetToken(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-token-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetToken("new-token")
	if err != nil {
		t.Errorf("SetToken() error = %v", err)
	}

	if config.Token != "new-token" {
		t.Errorf("SetToken() did not update config, got %v", config.Token)
	}
}

func TestInstallEmblem(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-install-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = InstallEmblem("test-emblem", "1.0.0")
	if err != nil {
		t.Errorf("InstallEmblem() error = %v", err)
	}

	version, ok := config.Installed["test-emblem"]
	if !ok {
		t.Error("InstallEmblem() did not add emblem to installed map")
	}
	if version != "1.0.0" {
		t.Errorf("InstallEmblem() version = %v, want 1.0.0", version)
	}
}

func TestUninstallEmblem(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-uninstall-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	config.Installed["test-emblem"] = "1.0.0"
	Save()

	err = UninstallEmblem("test-emblem")
	if err != nil {
		t.Errorf("UninstallEmblem() error = %v", err)
	}

	_, ok := config.Installed["test-emblem"]
	if ok {
		t.Error("UninstallEmblem() did not remove emblem from installed map")
	}
}

func TestGetInstalledVersion(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-version-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	config.Installed["my-emblem"] = "2.0.0"

	version, ok := GetInstalledVersion("my-emblem")
	if !ok {
		t.Error("GetInstalledVersion() returned false for existing emblem")
	}
	if version != "2.0.0" {
		t.Errorf("GetInstalledVersion() version = %v, want 2.0.0", version)
	}

	_, ok = GetInstalledVersion("nonexistent")
	if ok {
		t.Error("GetInstalledVersion() returned true for nonexistent emblem")
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-configdir-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	configDir := GetConfigDir()
	if configDir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	if configDir != filepath.Join(dir, ".elysium") {
		t.Errorf("GetConfigDir() = %v, want %v", configDir, filepath.Join(dir, ".elysium"))
	}
}

func TestGetCacheDir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-cachedir-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cacheDir := GetCacheDir()
	if cacheDir == "" {
		t.Error("GetCacheDir() returned empty string")
	}
}

func TestGetRegistry(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-getreg-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	registry := GetRegistry()
	if registry == "" {
		t.Error("GetRegistry() returned empty string")
	}
}

func TestGetInstalledEmblems(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-installed-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	config.Installed["emblem1"] = "1.0.0"
	config.Installed["emblem2"] = "2.0.0"

	installed := GetInstalledEmblems()
	if len(installed) != 2 {
		t.Errorf("GetInstalledEmblems() returned %d emblems, want 2", len(installed))
	}
}

func TestVersionCache(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-versioncache-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetVersionCache("test-emblem", "1.2.0", "Critical bug", "high")
	if err != nil {
		t.Errorf("SetVersionCache() error = %v", err)
	}

	info, ok := GetVersionCache("test-emblem")
	if !ok {
		t.Error("GetVersionCache() returned false for existing cache entry")
	}
	if info.LatestVersion != "1.2.0" {
		t.Errorf("GetVersionCache() LatestVersion = %v, want 1.2.0", info.LatestVersion)
	}
	if info.SecurityAdvisory != "Critical bug" {
		t.Errorf("GetVersionCache() SecurityAdvisory = %v, want 'Critical bug'", info.SecurityAdvisory)
	}
	if info.SecuritySeverity != "high" {
		t.Errorf("GetVersionCache() SecuritySeverity = %v, want 'high'", info.SecuritySeverity)
	}

	_, ok = GetVersionCache("nonexistent")
	if ok {
		t.Error("GetVersionCache() returned true for nonexistent cache entry")
	}
}

func TestLastUpdateCheck(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-lastupdate-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	before := time.Now()
	err = SetLastUpdateCheck()
	if err != nil {
		t.Errorf("SetLastUpdateCheck() error = %v", err)
	}

	after := time.Now()
	lastCheck := GetLastUpdateCheck()

	if lastCheck.Before(before) || lastCheck.After(after) {
		t.Errorf("GetLastUpdateCheck() returned unexpected time %v", lastCheck)
	}
}

func TestIsUpdateCheckEnabled(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-updatecheck-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	enabled := IsUpdateCheckEnabled()
	if !enabled {
		t.Error("IsUpdateCheckEnabled() returned false by default")
	}

	err = SetUpdateCheckEnabled(false)
	if err != nil {
		t.Errorf("SetUpdateCheckEnabled(false) error = %v", err)
	}

	enabled = IsUpdateCheckEnabled()
	if enabled {
		t.Error("IsUpdateCheckEnabled() returned true after disabling")
	}

	err = SetUpdateCheckEnabled(true)
	if err != nil {
		t.Errorf("SetUpdateCheckEnabled(true) error = %v", err)
	}

	enabled = IsUpdateCheckEnabled()
	if !enabled {
		t.Error("IsUpdateCheckEnabled() returned false after enabling")
	}
}

func TestClearAuth(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-clearauth-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	config.Token = "test-token"
	config.RefreshToken = "test-refresh"
	config.UserEmail = "test@example.com"
	config.Username = "testuser"

	err = ClearAuth()
	if err != nil {
		t.Errorf("ClearAuth() error = %v", err)
	}

	if config.Token != "" {
		t.Error("ClearAuth() did not clear Token")
	}
	if config.RefreshToken != "" {
		t.Error("ClearAuth() did not clear RefreshToken")
	}
	if config.UserEmail != "" {
		t.Error("ClearAuth() did not clear UserEmail")
	}
	if config.Username != "" {
		t.Error("ClearAuth() did not clear Username")
	}
}

func TestSetCurrentKey(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-currentkey-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetCurrentKey("key-123")
	if err != nil {
		t.Errorf("SetCurrentKey() error = %v", err)
	}

	if GetCurrentKey() != "key-123" {
		t.Errorf("GetCurrentKey() = %v, want 'key-123'", GetCurrentKey())
	}
}

func TestSetRefreshToken(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-refreshtoken-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	err = SetRefreshToken("refresh-token-456")
	if err != nil {
		t.Errorf("SetRefreshToken() error = %v", err)
	}

	if GetRefreshToken() != "refresh-token-456" {
		t.Errorf("GetRefreshToken() = %v, want 'refresh-token-456'", GetRefreshToken())
	}
}

func TestSetUserEmail(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-email-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	SetUserEmail("user@example.com")

	if GetUserEmail() != "user@example.com" {
		t.Errorf("GetUserEmail() = %v, want 'user@example.com'", GetUserEmail())
	}
}

func TestSetUsername(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "elysium-username-test-"+time.Now().Format("20060102150405"))
	os.MkdirAll(filepath.Join(dir, ".elysium"), 0755)
	defer os.RemoveAll(dir)

	os.Setenv("HOME", dir)
	config = nil
	configDir = ""

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	SetUsername("testuser")

	if GetUsername() != "testuser" {
		t.Errorf("GetUsername() = %v, want 'testuser'", GetUsername())
	}
}
