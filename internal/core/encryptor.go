package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureBinaryAvailable checks if a binary is available in PATH
func ensureBinaryAvailable(binaryPath, installMessage string) error {
	if _, err := exec.LookPath(binaryPath); err != nil {
		return fmt.Errorf("%s not found at %s. %s", filepath.Base(binaryPath), binaryPath, installMessage)
	}
	return nil
}

// Encryptor handles SOPS file encryption operations
type Encryptor struct {
	sopsPath string
}

// NewEncryptor creates a new encryptor instance with the given SOPS binary path
func NewEncryptor(sopsPath string) *Encryptor {
	if sopsPath == "" {
		sopsPath = DefaultSOPSBinary
	}
	cleanPath := filepath.Clean(sopsPath)
	return &Encryptor{
		sopsPath: cleanPath,
	}
}

// EncryptFile encrypts a file using SOPS with the provided age keys
func (e *Encryptor) EncryptFile(filePath string, ageKeys []string, inPlace bool, regex string) error {
	if err := e.validateEncryptionInputs(filePath); err != nil {
		return err
	}

	if err := e.checkSOPSConflicts(); err != nil {
		return err
	}

	cmd, err := e.buildEncryptCommand(filePath, ageKeys, inPlace, regex)
	if err != nil {
		return err
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sops encryption failed: %s", string(output))
	}

	e.displayEncryptionResult(filePath, inPlace, regex, output)
	return nil
}

func (e *Encryptor) validateEncryptionInputs(filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file %s does not exist: %w", filePath, err)
	}

	if err := ensureBinaryAvailable(e.sopsPath, "Please install SOPS"); err != nil {
		return err
	}

	return nil
}

func (e *Encryptor) checkSOPSConflicts() error {
	detector := NewSOPSDetector()
	sopsInfo, err := detector.DetectSOPSConfig()
	if err != nil {
		return fmt.Errorf("failed to detect SOPS configuration: %w", err)
	}

	if sopsInfo.ShouldWarn() {
		fmt.Printf("%s\n\n", sopsInfo.GetWarningMessage())
	}

	return nil
}

func (e *Encryptor) buildEncryptCommand(filePath string, ageKeys []string, inPlace bool, regex string) (*exec.Cmd, error) {
	args := e.buildSOPSArgs(filePath, inPlace, regex)

	if !isValidSOPSPath(e.sopsPath) {
		return nil, fmt.Errorf("invalid sops path: %s", e.sopsPath)
	}

	cmd := exec.Command(e.sopsPath, args...) //nolint:gosec // sopsPath validated by isValidSOPSPath()

	ageRecipients := strings.Join(ageKeys, ",")
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_RECIPIENTS=%s", ageRecipients))

	return cmd, nil
}

func (e *Encryptor) buildSOPSArgs(filePath string, inPlace bool, regex string) []string { //nolint:revive // inPlace is a legitimate CLI flag parameter
	args := []string{"-e"}
	if inPlace {
		args = append(args, "--in-place")
	}
	if regex != "" {
		args = append(args, "--encrypted-regex", regex)
	}
	args = append(args, filePath)
	return args
}

func (e *Encryptor) displayEncryptionResult(filePath string, inPlace bool, regex string, output []byte) { //nolint:revive // inPlace is a legitimate CLI flag parameter
	if inPlace {
		if regex != "" {
			fmt.Printf("🔒 Encrypted %s (partial: %s)\n", filePath, regex)
		} else {
			fmt.Printf("🔒 Encrypted %s (full file)\n", filePath)
		}
	} else {
		fmt.Print(string(output))
	}
}

// Decryptor handles SOPS decryption operations
type Decryptor struct {
	sopsPath string
}

// NewDecryptor creates a new decryptor instance
func NewDecryptor(sopsPath string) *Decryptor {
	// Validate and clean the sops path for security
	if sopsPath == "" {
		sopsPath = DefaultSOPSBinary
	}
	// Clean the path to prevent injection
	cleanPath := filepath.Clean(sopsPath)
	return &Decryptor{
		sopsPath: cleanPath,
	}
}

// DecryptFile decrypts a SOPS-encrypted file
func (d *Decryptor) DecryptFile(filePath, keyPath string, inPlace bool) error { //nolint:revive // inPlace is a legitimate CLI flag parameter
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file %s does not exist: %w", filePath, err)
	}

	// Check if key file exists
	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("age key file %s does not exist: %w", keyPath, err)
	}

	// Check if SOPS is available
	if err := ensureBinaryAvailable(d.sopsPath, "Please install SOPS"); err != nil {
		return err
	}

	// Build SOPS command
	args := []string{"-d"}
	if inPlace {
		args = append(args, "--in-place")
	}
	args = append(args, filePath)

	// Validate sopsPath for security (prevent command injection)
	if !isValidSOPSPath(d.sopsPath) {
		return fmt.Errorf("invalid sops path: %s", d.sopsPath)
	}
	cmd := exec.Command(d.sopsPath, args...) //nolint:gosec // sopsPath validated by isValidSOPSPath()

	// Set age identity file as environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_KEY_FILE=%s", keyPath))

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sops decryption failed: %s", string(output))
	}

	if inPlace {
		fmt.Printf("🔓 Decrypted %s\n", filePath)
	} else {
		fmt.Print(string(output))
	}

	return nil
}
