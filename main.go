package main

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v2"
)

// Version of the application.
const version = "v0.1.0"

// Profile represents a Git identity configuration.
type Profile struct {
    Name       string `yaml:"name"`
    Username   string `yaml:"username"`
    Email      string `yaml:"email"`
    SigningKey string `yaml:"signingkey,omitempty"`
}

// Config holds all profiles.
type Config struct {
    Profiles []Profile `yaml:"profiles"`
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

// getGitPath returns the git executable path.
func getGitPath() string {
    if env := os.Getenv("GIST_GIT_PATH"); env != "" {
        return env
    }
    if env := os.Getenv("GIT_PATH"); env != "" {
        return env
    }
    return "git"
}

// runGit runs a git command and returns trimmed stdout.
func runGit(args ...string) (string, error) {
    cmd := exec.Command(getGitPath(), args...)
    out, err := cmd.Output()
    if err != nil {
        // If git writes to stderr (e.g., when key not found), capture that.
        if ee, ok := err.(*exec.ExitError); ok {
            return strings.TrimSpace(string(ee.Stderr)), err
        }
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}

// isGitRepo checks if the current directory is inside a git repository.
func isGitRepo() (bool, string) {
    out, err := runGit("rev-parse", "--show-toplevel")
    if err != nil {
        return false, ""
    }
    return true, out
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

// loadConfig reads the configuration file.
func loadConfig(path string) (Config, error) {
    var cfg Config
    data, err := os.ReadFile(path)
    if err != nil {
        return cfg, err
    }
    
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return cfg, err
    }
    
    return cfg, nil
}

// saveConfig writes the configuration file.
func saveConfig(path string, cfg Config) error {
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return err
    }
    
    data, err := yaml.Marshal(&cfg)
    if err != nil {
        return err
    }
    
    return os.WriteFile(path, data, 0o644)
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

// findProfile returns a pointer to a profile by its name.
func findProfile(cfg *Config, name string) *Profile {
    for i, p := range cfg.Profiles {
        if p.Name == name {
            return &cfg.Profiles[i]
        }
    }
    return nil
}

// commandList prints all configured profiles.
func commandList(cfg Config) {
    fmt.Println("available profiles:")
    for _, p := range cfg.Profiles {
        // Use a bullet for each profile.
        fmt.Printf("  • %s\t(%s)\n", p.Name, p.Email)
    }
}

// commandInfo shows the current profile for the repository or globally.
func commandInfo(cfg Config) {
    // Determine if we are inside a repo.
    inRepo, _ := isGitRepo()
    var nameVal, emailVal string
    var err error
    if inRepo {
        nameVal, err = runGit("config", "user.name")
        if err != nil {
            nameVal = ""
        }
        emailVal, err = runGit("config", "user.email")
        if err != nil {
            emailVal = ""
        }
    } else {
        nameVal, err = runGit("config", "--global", "user.name")
        if err != nil {
            nameVal = ""
        }
        emailVal, err = runGit("config", "--global", "user.email")
        if err != nil {
            emailVal = ""
        }
    }
    // Find matching profile.
    var matched *Profile
    for i, p := range cfg.Profiles {
        if p.Username == nameVal && p.Email == emailVal {
            matched = &cfg.Profiles[i]
            break
        }
    }
    scope := "global"
    if inRepo {
        scope = "repo"
    }
    fmt.Printf("current profile (%s):\n", scope)
    if matched != nil {
        fmt.Printf("  name: %s\n", matched.Name)
        fmt.Printf("  user: %s <%s>\n", matched.Username, matched.Email)
        if matched.SigningKey != "" {
            fmt.Printf("  signingkey: %s\n", matched.SigningKey)
        }
    } else {
        fmt.Println("  (none)")
    }
}

// commandSet activates a profile for the current repository.
func commandSet(cfg Config, profileName string) error {
    p := findProfile(&cfg, profileName)
    if p == nil {
        return fmt.Errorf("profile %s not found", profileName)
    }
    // Ensure we are inside a git repository.
    inRepo, repoRoot := isGitRepo()
    if !inRepo {
        return errors.New("not inside a git repository")
    }
    // Set local git config values.
    if _, err := runGit("config", "user.name", p.Username); err != nil {
        return fmt.Errorf("failed to set user.name: %w", err)
    }
    if _, err := runGit("config", "user.email", p.Email); err != nil {
        return fmt.Errorf("failed to set user.email: %w", err)
    }
    if p.SigningKey != "" {
        if _, err := runGit("config", "user.signingkey", p.SigningKey); err != nil {
            // Non‑fatal, continue.
            fmt.Fprintf(os.Stderr, "warning: failed to set signingkey: %v\n", err)
        }
    }
    fmt.Printf("✔️  Set profile \"%s\" for repository %s\n", p.Name, repoRoot)
    return nil
}

// commandAdd interactively adds a new profile.
func commandAdd(cfg *Config) error {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter profile name: ")
    name, err := reader.ReadString('\n')
    if err != nil {
        return err
    }
    fmt.Print("Enter username (git user.name): ")
    username, err := reader.ReadString('\n')
    if err != nil {
        return err
    }
    fmt.Print("Enter email (git user.email): ")
    email, err := reader.ReadString('\n')
    if err != nil {
        return err
    }
    fmt.Print("Enter signing key (optional): ")
    signing, err := reader.ReadString('\n')
    if err != nil && err != io.EOF {
        return err
    }
    // Trim whitespace and newlines.
    name = strings.TrimSpace(name)
    username = strings.TrimSpace(username)
    email = strings.TrimSpace(email)
    signing = strings.TrimSpace(signing)
    if name == "" || username == "" || email == "" {
        return errors.New("profile name, username and email are required")
    }
    // Append new profile.
    newProf := Profile{Name: name, Username: username, Email: email, SigningKey: signing}
    cfg.Profiles = append(cfg.Profiles, newProf)
    fmt.Printf("Profile %s added.\n", name)
    return nil
}

// commandRemove deletes a profile from the config.
func commandRemove(cfg *Config, name string) error {
    idx := -1
    for i, p := range cfg.Profiles {
        if p.Name == name {
            idx = i
            break
        }
    }
    if idx == -1 {
        return fmt.Errorf("profile %s not found", name)
    }
    cfg.Profiles = append(cfg.Profiles[:idx], cfg.Profiles[idx+1:]...)
    fmt.Printf("Profile %s removed.\n", name)
    return nil
}

// printHelp displays usage information.
func printHelp() {
    fmt.Println("Usage: gist <command> [args]")
    fmt.Println("Commands:")
    fmt.Println("  init                 Create default config if missing")
    fmt.Println("  list                 Show all configured profiles")
    fmt.Println("  info                 Show current active profile")
    fmt.Println("  set <profile>        Activate a profile for the current repository")
    fmt.Println("  add                  Interactively add a new profile")
    fmt.Println("  remove <profile>     Delete a profile from config")
    fmt.Println("  --version            Print version and exit")
    fmt.Println("  --help               Show this help message")
}

func main() {
    args := os.Args[1:]
    if len(args) == 0 {
        printHelp()
        return
    }
    // Handle global flags.
    switch args[0] {
    case "--version":
        fmt.Println(version)
        return
    case "--help":
        printHelp()
        return
    }
    configPath := getConfigPath()
    // Load configuration; for commands that don't need config, we may ignore errors.
    cfg, cfgErr := loadConfig(configPath)

    switch args[0] {
    case "init":
        if err := initConfig(configPath); err != nil {
            fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
            os.Exit(1)
        }
        fmt.Println("Config initialized at", configPath)
    case "list":
        if cfgErr != nil {
            fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", cfgErr)
            os.Exit(1)
        }
        commandList(cfg)
    case "info":
        if cfgErr != nil {
            fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", cfgErr)
            os.Exit(1)
        }
        commandInfo(cfg)
    case "set":
        if len(args) < 2 {
            fmt.Fprintln(os.Stderr, "Usage: gist set <profile>")
            os.Exit(1)
        }
        if cfgErr != nil {
            fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", cfgErr)
            os.Exit(1)
        }
        if err := commandSet(cfg, args[1]); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    case "add":
        if cfgErr != nil {
            // If config doesn't exist, start with empty config.
            cfg = Config{}
        }
        if err := commandAdd(&cfg); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        // Save config after adding.
        if err := saveConfig(configPath, cfg); err != nil {
            fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
            os.Exit(1)
        }
    case "remove":
        if len(args) < 2 {
            fmt.Fprintln(os.Stderr, "Usage: gist remove <profile>")
            os.Exit(1)
        }
        if cfgErr != nil {
            fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", cfgErr)
            os.Exit(1)
        }
        if err := commandRemove(&cfg, args[1]); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        if err := saveConfig(configPath, cfg); err != nil {
            fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
            os.Exit(1)
        }
    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
        printHelp()
        os.Exit(1)
    }
}

