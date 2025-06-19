// Package config provides default embedded commands for goldfish.
// This file contains the default command definitions that are baked into
// the goldfish binary at build time using Go's embed functionality.
package config

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

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

// LoadWithDefaults loads configuration with embedded defaults as fallback
// It first loads the embedded defaults, then attempts to load and merge
// an optional runtime configuration file if it exists
func LoadWithDefaults(runtimeConfigPath string) (*Config, error) {
	// Always load embedded defaults first
	defaultConfig, err := LoadDefaults()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded defaults: %w", err)
	}

	// Try to load runtime configuration
	loader := NewLoader(runtimeConfigPath)
	runtimeConfig, err := loader.Load()
	if err != nil {
		// If runtime config doesn't exist or fails to load, use defaults only
		// This makes the runtime config optional
		return defaultConfig, nil
	}

	// Merge runtime config over defaults
	return MergeConfigs(defaultConfig, runtimeConfig), nil
}