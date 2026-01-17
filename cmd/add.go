package cmd

import (
	"grove/internal/manager"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [set-name] [feature-name]",
	Short: "Add a set or a feature",
	Long: `Add a defined set of repositories or create a new feature worktree.
	
Examples:
  gr add set my-set git@github.com:usr/repo1.git git@github.com:usr/repo2.git
  gr add my-set new-login-flow
`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 2 {
			// Assume "gr add [set] [feature]"
			mgr, err := manager.NewManager()
			if err != nil {
				return err
			}
			return mgr.CreateFeature(args[0], args[1])
		}
		return cmd.Help()
	},
}

var addSetCmd = &cobra.Command{
	Use:   "set [name] [repo-urls...]",
	Short: "Define a new set of repositories",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		repos := args[1:]

		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}
		return mgr.AddSet(name, repos)
	},
}

var addFeatureCmd = &cobra.Command{
	Use:   "feature [set-name] [feature-name]",
	Short: "Create a new feature worktree",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}
		return mgr.CreateFeature(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addSetCmd)
	addCmd.AddCommand(addFeatureCmd)
}
