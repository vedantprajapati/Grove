package cmd

import (
	"fmt"
	"grove/internal/manager"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [feature]",
	Short: "Check the status of all repositories in a feature",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}

		featureName := args[0]
		statuses, err := mgr.GetFeatureStatus(featureName)
		if err != nil {
			return err
		}

		fmt.Println(headerStyle.Render(fmt.Sprintf("📊 Feature Status: %s", featureName)))

		t := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#30363d")).
			Padding(1)

		var rows []string
		// Header
		rows = append(rows, lipgloss.NewStyle().Bold(true).Render("  REPO           BRANCH       STATUS       SYNC"))

		for _, s := range statuses {
			cleanStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#2ea043")).Render("CLEAN")
			if s.IsDirty {
				cleanStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff7b72")).Render("DIRTY")
			}

			// Format ABC (Ahead/Behind Count) e.g. "0	0"
			syncStatus := s.ABC
			parts := strings.Split(s.ABC, "\t")
			if len(parts) == 2 {
				ahead := parts[0]
				behind := parts[1]
				syncStatus = fmt.Sprintf("↑%s ↓%s", ahead, behind)
			}

			rows = append(rows, fmt.Sprintf("  %-14s %-12s %-12s %s",
				setNameStyle.Render(s.Name),
				featureNameStyle.Render(s.Branch),
				cleanStatus,
				syncStatus))
		}

		fmt.Println(t.Render(strings.Join(rows, "\n")))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
