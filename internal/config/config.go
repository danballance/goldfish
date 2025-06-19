// Package config provides YAML configuration parsing for goldfish commands.
// It defines the structure of command definitions and handles loading
// and validation of the commands.yml file.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Parameter represents a command parameter definition
// It defines the structure of parameters that commands can accept
type Parameter struct {
	// Name is the parameter identifier
	Name string `yaml:"name"`
	// Type defines the parameter type (string, bool, int)
	Type string `yaml:"type"`
	// Required indicates if this parameter is mandatory
	Required bool `yaml:"required"`
	// Flag is the CLI flag representation (e.g., "--in-place")
	Flag string `yaml:"flag,omitempty"`
	// Default provides a default value if not specified
	Default interface{} `yaml:"default,omitempty"`
	// Description explains what this parameter does
	Description string `yaml:"description,omitempty"`
}

// PlatformCommand represents a platform-specific command template
// It contains the template string that will be executed for a specific OS
type PlatformCommand struct {
	// Template is the Go template string for command generation
	Template string `yaml:"template"`
}

// Command represents a unified command definition
// It contains all the information needed to generate platform-specific commands
type Command struct {
	// Name is the primary command name
	Name string `yaml:"name"`
	// Alias provides an alternative shorter name
	Alias string `yaml:"alias,omitempty"`
	// Description explains what this command does
	Description string `yaml:"description"`
	// BaseCommand is the underlying system command (e.g., "sed", "find")
	BaseCommand string `yaml:"base_command"`
	// Parameters defines the accepted command parameters
	Parameters []Parameter `yaml:"params,omitempty"`
	// Platforms maps platform names to their command templates
	Platforms map[string]PlatformCommand `yaml:"platforms"`
}

// Config represents the complete goldfish configuration
// It contains all command definitions loaded from commands.yml
type Config struct {
	// Commands is the list of all available command definitions
	Commands []Command `yaml:"commands"`
}

// Loader handles loading and parsing of configuration files
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader
// configPath specifies the path to the commands.yml file
func NewLoader(configPath string) *Loader {
	return &Loader{
		configPath: configPath,
	}
}

// Load reads and parses the YAML configuration file
// It returns a Config struct containing all command definitions
func (l *Loader) Load() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", l.configPath)
	}

	// Read the config file
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", l.configPath, err)
	}

	// Parse YAML content
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate the loaded configuration
	if err := l.validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validate performs validation on the loaded configuration
// It checks for required fields and logical consistency
func (l *Loader) validate(config *Config) error {
	if len(config.Commands) == 0 {
		return fmt.Errorf("no commands defined in configuration")
	}

	// Track command names to detect duplicates
	nameMap := make(map[string]bool)
	aliasMap := make(map[string]bool)

	for i, cmd := range config.Commands {
		// Validate required fields
		if cmd.Name == "" {
			return fmt.Errorf("command at index %d: name is required", i)
		}
		if cmd.BaseCommand == "" {
			return fmt.Errorf("command '%s': base_command is required", cmd.Name)
		}
		if len(cmd.Platforms) == 0 {
			return fmt.Errorf("command '%s': at least one platform must be defined", cmd.Name)
		}

		// Check for duplicate names
		if nameMap[cmd.Name] {
			return fmt.Errorf("duplicate command name: %s", cmd.Name)
		}
		nameMap[cmd.Name] = true

		// Check for duplicate aliases
		if cmd.Alias != "" {
			if aliasMap[cmd.Alias] || nameMap[cmd.Alias] {
				return fmt.Errorf("duplicate command alias: %s", cmd.Alias)
			}
			aliasMap[cmd.Alias] = true
		}

		// Validate parameters
		for j, param := range cmd.Parameters {
			if param.Name == "" {
				return fmt.Errorf("command '%s': parameter at index %d: name is required", cmd.Name, j)
			}
			if param.Type == "" {
				return fmt.Errorf("command '%s': parameter '%s': type is required", cmd.Name, param.Name)
			}
			if !isValidParameterType(param.Type) {
				return fmt.Errorf("command '%s': parameter '%s': invalid type '%s'", cmd.Name, param.Name, param.Type)
			}
		}

		// Validate platform templates
		for platform, platformCmd := range cmd.Platforms {
			if platformCmd.Template == "" {
				return fmt.Errorf("command '%s': platform '%s': template is required", cmd.Name, platform)
			}
		}
	}

	return nil
}

// isValidParameterType checks if the parameter type is supported
func isValidParameterType(paramType string) bool {
	validTypes := []string{"string", "bool", "int", "float"}
	for _, validType := range validTypes {
		if paramType == validType {
			return true
		}
	}
	return false
}

// FindCommand searches for a command by name or alias
// It returns the command definition and true if found, nil and false otherwise
func (c *Config) FindCommand(nameOrAlias string) (*Command, bool) {
	for _, cmd := range c.Commands {
		if cmd.Name == nameOrAlias || cmd.Alias == nameOrAlias {
			return &cmd, true
		}
	}
	return nil, false
}

// GetCommandNames returns all command names and aliases
// Useful for generating help text and command completion
func (c *Config) GetCommandNames() []string {
	var names []string
	for _, cmd := range c.Commands {
		names = append(names, cmd.Name)
		if cmd.Alias != "" {
			names = append(names, cmd.Alias)
		}
	}
	return names
}

// LoadDefault loads the configuration from the default location
// It looks for commands.yml in the current directory
func LoadDefault() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	
	configPath := filepath.Join(wd, "commands.yml")
	loader := NewLoader(configPath)
	return loader.Load()
}

// LoadDefaultWithEmbedded loads configuration using embedded defaults with optional runtime override
// This is the recommended way to load configuration for distribution builds
func LoadDefaultWithEmbedded() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	
	configPath := filepath.Join(wd, "commands.yml")
	return LoadWithDefaults(configPath)
}