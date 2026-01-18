#!/bin/bash
# Grove shell integration
# Add this to your ~/.bashrc or ~/.zshrc:
# source <(curl -s https://raw.githubusercontent.com/vedantprajapati/Grove/main/gr.sh)

gr() {
    if [ "$1" = "switch" ]; then
        # Get the path from the gr binary
        local path=$(command gr switch "$2")
        if [ $? -eq 0 ]; then
            cd "$path" || return 1
        else
            return 1
        fi
    else
        # For all other commands, just run gr normally
        command gr "$@"
    fi
}
