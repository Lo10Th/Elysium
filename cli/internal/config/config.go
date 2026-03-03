package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Registry   string            `yaml:"registry"`
	Token      string            `yaml:"token,omitempty"`
	CurrentKey string            `yaml:"current_key,omitempty"`
	CacheDir   string            `yaml:"cache_dir"`
	Installed  map[string]string `yaml:"installed"`
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
		Registry:  getEnvOrDefault("ELYSIUM_REGISTRY", "https://registry.elysium.dev"),
		CacheDir:  cacheDir,
		Installed: make(map[string]string),
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
