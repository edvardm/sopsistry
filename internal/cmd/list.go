package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var listSafeCmd *SafeCommand

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List team members and managed files",
	Long: `Display current team configuration including:
- Team members and their age keys
- Encrypted files under management
- Current scope assignments`,
	RunE: func(_ *cobra.Command, _ []string) error {
		sopsPath := listSafeCmd.GetStringFlag("sops-path")
		jsonOutput := listSafeCmd.GetBoolFlag("json")

		service := core.NewSopsManager(sopsPath)
		return service.List(jsonOutput)
	},
}

func init() {
	listSafeCmd = NewSafeCommand(listCmd)
	// Uses persistent flags from root: sops-path, json

	rootCmd.AddCommand(listCmd)
}
