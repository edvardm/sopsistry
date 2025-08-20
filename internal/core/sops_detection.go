package core

import (
	"fmt"
	"os"
	"strings"
)

// SOPSDetector checks for existing SOPS configuration
type SOPSDetector struct{}

// NewSOPSDetector creates a new SOPS detector
func NewSOPSDetector() *SOPSDetector {
	return &SOPSDetector{}
}

// DetectSOPSConfig checks for existing .sops.yaml configuration
func (d *SOPSDetector) DetectSOPSConfig() (*SOPSConfigInfo, error) {
	info := &SOPSConfigInfo{
		ConfigPath: ".sops.yaml",
		Exists:     false,
	}

	if data, err := os.ReadFile(".sops.yaml"); err == nil {
		info.Exists = true
		info.Content = string(data)
		info.HasCreationRules = strings.Contains(info.Content, "creation_rules")
		info.HasAgeKeys = strings.Contains(info.Content, "age:")
		info.HasKMSKeys = strings.Contains(info.Content, "kms:") || strings.Contains(info.Content, "arn:aws:kms")
		info.HasPGPKeys = strings.Contains(info.Content, "pgp:")
		return info, nil
	}

	if data, err := os.ReadFile(".sops.yml"); err == nil {
		info.ConfigPath = ".sops.yml"
		info.Exists = true
		info.Content = string(data)
		info.HasCreationRules = strings.Contains(info.Content, "creation_rules")
		info.HasAgeKeys = strings.Contains(info.Content, "age:")
		info.HasKMSKeys = strings.Contains(info.Content, "kms:") || strings.Contains(info.Content, "arn:aws:kms")
		info.HasPGPKeys = strings.Contains(info.Content, "pgp:")
		return info, nil
	}

	return info, nil
}

// SOPSConfigInfo contains information about existing SOPS configuration
type SOPSConfigInfo struct {
	ConfigPath       string
	Content          string
	Exists           bool
	HasCreationRules bool
	HasAgeKeys       bool
	HasKMSKeys       bool
	HasPGPKeys       bool
}

// ShouldWarn determines if we should warn about conflicts
func (info *SOPSConfigInfo) ShouldWarn() bool {
	return info.Exists && (info.HasCreationRules || info.HasAgeKeys)
}

// GetWarningMessage returns an appropriate warning message
func (info *SOPSConfigInfo) GetWarningMessage() string {
	if !info.ShouldWarn() {
		return ""
	}

	var warnings []string
	warnings = append(warnings, fmt.Sprintf("⚠️  Detected existing %s", info.ConfigPath))

	if info.HasAgeKeys {
		warnings = append(warnings, "   • Contains age keys that may conflict with team settings")
	}
	if info.HasKMSKeys {
		warnings = append(warnings, "   • Contains KMS keys (consider using sops directly for these files)")
	}
	if info.HasPGPKeys {
		warnings = append(warnings, "   • Contains PGP keys (consider using sops directly for these files)")
	}

	warnings = append(warnings, EmptyString, "💡 Options:", "   • Use 'sops' directly for files managed by .sops.yaml", "   • Remove/rename .sops.yaml for full team management", "   • Continue anyway (team settings will be used)") //nolint:gocritic // Single append is more readable here

	return strings.Join(warnings, "\n")
}

// GetCoexistenceAdvice returns advice for using both tools
func (info *SOPSConfigInfo) GetCoexistenceAdvice() string {
	if !info.Exists {
		return ""
	}

	advice := []string{
		"🔧 Coexistence recommendations:",
		"   • Use 'sistry encrypt' for team-managed files",
		"   • Use 'sops -e' directly for files with complex key requirements",
		"   • Team settings will override .sops.yaml for 'sistry' commands",
	}

	return strings.Join(advice, "\n")
}
