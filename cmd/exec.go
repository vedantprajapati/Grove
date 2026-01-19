package cmd

import (
	"fmt"
	"grove/internal/manager"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [feature] -- [command] [args...]",
	Short: "Execute a command across all repositories in a feature",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}

		featureName := args[0]

		// Everything after -- is the command
		// If user didn't use --, cobra still passes args
		// Let's assume the first arg is feature, the rest is command
		command := args[1]
		commandArgs := args[2:]

		fmt.Printf("Running command in feature '%s'...\n", featureName)
		return mgr.ExecFeature(featureName, command, commandArgs)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
