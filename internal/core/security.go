package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DefaultSOPSBinary is the default name/path for the SOPS binary
const DefaultSOPSBinary = "sops"

// isValidSOPSPath validates that the sops path is safe to execute
func isValidSOPSPath(path string) bool {
	if strings.ContainsAny(path, ";|&$`\n\r") {
		return false
	}
	if path == DefaultSOPSBinary {
		return true
	}
	if filepath.IsAbs(path) && strings.HasSuffix(path, "sops") {
		return true
	}
	return false
}

// validateFilePath validates file paths to prevent directory traversal
func validateFilePath(path string) error {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: paths cannot contain '..'")
	}
	return nil
}
