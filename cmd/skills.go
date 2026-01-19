package cmd

import (
	"fmt"
	"github.com/vedantprajapati/Grove/internal/manager"

	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage skills for sets",
}

var skillsListCmd = &cobra.Command{
	Use:   "list [set-name]",
	Short: "List configuration for skills",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}

		setName := args[0]
		set, ok := mgr.Config.Sets[setName]
		if !ok {
			return fmt.Errorf("set '%s' not found", setName)
		}

		fmt.Printf("Skills directory for '%s': %s\n", setName, set.SkillsDir)
		return nil
	},
}

var skillsSetDirCmd = &cobra.Command{
	Use:   "set-dir [set-name] [path]",
	Short: "Set the source directory for skills",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}

		setName := args[0]
		path := args[1]

		// Validate set exists
		set, ok := mgr.Config.Sets[setName]
		if !ok {
			return fmt.Errorf("set '%s' not found", setName)
		}

		// Update path
		set.SkillsDir = path
		mgr.Config.Sets[setName] = set

		if err := mgr.SaveConfig(); err != nil {
			return err
		}

		fmt.Printf("Skills directory for '%s' updated to %s\n", setName, path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(skillsCmd)
	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsSetDirCmd)
}
