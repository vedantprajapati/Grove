package cmd

import (
	"fmt"
	"grove/internal/manager"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sets and active features",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := manager.NewManager()
		if err != nil {
			return err
		}
		
		fmt.Printf("Root Dir: %s\n", mgr.Config.RootDir)
		
		fmt.Println("\nSets:")
		for name, set := range mgr.Config.Sets {
			fmt.Printf("  - %s (%d repos)\n", name, len(set.Repos))
			fmt.Printf("    Skills: %s\n", set.SkillsDir)
		}
		
		fmt.Println("\nActive Features:")
		for name, feat := range mgr.Config.Features {
			fmt.Printf("  - %s (Set: %s)\n", name, feat.Set)
			
			// Format path as clickable URI
			// Windows: file:///C:/path
			// Unix: file:///path
			path := feat.Path
			if runtime.GOOS == "windows" {
				path = filepath.ToSlash(path)
                // Ensure config path starts with drive letter correctly for URI?
                // Actually, simple / is enough for some terminals, but full URI is safer.
                // Assuming absolute path.
                if !filepath.IsAbs(path) {
                    abs, _ := filepath.Abs(path)
                    path = filepath.ToSlash(abs)
                }
			}
			fmt.Printf("    Path: \033[4mfile:///%s\033[0m\n", path)
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
