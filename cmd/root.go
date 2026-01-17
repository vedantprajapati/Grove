package cmd

import (
	"fmt"
	"github.com/vedantprajapati/Grove/internal/manager"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gr",
	Short: "Grove: Manage git worktrees across multiple repositories",
	Long:  `Grove helps you manage development features that span multiple repositories by orchestrating git worktrees.`,
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		// Shortcut: gr [feature] [tool] ...
		featureName := args[0]

		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}

		feat, ok := mgr.Config.Features[featureName]
		if !ok {
			// Not a feature, and not a known command (Cobra handles valid commands)
			// Return error or help
			return fmt.Errorf("unknown command or feature: %s", featureName)
		}

		// It is a feature.
		// If additional args, treat as tool + args
		if len(args) > 1 {
			tool := args[1]
			toolArgs := args[2:]

			fmt.Printf("Launching %s in %s...\n", tool, feat.Path)

			// Execute command
			// On Windows, might need shell invocation?
			// Simple exec first.
			c := exec.Command(tool, toolArgs...)
			c.Dir = feat.Path
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			return c.Run()
		} else {
			// Just gr [feature] -> maybe just print path? or error?
			// "gr switch" prints path.
			// Let's print path for consistency if no tool specified.
			fmt.Println(feat.Path)
			return nil
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
