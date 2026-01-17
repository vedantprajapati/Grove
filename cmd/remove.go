package cmd

import (
	"grove/internal/manager"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [feature-name]",
	Short: "Remove a feature or a set",
	Long: `Remove a feature workspace or a set definition.
	
Examples:
  gr remove my-feature
  gr remove set my-set
`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			// "gr remove feature-name"
			mgr, err := manager.NewManager()
			if err != nil {
				return err
			}
			return mgr.RemoveFeature(args[0])
		}
		return cmd.Help()
	},
}

var removeSetCmd = &cobra.Command{
	Use:   "set [name]",
	Short: "Remove a set definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}
		return mgr.RemoveSet(args[0])
	},
}

var removeFeatureCmd = &cobra.Command{
	Use:   "feature [name]",
	Short: "Remove a feature workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}
		return mgr.RemoveFeature(args[0])
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.AddCommand(removeSetCmd)
	removeCmd.AddCommand(removeFeatureCmd)
}
