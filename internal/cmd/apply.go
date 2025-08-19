// Package cmd provides the CLI commands for the sopsistry tool
package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply planned changes to SOPS files",
	Long: `Execute the planned changes atomically. This command will:
- Verify git working tree is clean (unless --force is used)
- Apply all changes in a single transaction
- Rollback on first failure to maintain consistency`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path")
		requireCleanGit, _ := cmd.Flags().GetBool("require-clean-git")
		noRequireCleanGit, _ := cmd.Flags().GetBool("no-require-clean-git")
		force, _ := cmd.Flags().GetBool("force")
		yes, _ := cmd.Flags().GetBool("yes")

		actualRequireClean := requireCleanGit && !noRequireCleanGit && !force

		service := core.NewSopsManager(sopsPath)
		return service.Apply(actualRequireClean, yes)
	},
}

func init() {
	applyCmd.Flags().Bool("force", false, "skip git clean check")
	rootCmd.AddCommand(applyCmd)
}
