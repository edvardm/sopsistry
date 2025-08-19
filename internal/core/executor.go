package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Executor handles the actual execution of planned SOPS operations
type Executor struct {
	sopsPath string
}

// NewExecutor creates a new executor instance
func NewExecutor(sopsPath string) *Executor {
	if sopsPath == "" {
		sopsPath = "sops"
	}
	cleanPath := filepath.Clean(sopsPath)
	return &Executor{
		sopsPath: cleanPath,
	}
}

// Execute runs all actions in the plan atomically
func (e *Executor) Execute(plan *Plan) error {
	if len(plan.Actions) == 0 {
		fmt.Println("No actions to execute")
		return nil
	}

	backupDir, err := e.setupBackupDirectory()
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(backupDir) }()

	return e.executeActionsWithRollback(plan, backupDir)
}

func (e *Executor) setupBackupDirectory() (string, error) {
	backupDir := ".sopsistry-backup"
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	return backupDir, nil
}

func (e *Executor) executeActionsWithRollback(plan *Plan, backupDir string) error {
	executedActions := 0

	for i, action := range plan.Actions {
		if action.Type == ActionSkip {
			continue
		}

		if err := e.backupFileIfExists(action.File, backupDir, i); err != nil {
			return err
		}

		if err := e.executeAction(action); err != nil {
			return e.handleExecutionError(action, err, plan.Actions[:executedActions+1], backupDir)
		}

		executedActions++
		fmt.Printf("✓ %s %s\n", action.Type, action.File)
	}

	fmt.Printf("\nSuccessfully applied %d changes\n", executedActions)
	return nil
}

func (e *Executor) backupFileIfExists(filePath, backupDir string, index int) error {
	if _, err := os.Stat(filePath); err == nil {
		backupPath := filepath.Join(backupDir, fmt.Sprintf("%d-%s", index, filepath.Base(filePath)))
		if err := e.copyFile(filePath, backupPath); err != nil {
			return fmt.Errorf("failed to backup %s: %w", filePath, err)
		}
	}
	return nil
}

func (e *Executor) handleExecutionError(action Action, actionErr error, executedActions []Action, backupDir string) error {
	fmt.Printf("Error executing action for %s: %v\n", action.File, actionErr)
	fmt.Println("Rolling back changes...")

	if rollbackErr := e.rollback(executedActions, backupDir); rollbackErr != nil {
		return fmt.Errorf("execution failed and rollback failed: %w (original error: %w)", rollbackErr, actionErr)
	}

	return fmt.Errorf("execution failed: %w", actionErr)
}

// executeAction performs a single SOPS operation
func (e *Executor) executeAction(action Action) error {
	switch action.Type {
	case ActionEncrypt:
		return e.encryptFile(action.File, action.Recipients)
	case ActionReencrypt:
		return e.reencryptFile(action.File, action.Recipients)
	case ActionSkip:
		return nil // Skip action, nothing to do
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// encryptFile encrypts a new file with SOPS
func (e *Executor) encryptFile(file string, recipients []string) error {
	if !isValidSOPSPath(e.sopsPath) {
		return fmt.Errorf("invalid sops path: %s", e.sopsPath)
	}

	// Use environment variable to specify age recipients
	cmd := exec.Command(e.sopsPath, "-e", "--in-place", file) //nolint:gosec // sopsPath validated by isValidSOPSPath()
	cmd.Env = append(os.Environ(), fmt.Sprintf("SOPS_AGE_RECIPIENTS=%s", strings.Join(recipients, ",")))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sops encrypt failed: %s", string(output))
	}

	return nil
}

// reencryptFile re-encrypts an existing SOPS file with new recipients using native rotate
func (e *Executor) reencryptFile(file string, recipients []string) error {
	if !isValidSOPSPath(e.sopsPath) {
		return fmt.Errorf("invalid sops path: %s", e.sopsPath)
	}

	// Use SOPS native rotate command for atomic key rotation
	args := []string{"--rotate", "--age", strings.Join(recipients, ","), file}
	cmd := exec.Command(e.sopsPath, args...) //nolint:gosec // sopsPath validated by isValidSOPSPath()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sops rotate failed for %s: %s", file, string(output))
	}

	return nil
}

// rollback restores files from backup
func (e *Executor) rollback(actions []Action, backupDir string) error {
	for i, action := range actions {
		if action.Type == ActionSkip {
			continue
		}

		backupPath := filepath.Join(backupDir, fmt.Sprintf("%d-%s", i, filepath.Base(action.File)))
		if _, err := os.Stat(backupPath); err == nil {
			if err := e.copyFile(backupPath, action.File); err != nil {
				return fmt.Errorf("failed to restore %s: %w", action.File, err)
			}
			fmt.Printf("↺ Restored %s\n", action.File)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func (e *Executor) copyFile(src, dst string) error {
	if err := validateFilePath(src); err != nil {
		return err
	}
	if err := validateFilePath(dst); err != nil {
		return err
	}

	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	data, err := os.ReadFile(cleanSrc)
	if err != nil {
		return err
	}
	return os.WriteFile(cleanDst, data, 0o600)
}
