package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var rotateKeyCmd = &cobra.Command{
	Use:     "rotate-key",
	Aliases: []string{"rot"},
	Short:   "Rotate the current user's age key",
	Long: `Generate a new age key pair and re-encrypt all files with the new key.
This command will:
- Check if key rotation is needed based on max_key_age_days setting
- Generate a new age key pair
- Update the manifest with the new public key and timestamp
- Re-encrypt all affected files with the new key
- Backup and restore on failure

Use --force to skip age validation and rotate immediately.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path") //nolint:errcheck // Flag is defined, error impossible
		force, _ := cmd.Flags().GetBool("force")          //nolint:errcheck // Flag is defined, error impossible

		service := core.NewSopsManager(sopsPath)
		return service.RotateKey(force)
	},
}

func init() {
	rotateKeyCmd.Flags().BoolP("force", "f", false, "force rotation even if key is not expired")
	rootCmd.AddCommand(rotateKeyCmd)
}
