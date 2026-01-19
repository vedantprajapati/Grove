# Grove 🌳

**Manage Git worktrees across multiple repositories simultaneously.**

Grove (`gr`) is a CLI tool designed to orchestrate parallel feature development across microservices or multi-repo architectures using Git Worktrees.

## Features

- **Sets**: Define groups of related repositories (e.g., frontend, backend, shared-lib).
- **Features**: Create a unified workspace for a feature, automatically creating git worktrees for every repo in the set.
- **Skills**: Share AI context/prompts (skills) across your sets.
- **Switching**: Quickly jump between feature workspaces.

## Installation

### Prerequisites
- **Go 1.21+** installed ([Download](https://go.dev/dl/))
- **Git** installed and in PATH.

### Install via Go
```bash
go install github.com/vedantprajapati/Grove@latest
```
*(Note: Replace URL with actual path if private, or clone and install locally)*

**Local Installation:**
```bash
git clone git@github.com:vedantprajapati/Grove.git
cd Grove
go install .
```

Ensure your Go bin directory is in your PATH (`%USERPROFILE%\go\bin` on Windows, `$GOPATH/bin` on Linux/Mac).

## Usage

### 1. Define a Set
A "Set" is a collection of repositories you often work on together.
```bash
gr add set my-stack git@github.com:org/backend.git git@github.com:org/frontend.git
```

### 2. Start a Feature
Create a new feature workspace. This clones the repos (cached) and creates worktrees.
```bash
gr add feature my-stack new-login-flow
```
*Alias:* `gr add my-stack new-login-flow`

### 3. List Workspaces
See all your active features and sets.
```bash
gr list
```

### 4. Switch Context
Get the path to your feature workspace.
```bash
cd $(gr switch new-login-flow)
```
*Windows Powershell:*
```powershell
cd (gr switch new-login-flow)
```

### 5. Sync Changes
Pull updates for all repositories in a feature (Parallel).
```bash
gr sync new-login-flow
```

### 6. Execute Parallel Commands
Run a command across all repositories in a feature workspace simultaneously.
```bash
gr exec new-login-flow -- npm install
```

### 7. Check Feature Status
See a dashboard of the current branch, dirty status, and sync state for all repositories.
```bash
gr status new-login-flow
```

## Configuration

Configuration is stored in `~/.groverc`.
Bare repositories are cached in `~/.grove/cache`.

## License
MIT
