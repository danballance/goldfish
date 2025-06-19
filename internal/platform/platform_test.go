// Package platform_test provides unit tests for the platform detection module.
package platform

import (
	"runtime"
	"testing"
)

// TestDetector_Current tests the Current method of the Detector
func TestDetector_Current(t *testing.T) {
	detector := NewDetector()

	// Test that Current() returns a valid platform
	platform, err := detector.Current()
	if err != nil {
		t.Fatalf("Current() returned error: %v", err)
	}

	// Verify the platform matches the runtime GOOS
	expectedPlatform := ""
	switch runtime.GOOS {
	case "linux":
		expectedPlatform = "linux"
	case "darwin":
		expectedPlatform = "darwin"
	case "windows":
		expectedPlatform = "windows"
	default:
		// For unsupported platforms, we expect an error
		if err == nil {
			t.Fatalf("Expected error for unsupported platform %s, but got none", runtime.GOOS)
		}
		return
	}

	if platform.String() != expectedPlatform {
		t.Errorf("Expected platform %s, got %s", expectedPlatform, platform.String())
	}
}

// TestDetector_IsSupported tests the IsSupported method
func TestDetector_IsSupported(t *testing.T) {
	detector := NewDetector()

	// Test supported platforms
	supportedPlatforms := []string{"linux", "darwin", "windows"}
	for _, platform := range supportedPlatforms {
		if !detector.IsSupported(platform) {
			t.Errorf("Expected platform %s to be supported", platform)
		}
	}

	// Test unsupported platforms
	unsupportedPlatforms := []string{"freebsd", "openbsd", "netbsd", "solaris", "invalid"}
	for _, platform := range unsupportedPlatforms {
		if detector.IsSupported(platform) {
			t.Errorf("Expected platform %s to be unsupported", platform)
		}
	}
}

// TestDetector_GetSupportedPlatforms tests the GetSupportedPlatforms method
func TestDetector_GetSupportedPlatforms(t *testing.T) {
	detector := NewDetector()

	platforms := detector.GetSupportedPlatforms()

	// Check that we have the expected number of platforms
	expectedCount := 3
	if len(platforms) != expectedCount {
		t.Errorf("Expected %d supported platforms, got %d", expectedCount, len(platforms))
	}

	// Check that all expected platforms are present
	expectedPlatforms := map[SupportedPlatform]bool{
		Linux:   false,
		Darwin:  false,
		Windows: false,
	}

	for _, platform := range platforms {
		if _, exists := expectedPlatforms[platform]; !exists {
			t.Errorf("Unexpected platform in supported list: %s", platform)
		}
		expectedPlatforms[platform] = true
	}

	// Verify all expected platforms were found
	for platform, found := range expectedPlatforms {
		if !found {
			t.Errorf("Expected platform %s not found in supported list", platform)
		}
	}
}

// TestSupportedPlatform_String tests the String method of SupportedPlatform
func TestSupportedPlatform_String(t *testing.T) {
	testCases := []struct {
		platform SupportedPlatform
		expected string
	}{
		{Linux, "linux"},
		{Darwin, "darwin"},
		{Windows, "windows"},
	}

	for _, tc := range testCases {
		if tc.platform.String() != tc.expected {
			t.Errorf("Expected %s.String() to return %s, got %s", tc.platform, tc.expected, tc.platform.String())
		}
	}
}

// TestNewDetector tests the NewDetector constructor
func TestNewDetector(t *testing.T) {
	detector := NewDetector()
	if detector == nil {
		t.Error("NewDetector() returned nil")
	}
}

// BenchmarkDetector_Current benchmarks the Current method
func BenchmarkDetector_Current(b *testing.B) {
	detector := NewDetector()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = detector.Current()
	}
}

// BenchmarkDetector_IsSupported benchmarks the IsSupported method
func BenchmarkDetector_IsSupported(b *testing.B) {
	detector := NewDetector()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = detector.IsSupported("linux")
	}
}