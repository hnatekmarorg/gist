# GIST – Git Identity Switching Tool
A quick gist of who you are – switch git personas in a snap.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/Hnatekmar/gist?include_prereleases&sort=semver&color=brightgreen)](https://github.com/Hnatekmar/gist/releases)

GIST is a tiny, cross-platform CLI for managing multiple Git user profiles (name + email + optional GPG key) and switching between them on a per-directory basis. No more manual `git config` gymnastics!

---

## Configuration

GIST reads a YAML file (default: `$HOME/.config/gist/config.yaml`).  
You can override the location with the environment variable `GIST_CONFIG_PATH`.

```yaml
# $HOME/.config/gist/config.yaml
profiles:
  - name: work
    username: "Jane Doe"
    email: "jane@company.com"
    signingkey: "0xABCD1234"   # optional – GPG key used for signing commits
  - name: personal
    username: "jane-personal"
    email: "jane@example.com"
```

### Generating a starter config

```bash
gist init               # creates ~/.config/gist/config.yaml with an example entry
```

---

## Commands

| Command | Synopsis | Example |
|---------|----------|---------|
| `list` | Show all configured profiles. | `gist list` |
| `info` | Print the profile currently active **in the current repository** (or the global one if no repo). | `gist info` |
| `set <profile>` | Activate a profile for the current repository (writes `.git/config`). | `gist set work` |
| `add` | Interactively add a new profile (writes to the config file). | `gist add` |
| `remove <profile>` | Delete a profile from the config file. | `gist remove personal` |
| `init` | Create a default config file if none exists. | `gist init` |
| `--version` | Print the version and exit. | `gist --version` |
| `--help` | Show help for the top-level command or a sub-command (`gist help set`). | `gist --help` |

### Sample usage

```bash
$ gist list
available profiles:
  • work      (jane@company.com)
  • personal  (jane@example.com)

$ gist info
current profile (global):
  name: personal
  user: jane-personal <jane@example.com>

$ gist set work
Set profile "work" for repository /home/jane/project
$ git config user.name
Jane Doe
$ git config user.email
jane@company.com
```

---

## Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GIST_CONFIG_PATH` | Absolute path to the YAML configuration file. | `$HOME/.config/gist/config.yaml` |
| `GIT_PATH` (or `GIST_GIT_PATH`) | Path to the `git` executable (useful on Windows where `git.exe` lives elsewhere). | `git` (found on `$PATH`) |
| `GIST_VERBOSE` | Set to `1` to enable extra debug output. | unset |

---

## Development & Contributions

1. Clone the repo  
   ```bash
   git clone https://github.com/Hnatekmar/gist.git
   cd gist
   ```

2. Run the test suite  
   ```bash
   go test ./...
   ```

3. Open a PR – Follow the existing code style, update documentation, and add tests for new features.

---
