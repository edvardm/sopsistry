package core

import (
	"fmt"
	"regexp"
	"strings"
)

// MemberID represents a validated team member identifier
type MemberID string

// NewMemberID creates a validated member ID
func NewMemberID(id string) (MemberID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("member ID cannot be empty")
	}
	if strings.ContainsAny(id, " \t\n\r") {
		return "", fmt.Errorf("member ID cannot contain whitespace")
	}
	return MemberID(id), nil
}

// String returns the underlying string value
func (m MemberID) String() string {
	return string(m)
}

// ScopeName represents a validated scope identifier
type ScopeName string

// NewScopeName creates a validated scope name
func NewScopeName(name string) (ScopeName, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("scope name cannot be empty")
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return "", fmt.Errorf("scope name cannot contain whitespace")
	}
	return ScopeName(name), nil
}

// String returns the underlying string value
func (s ScopeName) String() string {
	return string(s)
}

// AgePublicKey represents a validated age public key
type AgePublicKey string

var agePublicKeyRegex = regexp.MustCompile(`^age1[a-z0-9]{58}$`)

// NewAgePublicKey creates a validated age public key
func NewAgePublicKey(key string) (AgePublicKey, error) {
	key = strings.TrimSpace(key)
	if !agePublicKeyRegex.MatchString(key) {
		return "", fmt.Errorf("invalid age public key format: must match age1[58 characters]")
	}
	return AgePublicKey(key), nil
}

// String returns the underlying string value
func (a AgePublicKey) String() string {
	return string(a)
}

// AgePrivateKey represents a validated age private key
type AgePrivateKey string

var agePrivateKeyRegex = regexp.MustCompile(`^AGE-SECRET-KEY-1[A-Z0-9]{58}$`)

// NewAgePrivateKey creates a validated age private key
func NewAgePrivateKey(key string) (AgePrivateKey, error) {
	key = strings.TrimSpace(key)
	if !agePrivateKeyRegex.MatchString(key) {
		return "", fmt.Errorf("invalid age private key format: must match AGE-SECRET-KEY-1[58 characters]")
	}
	return AgePrivateKey(key), nil
}

// String returns the underlying string value - should be used carefully for logging
func (a AgePrivateKey) String() string {
	return string(a)
}

const redactedPrefixLength = 20

// Redacted returns a redacted version safe for logging
func (a AgePrivateKey) Redacted() string {
	key := string(a)
	if len(key) > redactedPrefixLength {
		return key[:redactedPrefixLength] + "***REDACTED***"
	}
	return "***REDACTED***"
}
