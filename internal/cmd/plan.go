package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show planned changes without applying them",
	Long: `Compute and display what changes would be made to SOPS files
based on the current team configuration. This is a dry-run that shows:
- Which files will be re-encrypted
- What recipients will be added or removed
- Any validation errors or warnings`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path") //nolint:errcheck // Flag is defined, error impossible
		noColor, _ := cmd.Flags().GetBool("no-color")     //nolint:errcheck // Flag is defined, error impossible

		service := core.NewSopsManager(sopsPath)
		return service.Plan(noColor)
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
