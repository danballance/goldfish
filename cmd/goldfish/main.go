// Package main provides the goldfish CLI application entry point.
// It creates a dynamic CLI using Cobra that generates commands from YAML configuration.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/danballance/goldfish/internal/config"
	"github.com/danballance/goldfish/internal/engine"
	"github.com/danballance/goldfish/internal/platform"
)

const (
	// Version of the goldfish CLI
	Version = "0.1.0"
	// DefaultTimeout for command execution
	DefaultTimeout = 30 * time.Second
)

// GoldfishApp holds the application state
type GoldfishApp struct {
	config           *config.Config
	engine           *engine.Engine
	platformDetector *platform.Detector
	rootCmd          *cobra.Command
}

// main is the entry point for the goldfish CLI application
func main() {
	app := &GoldfishApp{
		engine:           engine.NewEngine(DefaultTimeout),
		platformDetector: platform.NewDetector(),
	}

	// Initialize the application
	if err := app.initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Execute the root command
	if err := app.rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// initialize sets up the CLI application
func (app *GoldfishApp) initialize() error {
	// Load configuration
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	app.config = cfg

	// Create root command
	app.rootCmd = &cobra.Command{
		Use:     "goldfish",
		Short:   "Cross-platform command unification",
		Long:    "Goldfish provides unified command interfaces that work consistently across different operating systems.",
		Version: Version,
		Example: "  goldfish replace --in-place 's/foo/bar/g' file.txt\n  goldfish help replace",
	}

	// Add version flag
	app.rootCmd.SetVersionTemplate("goldfish version {{.Version}}\n")

	// Generate commands from configuration
	if err := app.generateCommands(); err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	return nil
}

// generateCommands creates Cobra commands from the YAML configuration
func (app *GoldfishApp) generateCommands() error {
	// Get current platform
	currentPlatform, err := app.platformDetector.Current()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	// Generate a command for each configured command
	for _, cmdConfig := range app.config.Commands {
		// Create a copy of cmdConfig for the closure
		cmd := cmdConfig

		// Check if command is supported on current platform
		if _, exists := cmd.Platforms[currentPlatform.String()]; !exists {
			// Skip commands not supported on this platform
			continue
		}

		// Create the Cobra command
		cobraCmd := &cobra.Command{
			Use:   cmd.Name,
			Short: cmd.Description,
			Long:  fmt.Sprintf("%s\n\nThis command provides cross-platform compatibility for '%s'.", cmd.Description, cmd.BaseCommand),
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				return app.executeCommand(&cmd, cobraCmd, args, currentPlatform)
			},
		}

		// Add alias if specified
		if cmd.Alias != "" {
			cobraCmd.Aliases = []string{cmd.Alias}
		}

		// Add flags for each parameter
		for _, param := range cmd.Parameters {
			app.addParameterFlag(cobraCmd, &param)
		}

		// Add usage examples
		if examples := app.generateExamples(&cmd); examples != "" {
			cobraCmd.Example = examples
		}

		// Add the command to the root
		app.rootCmd.AddCommand(cobraCmd)
	}

	return nil
}

// addParameterFlag adds a flag to the Cobra command based on parameter definition
func (app *GoldfishApp) addParameterFlag(cobraCmd *cobra.Command, param *config.Parameter) {
	flagName := param.Name
	if param.Flag != "" {
		// Remove leading dashes from flag specification
		flagName = strings.TrimLeft(param.Flag, "-")
	}

	description := param.Description
	if description == "" {
		description = fmt.Sprintf("%s parameter", param.Name)
	}

	// Add the appropriate flag type
	switch param.Type {
	case "string":
		defaultValue := ""
		if param.Default != nil {
			if str, ok := param.Default.(string); ok {
				defaultValue = str
			}
		}
		cobraCmd.Flags().String(flagName, defaultValue, description)
		if param.Required {
			if err := cobraCmd.MarkFlagRequired(flagName); err != nil {
				// This should rarely fail, but we handle it gracefully
				fmt.Fprintf(os.Stderr, "Warning: failed to mark flag %s as required: %v\n", flagName, err)
			}
		}
	case "bool":
		defaultValue := false
		if param.Default != nil {
			if b, ok := param.Default.(bool); ok {
				defaultValue = b
			}
		}
		cobraCmd.Flags().Bool(flagName, defaultValue, description)
	case "int":
		defaultValue := 0
		if param.Default != nil {
			if i, ok := param.Default.(int); ok {
				defaultValue = i
			}
		}
		cobraCmd.Flags().Int(flagName, defaultValue, description)
	case "float":
		defaultValue := 0.0
		if param.Default != nil {
			if f, ok := param.Default.(float64); ok {
				defaultValue = f
			}
		}
		cobraCmd.Flags().Float64(flagName, defaultValue, description)
	}
}

// executeCommand handles the execution of a goldfish command
func (app *GoldfishApp) executeCommand(cmd *config.Command, cobraCmd *cobra.Command, args []string, currentPlatform platform.SupportedPlatform) error {
	// Parse flags
	flags := make(map[string]interface{})
	for _, param := range cmd.Parameters {
		flagName := param.Name
		if param.Flag != "" {
			flagName = strings.TrimLeft(param.Flag, "-")
		}

		switch param.Type {
		case "string":
			if val, err := cobraCmd.Flags().GetString(flagName); err == nil && val != "" {
				flags["--"+flagName] = val
			}
		case "bool":
			if val, err := cobraCmd.Flags().GetBool(flagName); err == nil && val {
				flags["--"+flagName] = val
			}
		case "int":
			if val, err := cobraCmd.Flags().GetInt(flagName); err == nil && cobraCmd.Flags().Changed(flagName) {
				flags["--"+flagName] = val
			}
		case "float":
			if val, err := cobraCmd.Flags().GetFloat64(flagName); err == nil && cobraCmd.Flags().Changed(flagName) {
				flags["--"+flagName] = val
			}
		}
	}

	// Parse parameters from arguments and flags
	params, err := app.engine.ParseParameters(cmd, args, flags)
	if err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Create execution context
	ctx := &engine.ExecutionContext{
		Command:    cmd,
		Platform:   currentPlatform,
		Parameters: params,
		Timeout:    DefaultTimeout,
	}

	// Execute the command
	return app.engine.Execute(ctx)
}

// generateExamples creates usage examples for a command
func (app *GoldfishApp) generateExamples(cmd *config.Command) string {
	examples := []string{}

	// Basic example with command name
	example := fmt.Sprintf("  goldfish %s", cmd.Name)
	
	// Add parameter examples
	for _, param := range cmd.Parameters {
		if param.Required {
			switch param.Type {
			case "string":
				example += fmt.Sprintf(" <%s>", param.Name)
			case "bool":
				if param.Flag != "" {
					example += fmt.Sprintf(" %s", param.Flag)
				}
			default:
				example += fmt.Sprintf(" <%s>", param.Name)
			}
		}
	}

	examples = append(examples, example)

	// Add alias example if available
	if cmd.Alias != "" {
		aliasExample := strings.Replace(example, cmd.Name, cmd.Alias, 1)
		examples = append(examples, aliasExample)
	}

	return strings.Join(examples, "\n")
}