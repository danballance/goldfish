// Package config provides tests for the defaults functionality
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDefaults tests the embedded defaults loading functionality
func TestLoadDefaults(t *testing.T) {
	// Load embedded defaults
	config, err := LoadDefaults()
	if err != nil {
		t.Fatalf("Failed to load defaults: %v", err)
	}

	// Verify we have some commands
	if len(config.Commands) == 0 {
		t.Error("Expected embedded defaults to contain commands")
	}

	// Verify command structure
	for _, cmd := range config.Commands {
		if cmd.Name == "" {
			t.Error("Command name should not be empty")
		}
		if cmd.BaseCommand == "" {
			t.Error("Base command should not be empty")
		}
		if len(cmd.Platforms) == 0 {
			t.Error("Command should have at least one platform")
		}
	}
}

// TestMergeConfigs tests the configuration merging functionality
func TestMergeConfigs(t *testing.T) {
	// Create base config
	base := &Config{
		Commands: []Command{
			{
				Name:        "test-cmd1",
				BaseCommand: "base1",
				Platforms: map[string]PlatformCommand{
					"linux": {Template: "base1 template"},
				},
			},
			{
				Name:        "test-cmd2",
				BaseCommand: "base2",
				Platforms: map[string]PlatformCommand{
					"linux": {Template: "base2 template"},
				},
			},
		},
	}

	// Create override config
	override := &Config{
		Commands: []Command{
			{
				Name:        "test-cmd1", // Override existing command
				BaseCommand: "override1",
				Platforms: map[string]PlatformCommand{
					"linux": {Template: "override1 template"},
				},
			},
			{
				Name:        "test-cmd3", // New command
				BaseCommand: "new1",
				Platforms: map[string]PlatformCommand{
					"linux": {Template: "new1 template"},
				},
			},
		},
	}

	// Merge configs
	merged := MergeConfigs(base, override)

	// Should have 3 commands total
	if len(merged.Commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(merged.Commands))
	}

	// Check that test-cmd1 was overridden
	cmd1, found := merged.FindCommand("test-cmd1")
	if !found {
		t.Error("test-cmd1 should be found in merged config")
	}
	if cmd1.BaseCommand != "override1" {
		t.Errorf("test-cmd1 should be overridden, got base command: %s", cmd1.BaseCommand)
	}

	// Check that test-cmd2 was preserved from base
	cmd2, found := merged.FindCommand("test-cmd2")
	if !found {
		t.Error("test-cmd2 should be found in merged config")
	}
	if cmd2.BaseCommand != "base2" {
		t.Errorf("test-cmd2 should be from base, got base command: %s", cmd2.BaseCommand)
	}

	// Check that test-cmd3 was added from override
	cmd3, found := merged.FindCommand("test-cmd3")
	if !found {
		t.Error("test-cmd3 should be found in merged config")
	}
	if cmd3.BaseCommand != "new1" {
		t.Errorf("test-cmd3 should be from override, got base command: %s", cmd3.BaseCommand)
	}
}

// TestMergeConfigsWithNil tests merging with nil configs
func TestMergeConfigsWithNil(t *testing.T) {
	base := &Config{
		Commands: []Command{
			{Name: "test", BaseCommand: "test", Platforms: map[string]PlatformCommand{"linux": {Template: "test"}}},
		},
	}

	// Test with nil override
	merged := MergeConfigs(base, nil)
	if merged != base {
		t.Error("Merging with nil override should return base config")
	}

	// Test with nil base
	merged = MergeConfigs(nil, base)
	if merged != base {
		t.Error("Merging with nil base should return override config")
	}

	// Test with both nil
	merged = MergeConfigs(nil, nil)
	if merged != nil {
		t.Error("Merging with both nil should return nil")
	}
}

// TestLoadWithDefaults tests loading with embedded defaults and runtime override
func TestLoadWithDefaults(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "goldfish-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-existent runtime config (should use defaults only)
	config, err := LoadWithDefaults(filepath.Join(tempDir, "nonexistent.yml"))
	if err != nil {
		t.Fatalf("Failed to load with defaults: %v", err)
	}

	// Should have default commands
	if len(config.Commands) == 0 {
		t.Error("Expected config to have default commands")
	}

	// Create a runtime config file
	runtimeConfig := `commands:
  - name: "runtime-cmd"
    description: "Runtime command"
    base_command: "runtime"
    platforms:
      linux:
        template: "runtime template"
`
	runtimePath := filepath.Join(tempDir, "runtime.yml")
	if err := os.WriteFile(runtimePath, []byte(runtimeConfig), 0644); err != nil {
		t.Fatalf("Failed to write runtime config: %v", err)
	}

	// Test with runtime config
	config, err = LoadWithDefaults(runtimePath)
	if err != nil {
		t.Fatalf("Failed to load with runtime config: %v", err)
	}

	// Should have both default and runtime commands
	runtimeCmd, found := config.FindCommand("runtime-cmd")
	if !found {
		t.Error("Runtime command should be found in merged config")
	}
	if runtimeCmd.BaseCommand != "runtime" {
		t.Errorf("Runtime command should have correct base command, got: %s", runtimeCmd.BaseCommand)
	}
}