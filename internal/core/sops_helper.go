package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SOPSHelper provides utilities for working with SOPS commands
type SOPSHelper struct {
	sopsPath   string
	secretsDir string
}

// NewSOPSHelper creates a new SOPS er instance
func NewSOPSHelper(sopsPath, secretsDir string) *SOPSHelper {
	if sopsPath == "" {
		sopsPath = "sops"
	}
	cleanPath := filepath.Clean(sopsPath)
	return &SOPSHelper{
		sopsPath:   cleanPath,
		secretsDir: secretsDir,
	}
}

// ShowCommand displays or executes a SOPS command with proper environment
func (h *SOPSHelper) ShowCommand(args []string, ageKeys []string, execute bool) error {
	keyPath := filepath.Join(h.secretsDir, "key.txt")
	ageRecipients := strings.Join(ageKeys, ",")

	envVars := []string{
		fmt.Sprintf("SOPS_AGE_RECIPIENTS=%s", ageRecipients),
		fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyPath),
	}

	if execute {
		if !isValidSOPSPath(h.sopsPath) {
			return fmt.Errorf("invalid sops path: %s", h.sopsPath)
		}
		cmd := exec.Command(h.sopsPath, args...) //nolint:gosec // sopsPath validated by isValidSOPSPath()

		cmd.Env = append(os.Environ(), envVars...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Printf("🔧 Executing: %s %s\n", h.sopsPath, strings.Join(args, " "))
		return cmd.Run()
	}

	fmt.Printf("🔧 SOPS command with team environment:\n\n")

	for _, env := range envVars {
		fmt.Printf("export %s\n", env)
	}

	fmt.Printf("%s %s\n", h.sopsPath, strings.Join(args, " "))

	fmt.Printf("\n💡 Common partial encryption examples:\n")
	fmt.Printf("# Encrypt only password/key fields in .env:\n")
	fmt.Printf("%s -e --encrypted-regex '^(.*password.*|.*key.*)$' .env\n\n", h.sopsPath)

	fmt.Printf("# Encrypt specific fields in YAML:\n")
	fmt.Printf("%s -e --encrypted-regex '^(password|secret|key)$' config.yaml\n\n", h.sopsPath)

	fmt.Printf("# Decrypt file:\n")
	fmt.Printf("%s -d encrypted-file.yaml\n", h.sopsPath)

	return nil
}
