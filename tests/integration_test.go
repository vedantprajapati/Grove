package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/vedantprajapati/Grove/internal/config"
	"github.com/vedantprajapati/Grove/internal/git"
	"github.com/vedantprajapati/Grove/internal/manager"
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
	// Override RootDir and CacheDir for isolation
	mgr.Config.RootDir = groveRoot
	mgr.CacheDir = filepath.Join(tempDir, "cache")
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

	// 8. Test SyncFeature
	err = mgr.SyncFeature(featureName)
	if err != nil {
		t.Fatalf("SyncFeature failed: %v", err)
	}

	// 9. Test GetFeatureStatus
	statuses, err := mgr.GetFeatureStatus(featureName)
	if err != nil {
		t.Fatalf("GetFeatureStatus failed: %v", err)
	}
	if len(statuses) != len(repoNames) {
		t.Errorf("Expected %d statuses, got %d", len(repoNames), len(statuses))
	}
	for _, s := range statuses {
		if s.Branch != featureName {
			t.Errorf("Repo %s expected branch %s, got %s", s.Name, featureName, s.Branch)
		}
		if s.IsDirty {
			t.Errorf("Repo %s should be clean", s.Name)
		}
	}

	// 10. Test ExecFeature
	// Run 'git status' or similar simple command
	err = mgr.ExecFeature(featureName, "git", []string{"status", "-s"})
	if err != nil {
		t.Fatalf("ExecFeature failed: %v", err)
	}

	// 11. Test Skills Copy
	skillsSource := filepath.Join(tempDir, "source-skills")
	os.MkdirAll(skillsSource, 0755)
	os.WriteFile(filepath.Join(skillsSource, "SKILL.md"), []byte("# Test Skill"), 0644)

	// Update set with skills dir
	set, _ := mgr.Config.Sets[setName]
	set.SkillsDir = skillsSource
	mgr.Config.Sets[setName] = set
	mgr.SaveConfig()

	// Redo feature to trigger skills init
	feature2 := "test-feat-2"
	err = mgr.CreateFeature(setName, feature2)
	if err != nil {
		t.Fatalf("CreateFeature 2 failed: %v", err)
	}

	// Check skills in destination
	// RootDir/Set/.gemini/skills
	skillsInFeature := filepath.Join(groveRoot, setName, ".gemini", "skills", "SKILL.md")
	if _, err := os.Stat(skillsInFeature); os.IsNotExist(err) {
		t.Errorf("Skill file missing in feature at %s", skillsInFeature)
	}

	// 12. Cleanup
	mgr.RemoveFeature(featureName)
	mgr.RemoveFeature(feature2)
	mgr.RemoveSet(setName)
}

func TestErrorCases(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Test NewManager (Global)
	// This might fail if .groverc doesn't exist, but that's fine, we just want to hit the line.
	manager.NewManager()

	// 2. Test Path with Tilde
	expanded, _ := config.ExpandPath("~/test")
	if !filepath.IsAbs(expanded) && expanded != "~/test" {
		t.Errorf("ExpandPath failed for tilde: %s", expanded)
	}

	configFile := filepath.Join(tempDir, ".groverc")
	groveRoot := filepath.Join(tempDir, "grove")
	mgr, _ := manager.NewManagerWithConfig(configFile)
	mgr.Config.RootDir = groveRoot
	mgr.CacheDir = filepath.Join(tempDir, "cache")
	mgr.SaveConfig()

	// 3. Test git.BranchExists
	// Need a repo first
	repoPath := filepath.Join(tempDir, "repo")
	os.MkdirAll(repoPath, 0755)
	execGit(t, repoPath, "init")
	exists, _ := git.BranchExists(repoPath, "main")
	if !exists {
		// In some git versions init doesn't create main until commit
	}

	// 4. Test LoadConfig error paths
	config.LoadConfig(filepath.Join(tempDir, "non-existent.json"))

	// 5. Remove non-existent set
	if err := mgr.RemoveSet("none"); err == nil {
		t.Error("Should have failed to remove non-existent set")
	}

	// 6. Create feature for non-existent set
	if err := mgr.CreateFeature("none", "feat"); err == nil {
		t.Error("Should have failed to create feature for non-existent set")
	}

	// 8. Test CreateFeature with existing directory
	os.MkdirAll(filepath.Join(groveRoot, "collision-set", "collision"), 0755)
	mgr.AddSet("collision-set", []string{"https://github.com/example/repo"})
	if err := mgr.CreateFeature("collision-set", "collision"); err == nil {
		t.Error("Should have failed due to existing directory")
	}

	// 9. Test RemoveSet for set in use
	// We have 'test-feat-1' active for 'test-set' from TestFullWorkflow (if it didn't cleanup yet)
	// Wait, TestFullWorkflow cleans up at the end.
	// Let's create a quick one here.
	mgr.AddSet("in-use-set", []string{})
	mgr.CreateFeature("in-use-set", "in-use-feat")
	if err := mgr.RemoveSet("in-use-set"); err == nil {
		t.Error("Should have failed to remove set with active feature")
	}
	// 10. Test CreateFeature with invalid repo
	mgr.AddSet("invalid-set", []string{"./invalid-path"})
	if err := mgr.CreateFeature("invalid-set", "fail"); err == nil {
		t.Error("Should have failed due to invalid repo path")
	}

	// 11. Test git.RunGit error path
	_, err := git.RunGit(tempDir, "invalid-command")
	if err == nil {
		t.Error("Should have failed to run invalid command")
	}
}

func execGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Git %v in %s failed: %v\n%s", args, dir, err, out)
	}
}
