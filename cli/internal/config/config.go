package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// CachedVersionInfo holds the cached update state for an installed emblem.
type CachedVersionInfo struct {
	LatestVersion     string    `yaml:"latest_version"`
	SecurityAdvisory  string    `yaml:"security_advisory,omitempty"`
	SecuritySeverity  string    `yaml:"security_severity,omitempty"`
	LastChecked       time.Time `yaml:"last_checked"`
}

type Config struct {
	Registry          string                       `yaml:"registry"`
	Token             string                       `yaml:"token,omitempty"`
	RefreshToken      string                       `yaml:"refresh_token,omitempty"`
	CurrentKey        string                       `yaml:"current_key,omitempty"`
	CacheDir          string                       `yaml:"cache_dir"`
	UserEmail         string                       `yaml:"user_email,omitempty"`
	Username          string                       `yaml:"username,omitempty"`
	Installed         map[string]string            `yaml:"installed"`
	UpdateCache       map[string]CachedVersionInfo `yaml:"update_cache,omitempty"`
	LastUpdateCheck   time.Time                    `yaml:"last_update_check,omitempty"`
	DisableUpdateCheck bool                        `yaml:"disable_update_check,omitempty"`
}

var (
	configDir string
	config    *Config
)

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir = filepath.Join(home, ".elysium")
	cacheDir := filepath.Join(configDir, "cache")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	config = &Config{
		Registry:    getEnvOrDefault("ELYSIUM_REGISTRY", "https://ely.karlharrenga.com"),
		CacheDir:    cacheDir,
		Installed:   make(map[string]string),
		UpdateCache: make(map[string]CachedVersionInfo),
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
		if config.UpdateCache == nil {
			config.UpdateCache = make(map[string]CachedVersionInfo)
		}
	}

	return nil
}

func Get() *Config {
	return config
}

func Save() error {
	configPath := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func SetRegistry(registry string) error {
	config.Registry = registry
	return Save()
}

func SetToken(token string) error {
	config.Token = token
	return Save()
}

func InstallEmblem(name, version string) error {
	config.Installed[name] = version
	return Save()
}

func UninstallEmblem(name string) error {
	delete(config.Installed, name)
	return Save()
}

func GetInstalledVersion(name string) (string, bool) {
	version, ok := config.Installed[name]
	return version, ok
}

func GetConfigDir() string {
	return configDir
}

func GetCacheDir() string {
	return config.CacheDir
}

func GetRegistry() string {
	return config.Registry
}

func GetOutput() string {
	return "table"
}

func GetInstalledEmblems() map[string]string {
	return config.Installed
}

func GetEmblemConfig(name string) (map[string]interface{}, error) {
	cfg := make(map[string]interface{})

	cfg["cache_dir"] = config.CacheDir
	cfg["registry"] = config.Registry

	if version, ok := config.Installed[name]; ok {
		cfg["version"] = version
	}

	return cfg, nil
}

func GetEmblemCachePath(name, version string) string {
	return filepath.Join(config.CacheDir, fmt.Sprintf("%s@%s", name, version))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetCurrentKey() string {
	return config.CurrentKey
}

func SetCurrentKey(keyID string) error {
	config.CurrentKey = keyID
	return Save()
}

func SetRefreshToken(token string) error {
	config.RefreshToken = token
	return Save()
}

func GetRefreshToken() string {
	return config.RefreshToken
}

func SetUserEmail(email string) {
	config.UserEmail = email
	Save()
}

func GetUserEmail() string {
	return config.UserEmail
}

func SetUsername(username string) {
	config.Username = username
	Save()
}

func GetUsername() string {
	return config.Username
}

func ClearAuth() error {
	config.Token = ""
	config.RefreshToken = ""
	config.UserEmail = ""
	config.Username = ""
	return Save()
}

// SetVersionCache stores the latest version info for an emblem in the update cache.
func SetVersionCache(name, latestVersion, securityAdvisory, securitySeverity string) error {
	config.UpdateCache[name] = CachedVersionInfo{
		LatestVersion:    latestVersion,
		SecurityAdvisory: securityAdvisory,
		SecuritySeverity: securitySeverity,
		LastChecked:      time.Now(),
	}
	return Save()
}

// GetVersionCache retrieves the cached version info for an emblem.
func GetVersionCache(name string) (CachedVersionInfo, bool) {
	info, ok := config.UpdateCache[name]
	return info, ok
}

// SetLastUpdateCheck records the timestamp of the most recent full update check.
func SetLastUpdateCheck() error {
	config.LastUpdateCheck = time.Now()
	return Save()
}

// GetLastUpdateCheck returns the timestamp of the last full update check.
func GetLastUpdateCheck() time.Time {
	return config.LastUpdateCheck
}

// IsUpdateCheckEnabled returns whether automatic update checks are enabled.
func IsUpdateCheckEnabled() bool {
	return !config.DisableUpdateCheck
}

// SetUpdateCheckEnabled sets whether automatic update checks are enabled.
func SetUpdateCheckEnabled(enabled bool) error {
	config.DisableUpdateCheck = !enabled
	return Save()
}
