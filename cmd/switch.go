package cmd

import (
	"fmt"
	"github.com/vedantprajapati/Grove/internal/manager"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [feature]",
	Short: "Switch to a feature workspace",
	Long: `Output the path of the feature workspace. 
Use with shell alias: cd $(gr switch my-feature)`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}
		
		feat, ok := mgr.Config.Features[args[0]]
		if !ok {
			return fmt.Errorf("feature '%s' not found", args[0])
		}
		
		fmt.Println(feat.Path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
