// Package config provides default embedded commands for goldfish.
// This file contains the default command definitions that are baked into
// the goldfish binary at build time using Go's embed functionality.
package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigSearchPaths defines the directories to search for commands.yml
// in order of precedence (highest to lowest)
var ConfigSearchPaths = []string{
	".", // Current working directory (highest precedence)
	"$HOME/.config/goldfish",
	"$HOME/.goldfish",
	"/etc/goldfish", // System-wide configuration (lowest precedence)
}

// defaultCommandsYAML contains the embedded default commands configuration
// This is loaded from default_commands.yml at build time
//go:embed default_commands.yml
var defaultCommandsYAML []byte

// LoadDefaults loads the embedded default commands configuration
// This provides a baseline set of commands that are always available
// without requiring an external commands.yml file
func LoadDefaults() (*Config, error) {
	// Parse the embedded YAML content
	var config Config
	if err := yaml.Unmarshal(defaultCommandsYAML, &config); err != nil {
		return nil, fmt.Errorf("failed to parse embedded default commands: %w", err)
	}

	// Validate the embedded configuration
	loader := &Loader{configPath: "embedded://defaults"}
	if err := loader.validate(&config); err != nil {
		return nil, fmt.Errorf("embedded default commands validation failed: %w", err)
	}

	return &config, nil
}

// MergeConfigs combines two configurations, with the override config
// taking precedence over the base config for commands with the same name or alias
func MergeConfigs(base, override *Config) *Config {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// Create a map of existing commands from override config for fast lookup
	overrideMap := make(map[string]bool)
	for _, cmd := range override.Commands {
		overrideMap[cmd.Name] = true
		if cmd.Alias != "" {
			overrideMap[cmd.Alias] = true
		}
	}

	// Start with all override commands
	merged := &Config{
		Commands: make([]Command, len(override.Commands)),
	}
	copy(merged.Commands, override.Commands)

	// Add base commands that aren't overridden
	for _, baseCmd := range base.Commands {
		// Check if this command is overridden by name or alias
		if !overrideMap[baseCmd.Name] && (baseCmd.Alias == "" || !overrideMap[baseCmd.Alias]) {
			merged.Commands = append(merged.Commands, baseCmd)
		}
	}

	return merged
}

// expandPath expands environment variables in a path
func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}
	
	// Expand $HOME or other environment variables
	if path[0] == '$' {
		end := 1
		for end < len(path) && (path[end] >= 'A' && path[end] <= 'Z' || path[end] == '_') {
			end++
		}
		if end > 1 {
			envVar := path[1:end]
			envVal := os.Getenv(envVar)
			if envVal != "" {
				return envVal + path[end:]
			}
		}
	}
	
	// Handle ~ expansion for home directory
	if path[0] == '~' && (len(path) == 1 || path[1] == '/') {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[1:])
		}
	}
	
	return os.ExpandEnv(path)
}

// findConfigFile searches for commands.yml in the configured search paths
// Returns the path to the first found file and true, or empty string and false if not found
func findConfigFile() (string, bool) {
	for _, searchPath := range ConfigSearchPaths {
		expandedPath := expandPath(searchPath)
		configPath := filepath.Join(expandedPath, "commands.yml")
		
		if _, err := os.Stat(configPath); err == nil {
			return configPath, true
		}
	}
	return "", false
}

// LoadWithDefaults loads configuration with embedded defaults as fallback
// It first loads the embedded defaults, then attempts to load and merge
// an optional runtime configuration file if it exists
func LoadWithDefaults(runtimeConfigPath string) (*Config, error) {
	// Always load embedded defaults first
	defaultConfig, err := LoadDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded defaults: %w", err)
	}

	// If a specific runtime config path was provided, try to load it
	if runtimeConfigPath != "" {
		loader := NewLoader(runtimeConfigPath)
		runtimeConfig, err := loader.Load()
		if err == nil {
			// Merge runtime config over defaults
			return MergeConfigs(defaultConfig, runtimeConfig), nil
		}
		// If runtime config doesn't exist or fails to load, use defaults only
		return defaultConfig, nil
	}
	
	// Otherwise, search for config files in the standard locations
	if configPath, found := findConfigFile(); found {
		loader := NewLoader(configPath)
		runtimeConfig, err := loader.Load()
		if err == nil {
			// Merge runtime config over defaults
			return MergeConfigs(defaultConfig, runtimeConfig), nil
		}
	}

	// If no runtime config found or loaded, use defaults only
	return defaultConfig, nil
}
