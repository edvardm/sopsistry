package core

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// checkGitClean verifies the git working tree is clean
func (s *SopsManager) checkGitClean() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not in a git repository")
	}

	cmd = exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(string(output)) != EmptyString {
		return fmt.Errorf("git working tree is not clean. Commit or stash changes first, or use --force")
	}

	return nil
}

// ensureGitignore adds .secrets to .gitignore if not already present
func (s *SopsManager) ensureGitignore() error {
	gitignorePath := ".gitignore"

	lines, secretsIgnored := s.readGitignoreLines(gitignorePath)

	if !secretsIgnored {
		return s.addSecretsToGitignore(gitignorePath, lines)
	}

	return nil
}

func (s *SopsManager) readGitignoreLines(gitignorePath string) ([]string, bool) {
	var lines []string
	secretsIgnored := false

	file, err := os.Open(gitignorePath) //nolint:gosec // False positive: gitignorePath is hardcoded as ".gitignore"
	if err != nil {
		return lines, secretsIgnored
	}
	defer func() { _ = file.Close() }() //nolint:errcheck // File cleanup, error not critical

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lines = append(lines, scanner.Text())

		if s.isSecretsIgnorePattern(line) {
			secretsIgnored = true
		}
	}

	return lines, secretsIgnored
}

func (s *SopsManager) isSecretsIgnorePattern(line string) bool {
	patterns := []string{".secrets", ".secrets/", "/.secrets"}
	return slices.Contains(patterns, line)
}

func (s *SopsManager) addSecretsToGitignore(gitignorePath string, lines []string) error {
	updatedLines := s.appendSecretsEntry(lines)
	content := strings.Join(updatedLines, "\n") + "\n"

	if err := os.WriteFile(gitignorePath, []byte(content), GitignoreFileMode); err != nil { //nolint:gosec // .gitignore uses standard permissions
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	fmt.Printf("Added .secrets to .gitignore\n")
	return nil
}

func (s *SopsManager) appendSecretsEntry(lines []string) []string {
	if len(lines) > 0 && lines[len(lines)-1] != EmptyString {
		lines = append(lines, EmptyString)
	}
	lines = append(lines, "# SOPS team private keys", ".secrets")
	return lines
}
