package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "sistry",
	Short: "SOPS Team management CLI",
	Long: `A CLI tool for managing SOPS team configurations.
Provides standardized workflows for team member onboarding/offboarding,
key rotation, and encrypted file management.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

// Execute runs the root command and returns any error
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().Bool("json", false, "output in JSON format")
	rootCmd.PersistentFlags().String("sops-path", "sops", "path to sops binary")
	rootCmd.PersistentFlags().Bool("require-clean-git", true, "require clean git working tree")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "automatically confirm prompts")

	rootCmd.SetOut(os.Stderr)
	rootCmd.SetErr(os.Stderr)
}
