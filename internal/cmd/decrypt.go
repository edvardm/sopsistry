package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var decryptCmd = &cobra.Command{
	Use:     "decrypt <file>",
	Aliases: []string{"dec"},
	Short:   "Decrypt a SOPS-encrypted file",
	Long: `Decrypt a SOPS-encrypted file using your local age key.
By default outputs to stdout. Use --in-place to decrypt the file directly.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		sopsPath, _ := cmd.Flags().GetString("sops-path")
		inPlace, _ := cmd.Flags().GetBool("in-place")

		service := core.NewSopsManager(sopsPath)
		return service.DecryptFile(filePath, inPlace)
	},
}

func init() {
	decryptCmd.Flags().BoolP("in-place", "i", false, "decrypt file in-place (default: output to stdout)")
	rootCmd.AddCommand(decryptCmd)
}
