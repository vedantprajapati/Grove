package cmd

import (
	"fmt"

	"github.com/vedantprajapati/Grove/internal/manager"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [feature]",
	Short: "Sync all repositories in a feature with remote",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}

		featureName := args[0]
		feat, ok := mgr.Config.Features[featureName]
		if !ok {
			return fmt.Errorf("feature '%s' not found", featureName)
		}

		if _, ok := mgr.Config.Sets[feat.Set]; !ok {
			return fmt.Errorf("set '%s' not found for feature", feat.Set)
		}

		fmt.Printf("Syncing feature '%s'...\n", featureName)
		if err := mgr.SyncFeature(featureName); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
