package cmd

import (
	"fmt"
	"github.com/vedantprajapati/Grove/internal/git"
	"github.com/vedantprajapati/Grove/internal/manager"
	"path/filepath"

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
		
		set, ok := mgr.Config.Sets[feat.Set]
		if !ok {
			return fmt.Errorf("set '%s' not found for feature", feat.Set)
		}

		fmt.Printf("Syncing feature '%s' (Set: %s)...\n", featureName, feat.Set)
		
		for _, repoURL := range set.Repos {
            repoName := git.GetRepoNameFromURL(repoURL)
            repoPath := filepath.Join(feat.Path, repoName)
            if err := git.SyncRepo(repoPath); err != nil {
                fmt.Printf("  Error syncing %s: %v\n", repoName, err)
            } else {
                fmt.Printf("  %s synced.\n", repoName)
            }
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
