package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RunGit executes a git command in the specified directory.
func RunGit(cwd string, args ...string) (string, error) {
	var output []byte
	var err error

	// Retry up to 3 times for lock errors or transient fs issues
	for i := 0; i < 3; i++ {
		cmd := exec.Command("git", args...)
		if cwd != "" {
			cmd.Dir = cwd
		}
		output, err = cmd.CombinedOutput()

		// Success
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}

		// Check for lock file errors to decide if we should retry
		outStr := string(output)
		if strings.Contains(outStr, "index.lock") || strings.Contains(outStr, "lock file") {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// If it's not a lock error, fail immediately
		break
	}

	return "", fmt.Errorf("git command failed: %s\nOutput: %s", strings.Join(args, " "), string(output))
}

// RunCommand executes an arbitrary command in the specified directory.
func RunCommand(cwd string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRepoNameFromURL extracts the repository name from a URL.
// e.g., git@github.com:user/repo.git -> repo
// e.g., C:\Users\user\repo -> repo
func GetRepoNameFromURL(url string) string {
	// Normalize path separators to forward slashes for consistency
	url = filepath.ToSlash(url)
	parts := strings.Split(url, "/")
	// Handle trailing slashes if necessary (unlikely for git URLs but good practice)
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return ""
	}
	name := parts[len(parts)-1]
	return strings.TrimSuffix(name, ".git")
}

// EnsureBareRepo clones a repository as a bare repo into the cache directory if it doesn't exist.
func EnsureBareRepo(url, cacheDir string) (string, error) {
	repoName := GetRepoNameFromURL(url)
	barePath := filepath.Join(cacheDir, repoName)

	if _, err := os.Stat(barePath); os.IsNotExist(err) {
		// Create cache dir if needed
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return "", err
		}

		fmt.Printf("Cloning %s to %s...\n", url, barePath)
		if _, err := RunGit("", "clone", "--bare", url, barePath); err != nil {
			return "", err
		}
	}
	return barePath, nil
}

// CreateWorktree creates a new worktree from a bare repository.
// It matches the functionality: git worktree add [-b branch] path [branch]
// CreateWorktree creates a new worktree from a bare repository.
// It matches the functionality: git worktree add -B branch path
func CreateWorktree(barePath, branchName, targetPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// Normalize paths for git command (Windows friendliness)
	targetPath = filepath.ToSlash(targetPath)
	barePath = filepath.ToSlash(barePath)

	// Use -B to force create/reset branch. This handles both new and existing branches.
	fmt.Printf("  Creating worktree at %s (branch %s)...\n", targetPath, branchName)
	_, err := RunGit(barePath, "worktree", "add", "-B", branchName, targetPath)

	if err != nil {
		// Cleanup target path if failed (git might leave lock files or empty dir)
		os.RemoveAll(targetPath)
		if strings.Contains(err.Error(), "checked out") {
			return fmt.Errorf("branch %s is already checked out in another worktree", branchName)
		}
		return fmt.Errorf("git worktree add failed: %v", err)
	}
	return nil
}

// BranchExists checks if a branch exists in the repository.
func BranchExists(repoPath, branchName string) (bool, error) {
	_, err := RunGit(repoPath, "rev-parse", "--verify", branchName)
	if err == nil {
		return true, nil
	}
	// If output contains "needed", it's an error.
	// Usually rev-parse exits non-zero if not found.
	return false, nil
}

// RemoveWorktree forcefully removes a worktree reference from the bare repo.
// Note: This expects the path to the repo inside the feature folder.
func RemoveWorktree(barePath, worktreePath string) error {
	// Correct way to remove worktree associated with a path from bare repo:
	// git worktree remove --force <path>
	_, err := RunGit(barePath, "worktree", "remove", "--force", worktreePath)
	return err
}

// SyncRepo fetches from remote and tries to update the current branch.
// Note: This is complex in a worktree.
// Ideally: git fetch origin, then git merge/rebase.
func SyncRepo(worktreePath string) error {
	// Verify it is a git repo
	if _, err := os.Stat(filepath.Join(worktreePath, ".git")); os.IsNotExist(err) {
		// Worktrees have a .git file, not a directory. os.Stat handles both.
	}

	fmt.Printf("Syncing %s...\n", worktreePath)
	if _, err := RunGit(worktreePath, "fetch", "--all"); err != nil {
		return fmt.Errorf("fetch failed: %v", err)
	}

	// Check if upstream exists
	_, err := RunGit(worktreePath, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		// No upstream, just fetch is fine
		fmt.Printf("  No upstream for %s, skipped pull.\n", worktreePath)
		return nil
	}

	// Attempt pull (which is fetch + merge)
	if _, err := RunGit(worktreePath, "pull"); err != nil {
		return fmt.Errorf("pull failed in %s: %v", worktreePath, err)
	}
	return nil
}

// GetStatus returns the status of a repository (dirty/clean and ahead/behind).
func GetStatus(repoPath string) (bool, string, error) {
	// Check if dirty
	status, err := RunGit(repoPath, "status", "--porcelain")
	if err != nil {
		return false, "", err
	}
	isDirty := len(status) > 0

	// Check ahead/behind
	// git rev-list --left-right --count HEAD...@{u}
	// Note: might fail if no upstream
	ab, err := RunGit(repoPath, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return isDirty, "no upstream", nil
	}

	return isDirty, ab, nil
}

// BranchName returns the current branch name.
func BranchName(repoPath string) (string, error) {
	return RunGit(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}
