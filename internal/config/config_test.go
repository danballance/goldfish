// Package config_test provides unit tests for the config parsing module.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoader_Load tests the Load method of the Loader
func TestLoader_Load(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_commands.yml")

	validYAML := `
commands:
  - name: "test-command"
    alias: "test"
    description: "A test command"
    base_command: "echo"
    params:
      - name: "message"
        type: "string"
        required: true
      - name: "verbose"
        type: "bool"
        flag: "--verbose"
    platforms:
      linux:
        template: "{{.base_command}} '{{.params.message}}'"
      darwin:
        template: "{{.base_command}} '{{.params.message}}'"
`

	// Write the valid YAML to the temp file
	err := os.WriteFile(configPath, []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test loading the valid config
	loader := NewLoader(configPath)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify the loaded config
	if len(config.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(config.Commands))
	}

	cmd := config.Commands[0]
	if cmd.Name != "test-command" {
		t.Errorf("Expected command name 'test-command', got '%s'", cmd.Name)
	}
	if cmd.Alias != "test" {
		t.Errorf("Expected command alias 'test', got '%s'", cmd.Alias)
	}
	if cmd.BaseCommand != "echo" {
		t.Errorf("Expected base command 'echo', got '%s'", cmd.BaseCommand)
	}
	if len(cmd.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(cmd.Parameters))
	}
	if len(cmd.Platforms) != 2 {
		t.Errorf("Expected 2 platforms, got %d", len(cmd.Platforms))
	}
}

// TestLoader_Load_FileNotFound tests loading a non-existent config file
func TestLoader_Load_FileNotFound(t *testing.T) {
	loader := NewLoader("/nonexistent/path/commands.yml")
	_, err := loader.Load()
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("Expected 'config file not found' error, got: %v", err)
	}
}

// TestLoader_Load_InvalidYAML tests loading invalid YAML
func TestLoader_Load_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yml")

	invalidYAML := `
commands:
  - name: "test"
    invalid_yaml: [unclosed bracket
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	loader := NewLoader(configPath)
	_, err = loader.Load()
	if err == nil {
		t.Error("Expected error for invalid YAML, got none")
	}
}

// TestLoader_validate tests the validation function
func TestLoader_validate(t *testing.T) {
	loader := NewLoader("")

	// Test valid config
	validConfig := &Config{
		Commands: []Command{
			{
				Name:        "test",
				Description: "Test command",
				BaseCommand: "echo",
				Parameters: []Parameter{
					{Name: "msg", Type: "string", Required: true},
				},
				Platforms: map[string]PlatformCommand{
					"linux": {Template: "echo {{.params.msg}}"},
				},
			},
		},
	}

	err := loader.validate(validConfig)
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test config with missing command name
	invalidConfig := &Config{
		Commands: []Command{
			{
				Description: "Test command",
				BaseCommand: "echo",
				Platforms: map[string]PlatformCommand{
					"linux": {Template: "echo test"},
				},
			},
		},
	}

	err = loader.validate(invalidConfig)
	if err == nil {
		t.Error("Expected error for missing command name, got none")
	}
}

// TestConfig_FindCommand tests the FindCommand method
func TestConfig_FindCommand(t *testing.T) {
	config := &Config{
		Commands: []Command{
			{
				Name:        "test-command",
				Alias:       "test",
				Description: "Test command",
				BaseCommand: "echo",
			},
			{
				Name:        "another-command",
				Description: "Another command",
				BaseCommand: "cat",
			},
		},
	}

	// Test finding by name
	cmd, found := config.FindCommand("test-command")
	if !found {
		t.Error("Expected to find command by name")
	}
	if cmd.Name != "test-command" {
		t.Errorf("Expected command name 'test-command', got '%s'", cmd.Name)
	}

	// Test finding by alias
	cmd, found = config.FindCommand("test")
	if !found {
		t.Error("Expected to find command by alias")
	}
	if cmd.Name != "test-command" {
		t.Errorf("Expected command name 'test-command', got '%s'", cmd.Name)
	}

	// Test finding non-existent command
	_, found = config.FindCommand("nonexistent")
	if found {
		t.Error("Expected not to find non-existent command")
	}
}

// TestConfig_GetCommandNames tests the GetCommandNames method
func TestConfig_GetCommandNames(t *testing.T) {
	config := &Config{
		Commands: []Command{
			{Name: "command1", Alias: "c1"},
			{Name: "command2"},
			{Name: "command3", Alias: "c3"},
		},
	}

	names := config.GetCommandNames()
	expectedNames := []string{"command1", "c1", "command2", "command3", "c3"}

	if len(names) != len(expectedNames) {
		t.Errorf("Expected %d names, got %d", len(expectedNames), len(names))
	}

	// Convert to map for easier checking
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expected := range expectedNames {
		if !nameMap[expected] {
			t.Errorf("Expected name '%s' not found in result", expected)
		}
	}
}

// TestIsValidParameterType tests the isValidParameterType function
func TestIsValidParameterType(t *testing.T) {
	validTypes := []string{"string", "bool", "int", "float"}
	for _, validType := range validTypes {
		if !isValidParameterType(validType) {
			t.Errorf("Expected type '%s' to be valid", validType)
		}
	}

	invalidTypes := []string{"invalid", "array", "object", ""}
	for _, invalidType := range invalidTypes {
		if isValidParameterType(invalidType) {
			t.Errorf("Expected type '%s' to be invalid", invalidType)
		}
	}
}

// TestNewLoader tests the NewLoader constructor
func TestNewLoader(t *testing.T) {
	configPath := "/test/path"
	loader := NewLoader(configPath)
	if loader == nil {
		t.Error("NewLoader() returned nil")
		return
	}
	if loader.configPath != configPath {
		t.Errorf("Expected config path '%s', got '%s'", configPath, loader.configPath)
	}
}

// TestLoadDefault tests the LoadDefault function
func TestLoadDefault(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create a temporary directory with a commands.yml file
	tempDir := t.TempDir()
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Failed to restore working directory: %v", err)
		}
	}() // Restore original working directory

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a test commands.yml file
	validYAML := `
commands:
  - name: "test"
    description: "Test command"
    base_command: "echo"
    platforms:
      linux:
        template: "echo test"
`

	err = os.WriteFile("commands.yml", []byte(validYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write commands.yml: %v", err)
	}

	// Test LoadDefault
	config, err := LoadDefault()
	if err != nil {
		t.Errorf("LoadDefault() failed: %v", err)
	}
	if config == nil {
		t.Error("LoadDefault() returned nil config")
	}
}

// BenchmarkLoader_Load benchmarks the Load method
func BenchmarkLoader_Load(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "bench_commands.yml")

	benchYAML := `
commands:
  - name: "bench-command"
    description: "Benchmark command"
    base_command: "echo"
    params:
      - name: "message"
        type: "string"
        required: true
    platforms:
      linux:
        template: "{{.base_command}} '{{.params.message}}'"
`

	err := os.WriteFile(configPath, []byte(benchYAML), 0644)
	if err != nil {
		b.Fatalf("Failed to write benchmark config file: %v", err)
	}

	loader := NewLoader(configPath)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = loader.Load()
	}
}