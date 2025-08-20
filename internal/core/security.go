package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// File and directory permissions
const (
	DefaultSOPSBinary = "sops"
	AgeKeygenBinary   = "age-keygen"

	PrivateKeyFileMode = 0o600 // Read/write for owner only
	BackupDirMode      = 0o700 // Read/write/execute for owner only
	GitignoreFileMode  = 0o644 // Read/write for owner, read for group/others

	HoursPerDay          = 24
	DefaultSliceCapacity = 32

	DateFormat              = "2006-01-02"
	FailedToLoadManifestMsg = "failed to load manifest: %w"

	DefaultMaxKeyAgeDays = 180
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
