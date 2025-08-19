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
func (s *TeamService) checkGitClean() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not in a git repository")
	}

	cmd = exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("git working tree is not clean. Commit or stash changes first, or use --force")
	}

	return nil
}

// ensureGitignore adds .secrets to .gitignore if not already present
func (s *TeamService) ensureGitignore() error {
	gitignorePath := ".gitignore"

	lines, secretsIgnored := s.readGitignoreLines(gitignorePath)

	if !secretsIgnored {
		return s.addSecretsToGitignore(gitignorePath, lines)
	}

	return nil
}

func (s *TeamService) readGitignoreLines(gitignorePath string) ([]string, bool) {
	var lines []string
	secretsIgnored := false

	file, err := os.Open(gitignorePath) //nolint:gosec // False positive: gitignorePath is hardcoded as ".gitignore"
	if err != nil {
		return lines, secretsIgnored
	}
	defer func() { _ = file.Close() }()

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

func (s *TeamService) isSecretsIgnorePattern(line string) bool {
	patterns := []string{".secrets", ".secrets/", "/.secrets"}
	return slices.Contains(patterns, line)
}

func (s *TeamService) addSecretsToGitignore(gitignorePath string, lines []string) error {
	updatedLines := s.appendSecretsEntry(lines)
	content := strings.Join(updatedLines, "\n") + "\n"

	if err := os.WriteFile(gitignorePath, []byte(content), 0o644); err != nil { //nolint:gosec // .gitignore uses standard permissions
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	fmt.Printf("Added .secrets to .gitignore\n")
	return nil
}

func (s *TeamService) appendSecretsEntry(lines []string) []string {
	if len(lines) > 0 && lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	lines = append(lines, "# SOPS team private keys")
	lines = append(lines, ".secrets")
	return lines
}
