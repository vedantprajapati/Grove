package tests

import (
	"grove/internal/manager"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFullWorkflow(t *testing.T) {
	// 1. Setup Temporary Directories
	tempDir := t.TempDir()
	remotesDir := filepath.Join(tempDir, "remotes")
	groveRoot := filepath.Join(tempDir, "grove")
	configFile := filepath.Join(tempDir, ".groverc")

	t.Logf("Running test in %s", tempDir)

	// 2. Create Dummy Remote Repositories
	repoNames := []string{"test-repo-1", "test-repo-2"}
	repoURLs := []string{}

	if err := os.MkdirAll(remotesDir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, name := range repoNames {
		startDir, _ := os.Getwd()
		path := filepath.Join(remotesDir, name)
		os.MkdirAll(path, 0755)

		execGit(t, path, "init")
		execGit(t, path, "config", "user.email", "test@example.com")
		execGit(t, path, "config", "user.name", "Test User")
		execGit(t, path, "branch", "-m", "main") // Ensure main branch
		os.WriteFile(filepath.Join(path, "README.md"), []byte("# Test "+name), 0644)
		execGit(t, path, "add", ".")
		execGit(t, path, "commit", "-m", "Initial commit")

		repoURLs = append(repoURLs, path)
		os.Chdir(startDir)
	}

	// 3. Initialize Manager
	mgr, err := manager.NewManagerWithConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to init manager: %v", err)
	}
	// Override RootDir in the loaded config manually and save
	mgr.Config.RootDir = groveRoot
	mgr.SaveConfig()

	// 4. Add Set
	setName := "test-set"
	err = mgr.AddSet(setName, repoURLs)
	if err != nil {
		t.Fatalf("AddSet failed: %v", err)
	}

	// Verify Set in config
	if _, ok := mgr.Config.Sets[setName]; !ok {
		t.Fatal("Set not found in config")
	}

	// 5. Add Feature
	featureName := "test-feat-1"
	err = mgr.CreateFeature(setName, featureName)
	if err != nil {
		t.Fatalf("CreateFeature failed: %v", err)
	}

	// 6. Verify Worktrees
	featurePath := filepath.Join(groveRoot, setName, featureName)
	for _, name := range repoNames {
		repoPath := filepath.Join(featurePath, name)
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			t.Errorf("Repo worktree %s missing at %s", name, repoPath)
		}
		if _, err := os.Stat(filepath.Join(repoPath, "README.md")); os.IsNotExist(err) {
			t.Errorf("README missing in worktree %s", name)
		}
	}

	// 7. Verify List/State logic
	if _, ok := mgr.Config.Features[featureName]; !ok {
		t.Error("Feature not registered in config")
	}

	// 8. Remove Feature
	err = mgr.RemoveFeature(featureName)
	if err != nil {
		t.Fatalf("RemoveFeature failed: %v", err)
	}
	if _, err := os.Stat(featurePath); !os.IsNotExist(err) {
		t.Error("Feature directory was not removed")
	}

	// 9. Remove Set
	err = mgr.RemoveSet(setName)
	if err != nil {
		t.Fatalf("RemoveSet failed: %v", err)
	}
}

func execGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Git %v in %s failed: %v\n%s", args, dir, err, out)
	}
}
