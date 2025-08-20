// Package cmd provides the CLI commands for the sopsistry tool
package cmd

import (
	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

// GitRequirement represents different git cleanliness requirements
type GitRequirement int

const (
	GitRequired GitRequirement = iota // Default: require clean git
	GitForced                         // Force flag overrides git check
	GitSkipped                        // Explicitly skip git check
)

// requiresCleanGit returns true if git working tree must be clean
func (g GitRequirement) requiresCleanGit() bool {
	return g == GitRequired
}

// determineGitRequirement resolves conflicting git flags into single requirement
func determineGitRequirement(requireClean, noRequireClean, force bool) GitRequirement {
	switch {
	case force:
		return GitForced
	case noRequireClean:
		return GitSkipped
	case requireClean:
		return GitRequired
	default:
		return GitRequired // Default behavior
	}
}

var applySafeCmd *SafeCommand

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply planned changes to SOPS files",
	Long: `Execute the planned changes atomically. This command will:
- Verify git working tree is clean (unless --force is used)
- Apply all changes in a single transaction
- Rollback on first failure to maintain consistency`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Guaranteed safe flag access - no errors possible
		sopsPath := applySafeCmd.GetStringFlag("sops-path")
		requireCleanGit := applySafeCmd.GetBoolFlag("require-clean-git")
		noRequireCleanGit := applySafeCmd.GetBoolFlag("no-require-clean-git")
		force := applySafeCmd.GetBoolFlag("force")
		yes := applySafeCmd.GetBoolFlag("yes")

		gitRequirement := determineGitRequirement(requireCleanGit, noRequireCleanGit, force)

		service := core.NewSopsManager(sopsPath)
		return service.Apply(gitRequirement.requiresCleanGit(), yes)
	},
}

func init() {
	// Create SafeCommand and register local flags (persistent flags from root are handled automatically)
	applySafeCmd = NewSafeCommand(applyCmd)
	applySafeCmd.RegisterBoolFlag("no-require-clean-git", false, "skip git clean check")
	applySafeCmd.RegisterBoolFlag("force", false, "skip git clean check")

	rootCmd.AddCommand(applyCmd)
}
