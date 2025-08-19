package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var sopsCmd = &cobra.Command{
	Use:   "sops-cmd [sops-args...]",
	Short: "Show the SOPS command with team environment variables",
	Long: `Display the SOPS command with proper environment variables set for the current team.
This is useful for partial encryption or when you need to use SOPS directly.

Examples:
  sistry sops-cmd -e secrets.yaml              # Show encrypt command
  sistry sops-cmd -e --encrypted-regex '^(password|key)' .env  # Partial encryption
  sistry sops-cmd -d secrets.yaml              # Show decrypt command

You can copy and run the displayed command directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path")
		execute, _ := cmd.Flags().GetBool("exec")

		service := core.NewSopsManager(sopsPath)
		return service.ShowSOPSCommand(args, execute)
	},
}

func init() {
	sopsCmd.Flags().Bool("exec", false, "execute the SOPS command instead of just showing it")
	rootCmd.AddCommand(sopsCmd)
}
