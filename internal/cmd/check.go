package cmd

import (
	"fmt"

	"github.com/edvardm/sopsistry/internal/core"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check SOPS configuration and key expiry status",
	Long: `Check for existing SOPS configuration, team compatibility, and key expiry status.
This command helps identify potential conflicts between existing .sops.yaml
files and team-managed encryption settings, and warns about expired or expiring keys.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		sopsPath, _ := cmd.Flags().GetString("sops-path") //nolint:errcheck // Flag is defined, error impossible
		verbose, _ := cmd.Flags().GetBool("verbose")      //nolint:errcheck // Flag is defined, error impossible

		// Check SOPS configuration compatibility
		detector := core.NewSOPSDetector()
		sopsInfo, err := detector.DetectSOPSConfig()
		if err != nil {
			return fmt.Errorf("failed to detect SOPS configuration: %w", err)
		}

		if !sopsInfo.Exists {
			fmt.Printf("✅ No existing SOPS configuration detected\n")
			fmt.Printf("   Team management will work without conflicts\n")
		} else {
			fmt.Printf("📋 SOPS Configuration Analysis:\n\n")
			fmt.Printf("  Found: %s\n", sopsInfo.ConfigPath)

			if sopsInfo.HasCreationRules {
				fmt.Printf("⚙️  Has creation rules\n")
			}
			if sopsInfo.HasAgeKeys {
				fmt.Printf("🔑 Contains age keys\n")
			}
			if sopsInfo.HasKMSKeys {
				fmt.Printf("☁️  Contains KMS keys\n")
			}
			if sopsInfo.HasPGPKeys {
				fmt.Printf("🔐 Contains PGP keys\n")
			}

			fmt.Printf("\n")

			if sopsInfo.ShouldWarn() {
				fmt.Printf("%s\n", sopsInfo.GetWarningMessage())
			} else {
				fmt.Printf("✅ Configuration looks compatible with team management\n")
			}

			fmt.Printf("\n%s\n", sopsInfo.GetCoexistenceAdvice())
		}

		// Check key expiry status
		fmt.Printf("\n🔑 Key Expiry Status:\n")
		service := core.NewSopsManager(sopsPath)
		if err := service.CheckKeyExpiry(verbose); err != nil {
			// Don't fail the whole command if key checking fails
			fmt.Printf("❌ Failed to check key expiry: %v\n", err)
		}

		return nil
	},
}

func init() {
	checkCmd.Flags().BoolP("verbose", "v", false, "show detailed key mapping information")
	rootCmd.AddCommand(checkCmd)
}
