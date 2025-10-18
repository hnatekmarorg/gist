package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Config holds all profiles.
type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

// Profile represents a Git identity configuration.
type Profile struct {
	Name       string `yaml:"name"`
	Username   string `yaml:"username"`
	Email      string `yaml:"email"`
	SigningKey string `yaml:"signingkey,omitempty"`
}

// getConfigPath returns the path to the configuration file.
func getConfigPath() string {
	// Check env var override.
	if env := os.Getenv("GIST_CONFIG_PATH"); env != "" {
		return env
	}
	// Default location: $HOME/.config/gist/config.yaml
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory (unlikely).
		return "config.yaml"
	}
	return filepath.Join(home, ".config", "gist", "config.yaml")
}

// loadConfig reads the configuration file.
func loadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	lines := strings.Split(string(data), "\n")
	var current *Profile
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "profiles:") {
			continue
		}
		key, value, ok := parseKeyValue(line)
		if !ok {
			continue
		}
		switch key {
		case "name":
			// start a new profile
			p := Profile{Name: value}
			cfg.Profiles = append(cfg.Profiles, p)
			// set pointer to the newly added profile
			current = &cfg.Profiles[len(cfg.Profiles)-1]
		case "username":
			if current != nil {
				current.Username = value
			}
		case "email":
			if current != nil {
				current.Email = value
			}
		case "signingkey":
			if current != nil {
				current.SigningKey = value
			}
		default:
			// ignore unknown keys
		}
	}
	return cfg, nil
}

// saveConfig writes the configuration file.
func saveConfig(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("profiles:\n")
	for _, p := range cfg.Profiles {
		sb.WriteString("  - name: " + p.Name + "\n")
		sb.WriteString("    username: \"" + p.Username + "\"\n")
		sb.WriteString("    email: \"" + p.Email + "\"\n")
		if p.SigningKey != "" {
			sb.WriteString("    signingkey: \"" + p.SigningKey + "\"\n")
		}
	}
	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

// initConfig creates a default config if missing.
func initConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		// Already exists.
		return nil
	}
	cfg := Config{Profiles: []Profile{{Name: "example", Username: "Your Name", Email: "you@example.com"}}}
	return saveConfig(path, cfg)
}

// parseKeyValue parses a line like "key: value" (optionally prefixed with "-").
func parseKeyValue(line string) (key, value string, ok bool) {
	// Remove any leading dash.
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "-") {
		// Remove leading dash and any following spaces.
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimSpace(line)
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])
	// Strip surrounding quotes if present.
	value = strings.Trim(value, "\"'")
	return key, value, true
}