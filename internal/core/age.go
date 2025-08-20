// Package core provides the business logic for SOPS team management
package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (s *SopsManager) generateAgeKey(keyPath string) (string, error) {
	if err := ensureBinaryAvailable(AgeKeygenBinary, "Please install age: https://github.com/FiloSottile/age"); err != nil {
		return "", err
	}

	cmd := exec.Command(AgeKeygenBinary)
	output, err := cmd.Output()
	if err != nil {
		return EmptyString, fmt.Errorf("failed to generate age key: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var publicKey string
	var privateKey string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "# public key: "); ok {
			publicKey = after
		} else if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			privateKey = line
		}
	}

	if privateKey == EmptyString || publicKey == EmptyString {
		return EmptyString, fmt.Errorf("failed to parse age-keygen output")
	}

	if err := os.WriteFile(keyPath, []byte(privateKey+"\n"), PrivateKeyFileMode); err != nil {
		return EmptyString, fmt.Errorf("failed to write private key: %w", err)
	}

	fmt.Printf("Generated age key pair:\n")
	fmt.Printf("  Public key:  %s\n", publicKey)
	fmt.Printf("  Private key: %s (saved)\n", keyPath)

	return publicKey, nil
}

func (s *SopsManager) getPublicKeyFromPrivateKey(keyPath string) (string, error) {
	if err := ensureBinaryAvailable(AgeKeygenBinary, "Please install age: https://github.com/FiloSottile/age"); err != nil {
		return "", err
	}

	cmd := exec.Command(AgeKeygenBinary, "-y", keyPath)
	output, err := cmd.Output()
	if err != nil {
		return EmptyString, fmt.Errorf("failed to extract public key from %s: %w", keyPath, err)
	}

	publicKey := strings.TrimSpace(string(output))
	if publicKey == EmptyString {
		return EmptyString, fmt.Errorf("failed to extract public key from %s", keyPath)
	}

	return publicKey, nil
}
