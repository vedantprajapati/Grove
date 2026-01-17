package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ConfigFileName = ".groverc"

type Config struct {
	filePath string             // Private field to store where to save
	RootDir  string             `json:"root_dir"`
	Sets     map[string]Set     `json:"sets"`
	Features map[string]Feature `json:"features"`
}

type Set struct {
	Repos     []string `json:"repos"`
	SkillsDir string   `json:"skills_dir"`
}

type Feature struct {
	Path string `json:"path"`
	Set  string `json:"set"`
}

// DefaultConfig returns defaults.
func DefaultConfig() Config {
	path, _ := GetDefaultPath()
	home, _ := os.UserHomeDir()
	return Config{
		filePath: path,
		RootDir:  filepath.Join(home, "Documents", "Grove"),
		Sets:     make(map[string]Set),
		Features: make(map[string]Feature),
	}
}

func GetDefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigFileName), nil
}

// LoadConfig loads from a specific path. If path is empty, uses default.
func LoadConfig(customPath string) (*Config, error) {
	var path string
	if customPath != "" {
		path = customPath
	} else {
		var err error
		path, err = GetDefaultPath()
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := DefaultConfig()
		cfg.filePath = path
		return &cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.filePath = path // Ensure we remember where to save

	// Ensure maps are initialized if nil
	if cfg.Sets == nil {
		cfg.Sets = make(map[string]Set)
	}
	if cfg.Features == nil {
		cfg.Features = make(map[string]Feature)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	if c.filePath == "" {
		return fmt.Errorf("no config file path set")
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.filePath, data, 0644)
}

// ExpandPath expands the tilde (~) in the path to the user's home directory.
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}
