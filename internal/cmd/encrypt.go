package cmd

import (
	"fmt"

	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var encryptSafeCmd *SafeCommand

var encryptCmd = &cobra.Command{
	Use:     "encrypt <file>",
	Aliases: []string{"enc"},
	Short:   "Encrypt a file with SOPS using team configuration",
	Long: `Encrypt a file using the current team configuration.
The file will be encrypted in-place using age keys from the team manifest.

Examples:
  st encrypt .env                            # Encrypt entire file
  st encrypt --regex '^(password|key)' .env # Encrypt only matching fields (partial)
  st encrypt --iregex '^(password|key)' .env # Case-insensitive partial encryption
  st encrypt --regex '.*secret.*' config.yaml # Encrypt fields containing 'secret'`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		filePath := args[0]

		sopsPath := encryptSafeCmd.GetStringFlag("sops-path")
		inPlace := encryptSafeCmd.GetBoolFlag("in-place")
		regex := encryptSafeCmd.GetStringFlag("regex")
		iregex := encryptSafeCmd.GetStringFlag("iregex")

		// Check that only one of regex or iregex is provided
		if regex != "" && iregex != "" {
			return fmt.Errorf("cannot use both --regex and --iregex at the same time")
		}

		// Convert iregex to case-insensitive regex
		if iregex != "" {
			regex = "(?i)" + iregex
		}

		service := core.NewSopsManager(sopsPath)
		return service.EncryptFile(filePath, inPlace, regex)
	},
}

func init() {
	encryptSafeCmd = NewSafeCommand(encryptCmd)
	encryptSafeCmd.RegisterBoolFlag("in-place", true, "encrypt file in-place")
	encryptSafeCmd.RegisterStringFlag("regex", "", "encrypt only fields matching this regex (partial encryption)")
	encryptSafeCmd.RegisterStringFlag("iregex", "", "encrypt only fields matching this case-insensitive regex (partial encryption)")

	rootCmd.AddCommand(encryptCmd)
}
