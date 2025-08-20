package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// File and directory permissions
const (
	// DefaultSOPSBinary is the default name/path for the SOPS binary
	DefaultSOPSBinary = "sops"

	// AgeKeygenBinary is the name of the age-keygen binary
	AgeKeygenBinary = "age-keygen"

	// File permissions
	PrivateKeyFileMode = 0o600 // Read/write for owner only
	BackupDirMode      = 0o700 // Read/write/execute for owner only
	GitignoreFileMode  = 0o644 // Read/write for owner, read for group/others

	// Empty string constant to avoid repeated string literals
	EmptyString = ""

	// Time-related constants
	HoursPerDay    = 24
	DaysInTwoWeeks = 14

	// Slice capacity hints for performance
	DefaultSliceCapacity = 32

	// Date format
	DateFormat = "2006-01-02"

	// Common error messages
	FailedToLoadManifestMsg = "failed to load manifest: %w"

	// Default key age settings (days)
	DefaultMaxKeyAgeDays = 180 // 6 months
)

// ValidSOPSPath represents a validated and safe SOPS executable path
type ValidSOPSPath string

// NewValidSOPSPath creates a ValidSOPSPath after validation, returning error if invalid
func NewValidSOPSPath(path string) (ValidSOPSPath, error) {
	if strings.ContainsAny(path, ";|&$`\n\r") {
		return "", fmt.Errorf("invalid SOPS path: contains unsafe characters")
	}
	if path == DefaultSOPSBinary {
		return ValidSOPSPath(path), nil
	}
	if filepath.IsAbs(path) && strings.HasSuffix(path, "sops") {
		return ValidSOPSPath(path), nil
	}
	return "", fmt.Errorf("invalid SOPS path: must be 'sops' or absolute path ending with 'sops'")
}

// String returns the underlying path string
func (v ValidSOPSPath) String() string {
	return string(v)
}

// ValidFilePath represents a validated file path safe from directory traversal
type ValidFilePath string

// NewValidFilePath creates a ValidFilePath after validation, returning error if unsafe
func NewValidFilePath(path string) (ValidFilePath, error) {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("invalid file path: paths cannot contain '..'")
	}
	return ValidFilePath(cleanPath), nil
}

// String returns the underlying path string
func (v ValidFilePath) String() string {
	return string(v)
}

// isValidSOPSPath validates that the sops path is safe to execute
// Deprecated: use NewValidSOPSPath instead
func isValidSOPSPath(path string) bool {
	_, err := NewValidSOPSPath(path)
	return err == nil
}

// validateFilePath validates file paths to prevent directory traversal
// Deprecated: use NewValidFilePath instead
func validateFilePath(path string) error {
	_, err := NewValidFilePath(path)
	return err
}
