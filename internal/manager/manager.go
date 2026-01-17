package manager

import (
	"fmt"
	"grove/internal/config"
	"grove/internal/git"
	"io"
	"os"
	"path/filepath"
)

type Manager struct {
	Config   *config.Config
	CacheDir string // Directory for bare repos
}

func NewManager() (*Manager, error) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		return nil, err
	}
	home, _ := os.UserHomeDir()
	return &Manager{
		Config:   cfg,
		CacheDir: filepath.Join(home, ".grove", "cache"),
	}, nil
}

func NewManagerWithConfig(path string) (*Manager, error) {
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	home, _ := os.UserHomeDir()
	return &Manager{
		Config:   cfg,
		CacheDir: filepath.Join(home, ".grove", "cache"),
	}, nil
}

func (m *Manager) SaveConfig() error {
	return m.Config.Save()
}

// --- Set Operations ---

func (m *Manager) AddSet(name string, repos []string) error {
	if _, exists := m.Config.Sets[name]; exists {
		return fmt.Errorf("set '%s' already exists", name)
	}

	home, _ := os.UserHomeDir()
	skillsDir := filepath.Join(home, ".grove", "skills", name)

	m.Config.Sets[name] = config.Set{
		Repos:     repos,
		SkillsDir: skillsDir,
	}
	return m.SaveConfig()
}

func (m *Manager) RemoveSet(name string) error {
	if _, exists := m.Config.Sets[name]; !exists {
		return fmt.Errorf("set '%s' not found", name)
	}

	// Check if any feature uses this set
	for _, feat := range m.Config.Features {
		if feat.Set == name {
			return fmt.Errorf("cannot remove set '%s': used by active feature", name)
		}
	}

	delete(m.Config.Sets, name)
	return m.SaveConfig()
}

// --- Feature Operations ---

func (m *Manager) CreateFeature(setName, featureName string) error {
	set, ok := m.Config.Sets[setName]
	if !ok {
		return fmt.Errorf("set '%s' not found", setName)
	}

	if _, exists := m.Config.Features[featureName]; exists {
		return fmt.Errorf("feature '%s' already exists", featureName)
	}

	// Prepare directories
	rootDir, err := config.ExpandPath(m.Config.RootDir)
	if err != nil {
		return err
	}

	// Structure: root_dir/set/feature
	featurePath := filepath.Join(rootDir, setName, featureName)
	if _, err := os.Stat(featurePath); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", featurePath)
	}

	cacheDir := m.CacheDir

	fmt.Printf("Creating feature '%s' for set '%s' at %s...\n", featureName, setName, featurePath)

	if err := os.MkdirAll(featurePath, 0755); err != nil {
		return fmt.Errorf("failed to create feature directory: %v", err)
	}

	// 1. Git Operations
	for _, repoURL := range set.Repos {
		bareRepo, err := git.EnsureBareRepo(repoURL, cacheDir)
		if err != nil {
			return fmt.Errorf("failed to ensure bare repo for %s: %v", repoURL, err)
		}

		repoName := git.GetRepoNameFromURL(repoURL)
		targetPath := filepath.Join(featurePath, repoName)

		fmt.Printf("Adding worktree for %s...\n", repoName)
		if err := git.CreateWorktree(bareRepo, featureName, targetPath); err != nil {
			return err
		}
	}

	// 2. Skills Initialization
	if err := m.initSkills(set, rootDir, setName); err != nil {
		fmt.Printf("Warning: Failed to init skills: %v\n", err)
	}

	// 3. Update Config
	m.Config.Features[featureName] = config.Feature{
		Path: featurePath,
		Set:  setName,
	}

	return m.SaveConfig()
}

func (m *Manager) RemoveFeature(featureName string) error {
	feat, ok := m.Config.Features[featureName]
	if !ok {
		return fmt.Errorf("feature '%s' not found", featureName)
	}

	fmt.Printf("Removing feature '%s'...\n", featureName)

	// 1. Remove Worktrees
	cacheDir := m.CacheDir

	// We need to know which repos were in this feature.
	// We iterate over the Set definition associated with the feature.
	if set, ok := m.Config.Sets[feat.Set]; ok {
		for _, repoURL := range set.Repos {
			repoName := git.GetRepoNameFromURL(repoURL)
			bareRepo := filepath.Join(cacheDir, repoName)
			worktreeStr := filepath.Join(feat.Path, repoName)

			// Try to remove from bare repo
			// Note: If folder was deleted manually, this might fail, but git worktree prune handles it.
			// We force remove.
			if err := git.RemoveWorktree(bareRepo, worktreeStr); err != nil {
				fmt.Printf("  Warning: failed to clean worktree for %s: %v\n", repoName, err)
			}
		}
	}

	// 2. Remove Directory
	if err := os.RemoveAll(feat.Path); err != nil {
		return fmt.Errorf("failed to remove directory %s: %v", feat.Path, err)
	}

	// 3. Update Config
	delete(m.Config.Features, featureName)
	return m.SaveConfig()
}

func (m *Manager) initSkills(set config.Set, rootDir, setName string) error {
	// Destination: root_dir/set/.gemini/skills (Shared for the set)
	// Actually, user said: grove_dir/worktreesetA/.gemini/skills
	// And features are siblings?
	// User said:
	// grove_dir/worktreesetA/feature1
	// grove_dir/worktreesetA/feature2
	// grove_dir/worktreesetA/.gemini/skills

	destDir := filepath.Join(rootDir, setName, ".gemini", "skills")

	// If source skills dir is defined and exists, copy it
	srcDir, err := config.ExpandPath(set.SkillsDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		// Source doesn't exist, skip
		return nil
	}

	// Remove old if exists (re-sync)
	os.RemoveAll(destDir)

	return copyDir(srcDir, destDir)
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist.
func copyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy file
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
