package core

import "fmt"

// SopsError represents different categories of SOPS-related errors
type SopsError interface {
	error
	Category() string
}

// ManifestError represents errors related to manifest operations
type ManifestError struct {
	Cause     error
	Operation string // "load", "save", "validate"
	Path      string
}

func (e *ManifestError) Error() string {
	return fmt.Sprintf("manifest %s failed for %s: %v", e.Operation, e.Path, e.Cause)
}

func (e *ManifestError) Category() string {
	return "manifest"
}

func (e *ManifestError) Unwrap() error {
	return e.Cause
}

// KeyError represents errors related to age key operations
type KeyError struct {
	Cause     error
	Operation string // "generate", "rotate", "validate"
	KeyID     string
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("key %s failed for %s: %v", e.Operation, e.KeyID, e.Cause)
}

func (e *KeyError) Category() string {
	return "key"
}

func (e *KeyError) Unwrap() error {
	return e.Cause
}

// CryptoError represents SOPS encryption/decryption errors
type CryptoError struct {
	Cause     error
	Operation string // "encrypt", "decrypt", "reencrypt"
	FilePath  string
}

func (e *CryptoError) Error() string {
	return fmt.Sprintf("crypto %s failed for %s: %v", e.Operation, e.FilePath, e.Cause)
}

func (e *CryptoError) Category() string {
	return "crypto"
}

func (e *CryptoError) Unwrap() error {
	return e.Cause
}

// Helper functions for creating typed errors
func NewManifestError(op, path string, cause error) *ManifestError {
	return &ManifestError{Operation: op, Path: path, Cause: cause}
}

func NewKeyError(op, keyID string, cause error) *KeyError {
	return &KeyError{Operation: op, KeyID: keyID, Cause: cause}
}

func NewCryptoError(op, filePath string, cause error) *CryptoError {
	return &CryptoError{Operation: op, FilePath: filePath, Cause: cause}
}
