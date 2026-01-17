# Contributing to Grove

We welcome contributions! Here's how to get started.

## Development Setup

1.  **Prerequisites**:
    - Go 1.21 or higher
    - Git

2.  **Clone the repo**:
    ```bash
    git clone git@github.com:vedantprajapati/Grove.git
    cd Grove
    ```

3.  **Run Tests**:
    We have an integration test suite that verifies the core workflow.
    ```bash
    go test -v ./tests/...
    ```

4.  **Pre-commit Hook**:
    This repo comes with a `pre-commit` hook that runs tests automatically.
    To enable it manually (if not picked up):
    ```bash
    cp .git/hooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    ```

## Project Structure

- `cmd/`: CLI commands (Cobra definitions).
- `internal/`:
    - `config/`: Configuration file management (`.groverc`).
    - `git/`: Git command wrappers and worktree logic.
    - `manager/`: Core business logic (Sets, Features).
- `tests/`: Integration tests.

## process

1.  Fork the repo.
2.  Create a feature branch.
3.  Ensure tests pass (`go test ./tests/...`).
4.  Submit a Pull Request.

## Coding Standards

- Use `go fmt` before committing.
- Keep dependencies minimal.
- Ensure cross-platform compatibility (use `filepath.Join` instead of manual slashes).
