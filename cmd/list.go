package cmd

import (
	"fmt"
	"github.com/vedantprajapati/Grove/internal/manager"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2ea043")).
			MarginTop(1).
			MarginBottom(1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#30363d")).
			Padding(0, 1).
			MarginBottom(1)

	setNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#58a6ff"))

	featureNameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ff7b72"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8b949e"))

	linkStyle = dimStyle.
			Underline(true)
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sets and active features",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}

		fmt.Println(headerStyle.Render("🌳 Grove Status"))
		fmt.Printf("%s %s\n\n", dimStyle.Render("Root directory:"), mgr.Config.RootDir)

		// Sets Section
		fmt.Println(lipgloss.NewStyle().Bold(true).Underline(true).Render("Sets"))
		if len(mgr.Config.Sets) == 0 {
			fmt.Println(dimStyle.Render("  No sets defined. Use 'gr add set' to create one."))
		} else {
			var setsOutput []string
			for name, set := range mgr.Config.Sets {
				content := fmt.Sprintf("%s\n%s %d repos\n%s %s",
					setNameStyle.Render(name),
					dimStyle.Render("Repositories:"), len(set.Repos),
					dimStyle.Render("Skills:"), set.SkillsDir)
				setsOutput = append(setsOutput, cardStyle.Render(content))
			}
			fmt.Println(lipgloss.JoinVertical(lipgloss.Left, setsOutput...))
		}

		// Features Section
		fmt.Println(lipgloss.NewStyle().Bold(true).Underline(true).Render("\nActive Features"))
		if len(mgr.Config.Features) == 0 {
			fmt.Println(dimStyle.Render("  No active features. Use 'gr add feature' to start work."))
		} else {
			var featsOutput []string
			for name, feat := range mgr.Config.Features {
				path := feat.Path
				if runtime.GOOS == "windows" {
					path = filepath.ToSlash(path)
					if !filepath.IsAbs(path) {
						abs, _ := filepath.Abs(path)
						path = filepath.ToSlash(abs)
					}
				}

				content := fmt.Sprintf("%s (Set: %s)\n%s %s",
					featureNameStyle.Render(name),
					dimStyle.Render(feat.Set),
					dimStyle.Render("Path:"), linkStyle.Render("file:///"+path))
				featsOutput = append(featsOutput, cardStyle.Render(content))
			}
			fmt.Println(lipgloss.JoinVertical(lipgloss.Left, featsOutput...))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
