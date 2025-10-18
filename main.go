package main

import (
	"errors"
	"fmt"
	"os"
)

// Version of the application.
const version = "v0.1.0"

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

