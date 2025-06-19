// Package engine_test provides unit tests for the command execution engine.
package engine

import (
	"testing"
	"time"

	"github.com/danballance/goldfish/internal/config"
	"github.com/danballance/goldfish/internal/platform"
)

// TestNewEngine tests the NewEngine constructor
func TestNewEngine(t *testing.T) {
	timeout := 10 * time.Second
	engine := NewEngine(timeout)

	if engine == nil {
		t.Error("NewEngine() returned nil")
		return
	}
	if engine.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, engine.timeout)
	}
	if engine.platformDetector == nil {
		t.Error("Expected platform detector to be initialized")
	}
}

// TestEngine_validateContext tests the validateContext method
func TestEngine_validateContext(t *testing.T) {
	engine := NewEngine(time.Second)

	// Test valid context
	validCmd := &config.Command{
		Name:        "test",
		BaseCommand: "echo",
		Parameters: []config.Parameter{
			{Name: "message", Type: "string", Required: true},
			{Name: "verbose", Type: "bool", Required: false},
		},
	}
	validParams := map[string]interface{}{
		"message": "hello",
		"verbose": false,
	}
	validCtx := &ExecutionContext{
		Command:    validCmd,
		Platform:   platform.Linux,
		Parameters: validParams,
	}

	err := engine.validateContext(validCtx)
	if err != nil {
		t.Errorf("Expected valid context to pass validation, got error: %v", err)
	}

	// Test nil command
	invalidCtx := &ExecutionContext{
		Command:    nil,
		Platform:   platform.Linux,
		Parameters: validParams,
	}
	err = engine.validateContext(invalidCtx)
	if err == nil {
		t.Error("Expected error for nil command")
	}

	// Test nil parameters
	invalidCtx = &ExecutionContext{
		Command:    validCmd,
		Platform:   platform.Linux,
		Parameters: nil,
	}
	err = engine.validateContext(invalidCtx)
	if err == nil {
		t.Error("Expected error for nil parameters")
	}

	// Test missing required parameter
	invalidParams := map[string]interface{}{
		"verbose": true,
		// missing required "message" parameter
	}
	invalidCtx = &ExecutionContext{
		Command:    validCmd,
		Platform:   platform.Linux,
		Parameters: invalidParams,
	}
	err = engine.validateContext(invalidCtx)
	if err == nil {
		t.Error("Expected error for missing required parameter")
	}
}

// TestEngine_validateParameterType tests the validateParameterType method
func TestEngine_validateParameterType(t *testing.T) {
	engine := NewEngine(time.Second)

	testCases := []struct {
		param    config.Parameter
		value    interface{}
		expected bool // true if should pass validation
	}{
		// String type tests
		{config.Parameter{Type: "string"}, "hello", true},
		{config.Parameter{Type: "string"}, 123, false},
		{config.Parameter{Type: "string"}, true, false},

		// Bool type tests
		{config.Parameter{Type: "bool"}, true, true},
		{config.Parameter{Type: "bool"}, false, true},
		{config.Parameter{Type: "bool"}, "true", false},
		{config.Parameter{Type: "bool"}, 1, false},

		// Int type tests
		{config.Parameter{Type: "int"}, 123, true},
		{config.Parameter{Type: "int"}, int32(123), true},
		{config.Parameter{Type: "int"}, int64(123), true},
		{config.Parameter{Type: "int"}, "123", true},
		{config.Parameter{Type: "int"}, "abc", false},
		{config.Parameter{Type: "int"}, 12.3, false},

		// Float type tests
		{config.Parameter{Type: "float"}, 12.3, true},
		{config.Parameter{Type: "float"}, float32(12.3), true},
		{config.Parameter{Type: "float"}, "12.3", true},
		{config.Parameter{Type: "float"}, "abc", false},
		{config.Parameter{Type: "float"}, true, false},

		// Invalid type
		{config.Parameter{Type: "invalid"}, "test", false},
	}

	for i, tc := range testCases {
		err := engine.validateParameterType(&tc.param, tc.value)
		if tc.expected && err != nil {
			t.Errorf("Test case %d: Expected validation to pass, got error: %v", i, err)
		}
		if !tc.expected && err == nil {
			t.Errorf("Test case %d: Expected validation to fail, but it passed", i)
		}
	}
}

// TestEngine_renderTemplate tests the renderTemplate method
func TestEngine_renderTemplate(t *testing.T) {
	engine := NewEngine(time.Second)

	cmd := &config.Command{
		Name:        "test",
		BaseCommand: "echo",
		Parameters: []config.Parameter{
			{Name: "message", Type: "string"},
			{Name: "verbose", Type: "bool"},
		},
	}

	platformCmd := &config.PlatformCommand{
		Template: "{{.base_command}}{{if .params.verbose}} -v{{end}} '{{.params.message}}'",
	}

	params := map[string]interface{}{
		"message": "hello world",
		"verbose": true,
	}

	result, err := engine.renderTemplate(cmd, platformCmd, params)
	if err != nil {
		t.Fatalf("renderTemplate() failed: %v", err)
	}

	expected := "echo -v 'hello world'"
	if result != expected {
		t.Errorf("Expected rendered command '%s', got '%s'", expected, result)
	}

	// Test with verbose = false
	params["verbose"] = false
	result, err = engine.renderTemplate(cmd, platformCmd, params)
	if err != nil {
		t.Fatalf("renderTemplate() failed: %v", err)
	}

	expected := "echo 'hello world'"
	if result != expected {
		t.Errorf("Expected rendered command '%s', got '%s'", expected, result)
	}
}

// TestEngine_renderTemplate_InvalidTemplate tests rendering with invalid template
func TestEngine_renderTemplate_InvalidTemplate(t *testing.T) {
	engine := NewEngine(time.Second)

	cmd := &config.Command{
		BaseCommand: "echo",
	}

	// Invalid template syntax
	platformCmd := &config.PlatformCommand{
		Template: "{{.base_command} {{if .params.invalid}}", // missing closing brace
	}

	params := map[string]interface{}{}

	_, err := engine.renderTemplate(cmd, platformCmd, params)
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}

// TestEngine_ParseParameters tests the ParseParameters method
func TestEngine_ParseParameters(t *testing.T) {
	engine := NewEngine(time.Second)

	cmd := &config.Command{
		Parameters: []config.Parameter{
			{Name: "message", Type: "string", Required: true},
			{Name: "count", Type: "int", Required: false, Default: 1},
			{Name: "verbose", Type: "bool", Flag: "--verbose"},
		},
	}

	// Test parsing positional arguments
	args := []string{"hello world", "5"}
	flags := map[string]interface{}{
		"--verbose": true,
	}

	params, err := engine.ParseParameters(cmd, args, flags)
	if err != nil {
		t.Fatalf("ParseParameters() failed: %v", err)
	}

	if params["message"] != "hello world" {
		t.Errorf("Expected message 'hello world', got %v", params["message"])
	}
	if params["count"] != 5 {
		t.Errorf("Expected count 5, got %v", params["count"])
	}
	if params["verbose"] != true {
		t.Errorf("Expected verbose true, got %v", params["verbose"])
	}
}

// TestEngine_convertArgument tests the convertArgument method
func TestEngine_convertArgument(t *testing.T) {
	engine := NewEngine(time.Second)

	testCases := []struct {
		arg       string
		paramType string
		expected  interface{}
		shouldErr bool
	}{
		{"hello", "string", "hello", false},
		{"123", "int", 123, false},
		{"abc", "int", nil, true},
		{"12.3", "float", 12.3, false},
		{"abc", "float", nil, true},
		{"true", "bool", true, false},
		{"false", "bool", false, false},
		{"invalid", "bool", nil, true},
		{"test", "invalid", nil, true},
	}

	for i, tc := range testCases {
		result, err := engine.convertArgument(tc.arg, tc.paramType)
		if tc.shouldErr && err == nil {
			t.Errorf("Test case %d: Expected error but got none", i)
			continue
		}
		if !tc.shouldErr && err != nil {
			t.Errorf("Test case %d: Unexpected error: %v", i, err)
			continue
		}
		if !tc.shouldErr && result != tc.expected {
			t.Errorf("Test case %d: Expected %v, got %v", i, tc.expected, result)
		}
	}
}

// TestExecutionContext tests the ExecutionContext struct
func TestExecutionContext(t *testing.T) {
	cmd := &config.Command{
		Name:        "test",
		BaseCommand: "echo",
	}
	platform := platform.Linux
	params := map[string]interface{}{"test": "value"}
	timeout := 5 * time.Second

	ctx := &ExecutionContext{
		Command:    cmd,
		Platform:   platform,
		Parameters: params,
		Timeout:    timeout,
	}

	if ctx.Command != cmd {
		t.Error("ExecutionContext.Command not set correctly")
	}
	if ctx.Platform != platform {
		t.Error("ExecutionContext.Platform not set correctly")
	}
	if ctx.Parameters["test"] != "value" {
		t.Error("ExecutionContext.Parameters not set correctly")
	}
	if ctx.Timeout != timeout {
		t.Error("ExecutionContext.Timeout not set correctly")
	}
}

// BenchmarkEngine_validateContext benchmarks the validateContext method
func BenchmarkEngine_validateContext(b *testing.B) {
	engine := NewEngine(time.Second)
	cmd := &config.Command{
		Name:        "test",
		BaseCommand: "echo",
		Parameters: []config.Parameter{
			{Name: "message", Type: "string", Required: true},
		},
	}
	params := map[string]interface{}{
		"message": "test message",
	}
	ctx := &ExecutionContext{
		Command:    cmd,
		Platform:   platform.Linux,
		Parameters: params,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.validateContext(ctx)
	}
}

// BenchmarkEngine_renderTemplate benchmarks the renderTemplate method
func BenchmarkEngine_renderTemplate(b *testing.B) {
	engine := NewEngine(time.Second)
	cmd := &config.Command{
		BaseCommand: "echo",
	}
	platformCmd := &config.PlatformCommand{
		Template: "{{.base_command}} '{{.params.message}}'",
	}
	params := map[string]interface{}{
		"message": "benchmark test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.renderTemplate(cmd, platformCmd, params)
	}
}