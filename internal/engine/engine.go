// Package engine provides command execution functionality for goldfish.
// It handles template processing, parameter validation, and command execution
// with proper stdio handling and exit code propagation.
package engine

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/danballance/goldfish/internal/config"
	"github.com/danballance/goldfish/internal/platform"
)

// ExitErrorWithCode is an error type that includes an exit code.
type ExitErrorWithCode struct {
	Err      error
	ExitCode int
}

// Error returns the error message.
func (e *ExitErrorWithCode) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *ExitErrorWithCode) Unwrap() error {
	return e.Err
}

// ExecutionContext holds the context for command execution
// It contains all the information needed to execute a command
type ExecutionContext struct {
	// Command is the command definition from the config
	Command *config.Command
	// Platform is the current platform
	Platform platform.SupportedPlatform
	// Parameters contains the parsed user parameters
	Parameters map[string]interface{}
	// Timeout specifies the maximum execution time
	Timeout time.Duration
}

// Engine handles command execution and template rendering
type Engine struct {
	platformDetector *platform.Detector
	timeout          time.Duration
}

// NewEngine creates a new command execution engine
// timeout specifies the default timeout for command execution
func NewEngine(timeout time.Duration) *Engine {
	return &Engine{
		platformDetector: platform.NewDetector(),
		timeout:          timeout,
	}
}

// Execute runs a command with the given parameters
// It validates parameters, renders the template, and executes the resulting command
func (e *Engine) Execute(ctx *ExecutionContext) error {
	// Validate the execution context
	if err := e.validateContext(ctx); err != nil {
		return fmt.Errorf("invalid execution context: %w", err)
	}

	// Get the platform-specific template
	platformCmd, exists := ctx.Command.Platforms[ctx.Platform.String()]
	if !exists {
		return fmt.Errorf("command '%s' not supported on platform '%s'", ctx.Command.Name, ctx.Platform)
	}

	// Render the command template
	renderedCmd, err := e.renderTemplate(ctx.Command, &platformCmd, ctx.Parameters)
	if err != nil {
		return fmt.Errorf("failed to render command template: %w", err)
	}

	// Execute the rendered command
	return e.executeCommand(renderedCmd, ctx.Timeout)
}

// validateContext validates the execution context
func (e *Engine) validateContext(ctx *ExecutionContext) error {
	if ctx.Command == nil {
		return fmt.Errorf("command is nil")
	}
	if ctx.Parameters == nil {
		return fmt.Errorf("parameters map is nil")
	}

	// Validate required parameters
	for _, param := range ctx.Command.Parameters {
		if param.Required {
			if _, exists := ctx.Parameters[param.Name]; !exists {
				return fmt.Errorf("required parameter '%s' is missing", param.Name)
			}
		}
	}

	// Validate parameter types
	for paramName, paramValue := range ctx.Parameters {
		// Find the parameter definition
		var paramDef *config.Parameter
		for _, p := range ctx.Command.Parameters {
			if p.Name == paramName {
				paramDef = &p
				break
			}
		}

		if paramDef == nil {
			return fmt.Errorf("unknown parameter: %s", paramName)
		}

		// Validate parameter type
		if err := e.validateParameterType(paramDef, paramValue); err != nil {
			return fmt.Errorf("parameter '%s': %w", paramName, err)
		}
	}

	return nil
}

// validateParameterType validates that a parameter value matches its expected type
func (e *Engine) validateParameterType(param *config.Parameter, value interface{}) error {
	switch param.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case "int":
		switch v := value.(type) {
		case int, int32, int64:
			// Already an integer type
		case string:
			// Try to parse string as integer
			if _, err := strconv.Atoi(v); err != nil {
				return fmt.Errorf("expected int, got unparseable string: %s", v)
			}
		default:
			return fmt.Errorf("expected int, got %T", value)
		}
	case "float":
		switch v := value.(type) {
		case float32, float64:
			// Already a float type
		case string:
			// Try to parse string as float
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				return fmt.Errorf("expected float, got unparseable string: %s", v)
			}
		default:
			return fmt.Errorf("expected float, got %T", value)
		}
	default:
		return fmt.Errorf("unsupported parameter type: %s", param.Type)
	}
	return nil
}

// renderTemplate renders the command template with the given parameters
func (e *Engine) renderTemplate(cmd *config.Command, platformCmd *config.PlatformCommand, params map[string]interface{}) (string, error) {
	// Create template data
	templateData := map[string]interface{}{
		"base_command": cmd.BaseCommand,
		"params":       params,
	}

	// Parse the template
	tmpl, err := template.New("command").Parse(platformCmd.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// executeCommand executes the rendered command using the system shell
func (e *Engine) executeCommand(command string, timeout time.Duration) error {
	// Use the specified timeout or fall back to the engine default
	if timeout == 0 {
		timeout = e.timeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Prepare the command
	// On Unix systems, use sh -c to execute the command
	// This allows for complex shell commands with pipes, redirects, etc.
	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// Connect stdio to allow interactive commands and proper output handling
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command
	err := cmd.Run()
	
	// Handle different types of errors
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out after %v: %s", timeout, command)
		}
		
		// For exit code errors, we want to preserve the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			return &ExitErrorWithCode{
				Err:      fmt.Errorf("command failed with exit code %d", exitError.ExitCode()),
				ExitCode: exitError.ExitCode(),
			}
		}
		
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// isWindows checks if the current platform is Windows
func isWindows() bool {
	detector := platform.NewDetector()
	currentPlatform, err := detector.Current()
	if err != nil {
		return false
	}
	return currentPlatform == platform.Windows
}

// ParseParameters parses command line arguments into a parameter map
// This function takes the raw arguments and converts them according to parameter definitions
func (e *Engine) ParseParameters(cmd *config.Command, args []string, flags map[string]interface{}) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	
	// Process flags first
	for flagName, flagValue := range flags {
		// Find the parameter that corresponds to this flag
		for _, param := range cmd.Parameters {
			if param.Flag == flagName || param.Flag == "--"+flagName {
				params[param.Name] = flagValue
				break
			}
		}
	}
	
	// Process positional arguments
	argIndex := 0
	for _, param := range cmd.Parameters {
		// Skip parameters that were set via flags
		if _, exists := params[param.Name]; exists {
			continue
		}
		
		// Switch on argument availability to improve readability
		switch {
		case argIndex < len(args):
			// Convert the argument to the appropriate type
			convertedValue, err := e.convertArgument(args[argIndex], param.Type)
			if err != nil {
				return nil, fmt.Errorf("parameter '%s': %w", param.Name, err)
			}
			params[param.Name] = convertedValue
			argIndex++
		case param.Required:
			return nil, fmt.Errorf("required parameter '%s' not provided", param.Name)
		case param.Default != nil:
			params[param.Name] = param.Default
		}
	}
	
	return params, nil
}

// convertArgument converts a string argument to the specified type
func (e *Engine) convertArgument(arg, paramType string) (interface{}, error) {
	switch paramType {
	case "string":
		return arg, nil
	case "bool":
		return strconv.ParseBool(arg)
	case "int":
		return strconv.Atoi(arg)
	case "float":
		return strconv.ParseFloat(arg, 64)
	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", paramType)
	}
}