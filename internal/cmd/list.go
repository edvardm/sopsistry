package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List team members and managed files",
	Long: `Display current team configuration including:
- Team members and their age keys
- Encrypted files under management
- Current scope assignments`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path") //nolint:errcheck // Flag is defined, error impossible
		jsonOutput, _ := cmd.Flags().GetBool("json")      //nolint:errcheck // Flag is defined, error impossible

		service := core.NewSopsManager(sopsPath)
		return service.List(jsonOutput)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
