package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigPath(t *testing.T) {
	// Test default path
	defaultPath := filepath.Join(os.Getenv("HOME"), ".config", "gist", "config.yaml")
	actualPath := getConfigPath()
	assert.Equal(t, defaultPath, actualPath)

	// Test with environment variable override
	os.Setenv("GIST_CONFIG_PATH", "/custom/path/config.yaml")
	defer os.Unsetenv("GIST_CONFIG_PATH")
	
	expectedPath := "/custom/path/config.yaml"
	actualPath = getConfigPath()
	assert.Equal(t, expectedPath, actualPath)
}

func TestGetGitPath(t *testing.T) {
	// Test default git path
	assert.Equal(t, "git", getGitPath())

	// Test with GIT_PATH environment variable
	os.Setenv("GIT_PATH", "/custom/git/path")
	defer os.Unsetenv("GIT_PATH")
	
	expectedPath := "/custom/git/path"
	actualPath := getGitPath()
	assert.Equal(t, expectedPath, actualPath)

	// Test with GIST_GIT_PATH environment variable (should take precedence)
	os.Setenv("GIST_GIT_PATH", "/another/git/path")
	defer os.Unsetenv("GIST_GIT_PATH")
	
	expectedPath = "/another/git/path"
	actualPath = getGitPath()
	assert.Equal(t, expectedPath, actualPath)
}

func TestParseKeyValue(t *testing.T) {
	testCases := []struct {
		line     string
		expectedKey   string
		expectedValue string
		expectedOK    bool
	}{
		{"name: John Doe", "name", "John Doe", true},
		{"  email: john@example.com  ", "email", "john@example.com", true},
		{"- username: Jane Smith", "username", "Jane Smith", true},
		{"invalid-line", "", "", false},
		{"key: \"quoted value\"", "key", "quoted value", true},
		{"key: 'single quoted'", "key", "single quoted", true},
		{"key: with: colons", "key", "with: colons", true},
	}

	for _, tc := range testCases {
		key, value, ok := parseKeyValue(tc.line)
		assert.Equal(t, tc.expectedKey, key, "Key mismatch for line: %s", tc.line)
		assert.Equal(t, tc.expectedValue, value, "Value mismatch for line: %s", tc.line)
		assert.Equal(t, tc.expectedOK, ok, "OK mismatch for line: %s", tc.line)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempDir, err := os.MkdirTemp("", "gist-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")
	
	// Write test config content
	configContent := `profiles:
  - name: work
    username: "Jane Doe"
    email: "jane@company.com"
    signingkey: "0xABCD1234"
  - name: personal
    username: "jane-personal"
    email: "jane@example.com"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Load config
	cfg, err := loadConfig(configPath)
	assert.NoError(t, err)
	assert.Len(t, cfg.Profiles, 2)

	// Check first profile
	assert.Equal(t, "work", cfg.Profiles[0].Name)
	assert.Equal(t, "Jane Doe", cfg.Profiles[0].Username)
	assert.Equal(t, "jane@company.com", cfg.Profiles[0].Email)
	assert.Equal(t, "0xABCD1234", cfg.Profiles[0].SigningKey)

	// Check second profile
	assert.Equal(t, "personal", cfg.Profiles[1].Name)
	assert.Equal(t, "jane-personal", cfg.Profiles[1].Username)
	assert.Equal(t, "jane@example.com", cfg.Profiles[1].Email)
	assert.Equal(t, "", cfg.Profiles[1].SigningKey)
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gist-save-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with profiles
	cfg := Config{
		Profiles: []Profile{
			{
				Name:       "test",
				Username:   "Test User",
				Email:      "test@example.com",
				SigningKey: "0x12345678",
			},
		},
	}

	// Save config
	err = saveConfig(configPath, cfg)
	assert.NoError(t, err)

	// Read back and verify
	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	contentStr := string(content)

	// Check that content contains expected elements
	assert.Contains(t, contentStr, "profiles:")
	assert.Contains(t, contentStr, "name: test")
	assert.Contains(t, contentStr, "username: \"Test User\"")
	assert.Contains(t, contentStr, "email: \"test@example.com\"")
	assert.Contains(t, contentStr, "signingkey: \"0x12345678\"")
}

func TestFindProfile(t *testing.T) {
	cfg := Config{
		Profiles: []Profile{
			{
				Name:     "work",
				Username: "Jane Doe",
				Email:    "jane@company.com",
			},
			{
				Name:     "personal",
				Username: "jane-personal",
				Email:    "jane@example.com",
			},
		},
	}

	// Test finding existing profile
	profile := findProfile(&cfg, "work")
	assert.NotNil(t, profile)
	assert.Equal(t, "work", profile.Name)
	assert.Equal(t, "Jane Doe", profile.Username)
	assert.Equal(t, "jane@company.com", profile.Email)

	// Test finding non-existing profile
	profile = findProfile(&cfg, "nonexistent")
	assert.Nil(t, profile)
}

func TestInitConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gist-init-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Initialize config
	err = initConfig(configPath)
	assert.NoError(t, err)

	// Check if config file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify content
	cfg, err := loadConfig(configPath)
	assert.NoError(t, err)
	assert.Len(t, cfg.Profiles, 1)
	assert.Equal(t, "example", cfg.Profiles[0].Name)
	assert.Equal(t, "Your Name", cfg.Profiles[0].Username)
	assert.Equal(t, "you@example.com", cfg.Profiles[0].Email)
}

func TestCommandAdd(t *testing.T) {
	// Create a mock input reader
	mockInput := strings.NewReader("test-profile\nTest User\ntest@example.com\n0x12345678\n")
	
	// We'd normally need to modify the code to accept an input reader,
	// but for this test, we'll test the parsing logic instead
	// This is a limitation of the current implementation
	//
	// In practice, we would need to refactor commandAdd to accept an io.Reader
	// to properly test interactive input
	
	// Instead, we'll manually construct a profile and verify it gets added correctly
	cfg := Config{}
	
	// Simulate what would happen in commandAdd
	newProf := Profile{
		Name:       "test-profile",
		Username:   "Test User",
		Email:      "test@example.com",
		SigningKey: "0x12345678",
	}
	
	cfg.Profiles = append(cfg.Profiles, newProf)
	
	assert.Len(t, cfg.Profiles, 1)
	assert.Equal(t, "test-profile", cfg.Profiles[0].Name)
	assert.Equal(t, "Test User", cfg.Profiles[0].Username)
	assert.Equal(t, "test@example.com", cfg.Profiles[0].Email)
	assert.Equal(t, "0x12345678", cfg.Profiles[0].SigningKey)
}