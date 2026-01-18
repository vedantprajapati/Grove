package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate shell integration code",
	Long: `Generate shell integration code for gr.

Add this to your ~/.bashrc or ~/.zshrc:
  eval "$(gr init)"

This enables the 'gr switch' command to change directories in your current shell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellIntegration := `# Grove shell integration
gr() {
    if [ "$1" = "switch" ]; then
        # Get the path from the gr binary
        local path=$(command gr switch "$2" 2>&1)
        local exit_code=$?
        if [ $exit_code -eq 0 ] && [ -n "$path" ]; then
            cd "$path" || return 1
        else
            echo "$path" >&2
            return $exit_code
        fi
    else
        # For all other commands, just run gr normally
        command gr "$@"
    fi
}`
		fmt.Println(shellIntegration)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
