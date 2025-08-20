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

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply planned changes to SOPS files",
	Long: `Execute the planned changes atomically. This command will:
- Verify git working tree is clean (unless --force is used)
- Apply all changes in a single transaction
- Rollback on first failure to maintain consistency`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path")                   //nolint:errcheck // Flag is defined, error impossible
		requireCleanGit, _ := cmd.Flags().GetBool("require-clean-git")      //nolint:errcheck // Flag is defined, error impossible
		noRequireCleanGit, _ := cmd.Flags().GetBool("no-require-clean-git") //nolint:errcheck // Flag is defined, error impossible
		force, _ := cmd.Flags().GetBool("force")                            //nolint:errcheck // Flag is defined, error impossible
		yes, _ := cmd.Flags().GetBool("yes")                                //nolint:errcheck // Flag is defined, error impossible

		gitRequirement := determineGitRequirement(requireCleanGit, noRequireCleanGit, force)

		service := core.NewSopsManager(sopsPath)
		return service.Apply(gitRequirement.requiresCleanGit(), yes)
	},
}

func init() {
	applyCmd.Flags().Bool("force", false, "skip git clean check")
	rootCmd.AddCommand(applyCmd)
}
