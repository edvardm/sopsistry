package cmd

import (
	"fmt"

	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var addMemberSafeCmd *SafeCommand
var removeMemberSafeCmd *SafeCommand

var addMemberCmd = &cobra.Command{
	Use:     "add-member <id>",
	Aliases: []string{"add"},
	Short:   "Add a team member",
	Long: `Add a new team member to the default scope.
This command updates the team configuration but does not immediately
re-encrypt files. Use 'st plan' and 'st apply' to see and execute changes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		memberID := args[0]
		ageKey := addMemberSafeCmd.GetStringFlag("key")

		if ageKey == "" {
			return fmt.Errorf("--key flag is required")
		}

		sopsPath := addMemberSafeCmd.GetStringFlag("sops-path")
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
	RunE: func(_ *cobra.Command, args []string) error {
		memberID := args[0]

		sopsPath := removeMemberSafeCmd.GetStringFlag("sops-path")
		service := core.NewSopsManager(sopsPath)
		return service.RemoveMember(memberID)
	},
}

func init() {
	addMemberSafeCmd = NewSafeCommand(addMemberCmd)
	addMemberSafeCmd.RegisterStringFlag("key", "", "age public key for the member (required)")
	_ = addMemberCmd.MarkFlagRequired("key") // Error is not critical for flag setup

	removeMemberSafeCmd = NewSafeCommand(removeMemberCmd)

	rootCmd.AddCommand(addMemberCmd)
	rootCmd.AddCommand(removeMemberCmd)
}
