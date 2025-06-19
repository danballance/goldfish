// Package platform provides cross-platform OS detection functionality.
// It abstracts the runtime OS detection to provide consistent platform
// identification for command template selection.
package platform

import (
	"fmt"
	"runtime"
)

// SupportedPlatform represents the platforms that goldfish supports
type SupportedPlatform string

const (
	// Linux represents GNU/Linux systems
	Linux SupportedPlatform = "linux"
	// Darwin represents macOS systems (BSD-based)
	Darwin SupportedPlatform = "darwin"
	// Windows represents Windows systems
	Windows SupportedPlatform = "windows"
)

// String returns the string representation of the platform
func (p SupportedPlatform) String() string {
	return string(p)
}

// Detector provides methods for detecting the current platform
type Detector struct{}

// NewDetector creates a new platform detector instance
func NewDetector() *Detector {
	return &Detector{}
}

// Current returns the current platform based on runtime.GOOS
// It maps Go's GOOS values to goldfish's supported platforms
func (d *Detector) Current() (SupportedPlatform, error) {
	switch runtime.GOOS {
	case "linux":
		return Linux, nil
	case "darwin":
		return Darwin, nil
	case "windows":
		return Windows, nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// IsSupported checks if the given platform string is supported
func (d *Detector) IsSupported(platform string) bool {
	switch SupportedPlatform(platform) {
	case Linux, Darwin, Windows:
		return true
	default:
		return false
	}
}

// GetSupportedPlatforms returns a slice of all supported platforms
func (d *Detector) GetSupportedPlatforms() []SupportedPlatform {
	return []SupportedPlatform{Linux, Darwin, Windows}
}