package cmd

import (
	"fmt"

	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var addMemberCmd = &cobra.Command{
	Use:     "add-member <id>",
	Aliases: []string{"add"},
	Short:   "Add a team member",
	Long: `Add a new team member to the default scope.
This command updates the team configuration but does not immediately
re-encrypt files. Use 'st plan' and 'st apply' to see and execute changes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		memberID := args[0]
		ageKey, _ := cmd.Flags().GetString("key") //nolint:errcheck // Flag is defined, error impossible

		if ageKey == "" {
			return fmt.Errorf("--key flag is required")
		}

		sopsPath, _ := cmd.Flags().GetString("sops-path") //nolint:errcheck // Flag is defined, error impossible
		service := core.NewSopsManager(sopsPath)
		return service.AddMember(memberID, ageKey)
	},
}

var removeMemberCmd = &cobra.Command{
	Use:     "remove-member <id>",
	Aliases: []string{"rm"},
	Short:   "Remove a team member",
	Long: `Remove a team member from all scopes.
This command updates the team configuration but does not immediately
re-encrypt files. Use 'st plan' and 'st apply' to see and execute changes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		memberID := args[0]

		sopsPath, _ := cmd.Flags().GetString("sops-path") //nolint:errcheck // Flag is defined, error impossible
		service := core.NewSopsManager(sopsPath)
		return service.RemoveMember(memberID)
	},
}

func init() {
	addMemberCmd.Flags().String("key", "", "age public key for the member (required)")
	_ = addMemberCmd.MarkFlagRequired("key") // Error is not critical for flag setup

	rootCmd.AddCommand(addMemberCmd)
	rootCmd.AddCommand(removeMemberCmd)
}
