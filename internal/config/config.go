package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	AppName    = "incidencias-tui"
	ConfigDir  = ".config/" + AppName
	ConfigFile = "config.json"
	DefaultAPI = "http://localhost:8190"
)

// Config holds the application configuration
type Config struct {
	APIURL string `json:"api_url"`
	Token  string `json:"token,omitempty"`
}

// configPath returns the full path to the config file
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home directory: %w", err)
	}
	dir := filepath.Join(home, ConfigDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}
	return filepath.Join(dir, ConfigFile), nil
}

// Load reads config from disk, returns defaults if not found
func Load() (*Config, error) {
	cfg := &Config{
		APIURL: DefaultAPI,
	}

	path, err := configPath()
	if err != nil {
		return cfg, nil // return defaults if we can't determine path
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config yet, return defaults
		}
		return cfg, nil // ignore other errors, return defaults
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, nil // ignore parse errors
	}

	return cfg, nil
}

// Save persists the config to disk
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	return nil
}

// ClearToken removes the stored token
func (c *Config) ClearToken() {
	c.Token = ""
	c.Save() //nolint:errcheck
}

// HasToken returns true if a token is stored
func (c *Config) HasToken() bool {
	return c.Token != ""
}
