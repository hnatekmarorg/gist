package gist

import (
	"os"
	"os/exec"
	"strings"
)

// GetGitPath returns the git executable path.
func GetGitPath() string {
	if env := os.Getenv("GIST_GIT_PATH"); env != "" {
		return env
	}
	if env := os.Getenv("GIT_PATH"); env != "" {
		return env
	}
	return "git"
}

// RunGit runs a git command and returns trimmed stdout.
func RunGit(args ...string) (string, error) {
	cmd := exec.Command(GetGitPath(), args...)
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

// IsGitRepo checks if the current directory is inside a git repository.
func IsGitRepo() (bool, string) {
	out, err := RunGit("rev-parse", "--show-toplevel")
	if err != nil {
		return false, ""
	}
	return true, out
}