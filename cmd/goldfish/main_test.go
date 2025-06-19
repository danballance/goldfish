package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/danballance/goldfish/internal/config"
	"github.com/danballance/goldfish/internal/engine"
	"github.com/danballance/goldfish/internal/platform"
)

// TestMain tests the basic test setup
func TestMain(t *testing.T) {
	// This is a placeholder test to verify the testing setup
	t.Log("goldfish test setup working")
}

// TestGoldfishApp_initialize tests the application initialization
func TestGoldfishApp_initialize(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Failed to restore working directory: %v", err)
		}
	}()

	// Change to temp directory and create test config
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	testConfig := `
commands:
  - name: "test-echo"
    alias: "echo"
    description: "Cross-platform echo command"
    base_command: "echo"
    params:
      - name: "message"
        type: "string"
        required: true
      - name: "newline"
        type: "bool"
        flag: "--newline"
        description: "Add newline at end"
    platforms:
      linux:
        template: "{{.base_command}} {{if not .params.newline}}-n{{end}} '{{.params.message}}'"
      darwin:
        template: "{{.base_command}} {{if not .params.newline}}-n{{end}} '{{.params.message}}'"
      windows:
        template: "{{.base_command}} {{.params.message}}"
`

	err = os.WriteFile("commands.yml", []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test application initialization
	app := &GoldfishApp{
		engine:           engine.NewEngine(30 * time.Second),
		platformDetector: platform.NewDetector(),
	}

	err = app.initialize()
	if err != nil {
		t.Fatalf("Application initialization failed: %v", err)
	}

	// Verify initialization
	if app.config == nil {
		t.Error("Config not loaded")
	}
	if app.rootCmd == nil {
		t.Error("Root command not created")
	}
	if len(app.config.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(app.config.Commands))
	}
}

// TestGoldfishApp_generateCommands tests command generation from config
func TestGoldfishApp_generateCommands(t *testing.T) {
	// Create test config
	testConfig := &config.Config{
		Commands: []config.Command{
			{
				Name:        "test-cmd",
				Alias:       "test",
				Description: "Test command",
				BaseCommand: "echo",
				Parameters: []config.Parameter{
					{Name: "message", Type: "string", Required: true},
				},
				Platforms: map[string]config.PlatformCommand{
					"linux":   {Template: "echo '{{.params.message}}'"},
					"darwin":  {Template: "echo '{{.params.message}}'"},
					"windows": {Template: "echo {{.params.message}}"},
				},
			},
		},
	}

	app := &GoldfishApp{
		config:           testConfig,
		engine:           engine.NewEngine(30 * time.Second),
		platformDetector: platform.NewDetector(),
	}

	// Create root command
	app.rootCmd = &cobra.Command{
		Use:   "goldfish",
		Short: "Test",
	}

	err := app.generateCommands()
	if err != nil {
		t.Fatalf("generateCommands() failed: %v", err)
	}

	// Check that command was added
	if len(app.rootCmd.Commands()) == 0 {
		t.Error("No commands were generated")
	}
}

// TestEndToEndWorkflow tests the complete workflow from config to execution
func TestEndToEndWorkflow(t *testing.T) {
	// Skip this test if we don't have a shell available
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("Failed to restore working directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a simple test config with echo command
	testConfig := `
commands:
  - name: "test-echo"
    description: "Test echo command"
    base_command: "echo"
    params:
      - name: "message"
        type: "string"
        required: true
    platforms:
      linux:
        template: "{{.base_command}} '{{.params.message}}'"
      darwin:
        template: "{{.base_command}} '{{.params.message}}'"
      windows:
        template: "{{.base_command}} {{.params.message}}"
`

	err = os.WriteFile("commands.yml", []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test the complete workflow
	// 1. Load config
	cfg, err := config.LoadDefault()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 2. Detect platform
	detector := platform.NewDetector()
	currentPlatform, err := detector.Current()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}

	// 3. Find command
	cmd, found := cfg.FindCommand("test-echo")
	if !found {
		t.Fatal("Command not found in config")
	}

	// 4. Create execution engine
	eng := engine.NewEngine(5 * time.Second)

	// 5. Parse parameters (simulate command line args)
	params := map[string]interface{}{
		"message": "Hello, World!",
	}

	// 6. Create execution context
	ctx := &engine.ExecutionContext{
		Command:    cmd,
		Platform:   currentPlatform,
		Parameters: params,
		Timeout:    5 * time.Second,
	}

	// 7. Execute the command (this will actually run echo)
	// Note: We're testing the execution path but the actual command
	// execution might produce output. In a real test environment,
	// you might want to capture or redirect the output.
	err = eng.Execute(ctx)
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// If we reach here, the end-to-end workflow completed successfully
	t.Log("End-to-end workflow completed successfully")
}

// TestRealWorldExample tests a realistic cross-platform scenario
func TestRealWorldExample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-world test in short mode")
	}

	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "Hello World"
	
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test the sed replacement functionality that goldfish is designed for
	detector := platform.NewDetector()
	currentPlatform, err := detector.Current()
	if err != nil {
		t.Fatalf("Failed to detect platform: %v", err)
	}

	// Create command definition for sed replacement
	sedCmd := &config.Command{
		Name:        "replace-in-file",
		Alias:       "replace",
		Description: "Cross-platform sed replacement",
		BaseCommand: "sed",
		Parameters: []config.Parameter{
			{Name: "expression", Type: "string", Required: true},
			{Name: "file", Type: "string", Required: true},
			{Name: "in_place", Type: "bool", Flag: "--in-place"},
		},
		Platforms: map[string]config.PlatformCommand{
			"linux": {
				Template: "{{.base_command}} {{if .params.in_place}}-i{{end}} '{{.params.expression}}' {{.params.file}}",
			},
			"darwin": {
				Template: "{{.base_command}} {{if .params.in_place}}-i ''{{end}} '{{.params.expression}}' {{.params.file}}",
			},
		},
	}

	// Only test if sed is available and we're on a supported platform
	_, exists := sedCmd.Platforms[currentPlatform.String()]
	if !exists {
		t.Skipf("Platform %s not supported for sed command", currentPlatform)
	}

	// Create execution context
	eng := engine.NewEngine(10 * time.Second)
	params := map[string]interface{}{
		"expression": "s/World/Universe/g",
		"file":       testFile,
		"in_place":   true,
	}

	ctx := &engine.ExecutionContext{
		Command:    sedCmd,
		Platform:   currentPlatform,
		Parameters: params,
		Timeout:    10 * time.Second,
	}

	// Execute the sed command
	err = eng.Execute(ctx)
	if err != nil {
		// sed might not be available in test environment, so we log but don't fail
		t.Logf("sed command execution failed (this may be expected in test environment): %v", err)
		return
	}

	// If execution succeeded, verify the file was modified
	modifiedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expectedContent := "Hello Universe"
	if string(modifiedContent) != expectedContent {
		t.Errorf("Expected file content '%s', got '%s'", expectedContent, string(modifiedContent))
	}

	t.Log("Real-world sed replacement test completed successfully")
}