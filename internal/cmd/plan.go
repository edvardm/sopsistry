package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var planSafeCmd *SafeCommand

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show planned changes without applying them",
	Long: `Compute and display what changes would be made to SOPS files
based on the current team configuration. This is a dry-run that shows:
- Which files will be re-encrypted
- What recipients will be added or removed
- Any validation errors or warnings`,
	RunE: func(_ *cobra.Command, _ []string) error {
		sopsPath := planSafeCmd.GetStringFlag("sops-path")
		noColor := planSafeCmd.GetBoolFlag("no-color")

		service := core.NewSopsManager(sopsPath)
		return service.Plan(noColor)
	},
}

func init() {
	planSafeCmd = NewSafeCommand(planCmd)
	// Uses persistent flags from root: sops-path, no-color

	rootCmd.AddCommand(planCmd)
}
