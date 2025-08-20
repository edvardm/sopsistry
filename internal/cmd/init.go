package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SOPS team configuration",
	Long: `Create a new sopsistry.yaml manifest and set up the project for team-based SOPS management.
This command will:
- Create a basic sopsistry.yaml configuration
- Detect existing SOPS files in the repository
- Set up .secrets directory for age keys
- Generate initial age key pair for local development

Use --force to overwrite existing configuration files. The .secrets directory and 
any existing age keys will be preserved.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path") //nolint:errcheck // Flag is defined, error impossible
		force, _ := cmd.Flags().GetBool("force")          //nolint:errcheck // Flag is defined, error impossible

		service := core.NewSopsManager(sopsPath)
		return service.Init(force)
	},
}

func init() {
	initCmd.Flags().BoolP("force", "f", false, "overwrite existing files (preserves .secrets directory)")
	rootCmd.AddCommand(initCmd)
}
