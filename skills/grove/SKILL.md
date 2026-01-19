---
name: Grove
description: Manage Git worktrees across multiple repositories simultaneously using the `gr` CLI tool.
---

# Grove Skill

Grove (`gr`) is a CLI tool designed to orchestrate parallel feature development across microservices or multi-repo architectures using Git Worktrees.

## Core Concepts

- **Set**: A collection of related repositories (e.g., frontend, backend).
- **Feature**: A unified workspace for a specific feature, consisting of git worktrees for every repository in a set.
- **Cache**: Bare repositories are cached locally to speed up worktree creation.

## Usage Instructions for Agents

When using Grove, follow these common workflows:

### 1. Defining a Repository Set
If the user wants to group repositories together:
```bash
gr add set <set-name> <repo-url-1> <repo-url-2> ...
```

### 2. Starting a New Feature
To create a parallel workspace for a feature across all repos in a set:
```bash
gr add feature <set-name> <feature-name>
```
*Tip: This will create a directory structure like `root_dir/set-name/feature-name` containing worktrees for each repo.*

### 3. Listing Status
To see defined sets and active features:
```bash
gr list
```

### 4. Syncing Repositories
To pull updates for all repositories in a feature (Parallel):
```bash
gr sync <feature-name>
```

### 5. Executing Parallel Commands
To run a command across all repositories in a feature:
```bash
gr exec <feature-name> -- <command>
```

### 6. Checking Status
To see the internal state (dirty/clean, sync status) of all repos in a feature:
```bash
gr status <feature-name>
```

### 7. Switching to a Feature
To get the path to a feature workspace:
```bash
gr switch <feature-name>
```

### 8. Removing Features or Sets
To clean up:
```bash
gr remove feature <feature-name>
gr remove set <set-name>
```

## Best Practices
- Always check `gr list` to understand the current configuration before making changes.
- Ensure the user's SSH keys are configured if using private repositories.
- When creating features, use descriptive names (e.g., `fix-auth-bug` instead of `feature1`).
