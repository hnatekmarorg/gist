package gist

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// CommandList prints all configured profiles.
func CommandList(cfg Config) {
	fmt.Println("available profiles:")
	for _, p := range cfg.Profiles {
		// Use a bullet for each profile.
		fmt.Printf("  • %s\t(%s)\n", p.Name, p.Email)
	}
}

// CommandInfo shows the current profile for the repository or globally.
func CommandInfo(cfg Config) {
	// Determine if we are inside a repo.
	inRepo, _ := IsGitRepo()
	var nameVal, emailVal string
	var err error
	if inRepo {
		nameVal, err = RunGit("config", "user.name")
		if err != nil {
			nameVal = ""
		}
		emailVal, err = RunGit("config", "user.email")
		if err != nil {
			emailVal = ""
		}
	} else {
		nameVal, err = RunGit("config", "--global", "user.name")
		if err != nil {
			nameVal = ""
		}
		emailVal, err = RunGit("config", "--global", "user.email")
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

// CommandSet activates a profile for the current repository.
func CommandSet(cfg Config, profileName string) error {
	p := FindProfile(&cfg, profileName)
	if p == nil {
		return fmt.Errorf("profile %s not found", profileName)
	}
	// Ensure we are inside a git repository.
	inRepo, repoRoot := IsGitRepo()
	if !inRepo {
		return errors.New("not inside a git repository")
	}
	// Set local git config values.
		if _, err := RunGit("config", "user.name", p.Username); err != nil {
		return fmt.Errorf("failed to set user.name: %w", err)
	}
		if _, err := RunGit("config", "user.email", p.Email); err != nil {
		return fmt.Errorf("failed to set user.email: %w", err)
	}
	if p.SigningKey != "" {
		if _, err := RunGit("config", "user.signingkey", p.SigningKey); err != nil {
			// Non‑fatal, continue.
			fmt.Fprintf(os.Stderr, "warning: failed to set signingkey: %v\n", err)
		}
	}
	fmt.Printf("✔️  Set profile \"%s\" for repository %s\n", p.Name, repoRoot)
	return nil
}

// CommandAdd interactively adds a new profile.
func CommandAdd(cfg *Config) error {
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

// CommandRemove deletes a profile from the config.
func CommandRemove(cfg *Config, name string) error {
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