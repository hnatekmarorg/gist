package main

import (
	"os"
	"os/exec"
	"strings"
)

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