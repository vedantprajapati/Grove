import os
import shutil
from pathlib import Path
from typing import List, Optional
from grove.config import ConfigManager
from grove import git

class GroveManager:
    def __init__(self):
        self.config = ConfigManager()
        self.cache_dir = Path.home() / ".grove" / "cache"

    # --- Set Operations ---
    def add_set(self, name: str, repos: List[str]):
        """Define a new set of repositories."""
        if self.config.get_set(name):
            raise ValueError(f"Set '{name}' already exists.")
        
        self.config.add_set(name, repos)
        print(f"Set '{name}' created with {len(repos)} repositories.")

    def update_set(self, name: str, new_name: Optional[str] = None, 
                   add_repos: Optional[List[str]] = None, 
                   remove_repos: Optional[List[str]] = None):
        self.config.update_set(name, new_name, add_repos, remove_repos)
        print(f"Set updated.")

    def remove_set(self, name: str):
        self.config.remove_set(name)
        print(f"Set '{name}' removed.")

    # --- Feature Operations ---
    def create_feature(self, set_name: str, feature_name: str):
        """Create a new worktree feature for a given set."""
        set_data = self.config.get_set(set_name)
        if not set_data:
            raise ValueError(f"Set '{set_name}' not found.")
        
        if self.config.get_feature(feature_name):
            raise ValueError(f"Feature '{feature_name}' already exists.")
        
        root_dir = self.config.get_root_dir()
        feature_path = root_dir / set_name / feature_name
        
        if feature_path.exists():
             raise ValueError(f"Directory {feature_path} already exists. Please manually remove it or choose a different name.")

        print(f"Creating feature '{feature_name}' for set '{set_name}' at {feature_path}...")
        feature_path.mkdir(parents=True)

        for repo_url in set_data["repos"]:
            try:
                # 1. Ensure bare repo
                bare_repo = git.ensure_bare_repo(repo_url, self.cache_dir)
                
                # 2. Determine inner path (repo name)
                repo_name = git.get_repo_name_from_url(repo_url)
                target_path = feature_path / repo_name
                
                # 3. Create worktree
                print(f"  Adding worktree for {repo_name}...")
                git.create_worktree(bare_repo, feature_name, target_path)
                
            except Exception as e:
                print(f"  Error processing {repo_url}: {e}")
                # TODO: Cleanup logic if partial failure?
                
        # 4. Initialize Skills
        self._init_skills(set_data.get("skills_dir"), feature_path)

        # 5. Register in config
        self.config.add_feature(feature_name, set_name, str(feature_path))
        print(f"Feature '{feature_name}' ready.")

    def remove_feature(self, feature_name: str):
        feature_data = self.config.get_feature(feature_name)
        if not feature_data:
            raise ValueError(f"Feature '{feature_name}' not found.")
        
        path = Path(feature_data["path"])
        set_name = feature_data["set"]
        set_data = self.config.get_set(set_name)
        
        print(f"Removing feature '{feature_name}' at {path}...")
        
        # 1. Remove worktrees
        # We need to cleanup the bare repos basically. 
        # The cleanest is to go into each repo dir and `git worktree remove .`? 
        # But we deleted the lines in git.py.
        # Let's retry: for each repo in the Set, go to its bare repo and remove the worktree matching this path.
        if set_data: 
             for repo_url in set_data["repos"]:
                repo_name = git.get_repo_name_from_url(repo_url)
                bare_repo = self.cache_dir / repo_name
                target_path = path / repo_name
                
                if bare_repo.exists():
                     try:
                         # Force remove the worktree reference
                         git.run_git(["worktree", "remove", "--force", str(target_path)], cwd=bare_repo)
                     except Exception as e:
                         print(f"  Warning: Failed to remove worktree for {repo_name}: {e}")

        # 2. Delete directory (if allowed)
        if path.exists():
            shutil.rmtree(path)
            
        # 3. Update config
        self.config.remove_feature(feature_name)
        print(f"Feature '{feature_name}' removed.")

    def _init_skills(self, skills_src: Optional[str], feature_path: Path):
        """Initialize skills in the feature directory."""
        if not skills_src:
            return

        src_path = Path(skills_src).expanduser()
        if not src_path.exists():
            # If source doesn't exist, maybe we just create the dest folder empty?
            # Or just warn.
            return 

        dest_path = feature_path / ".agent" / "skills"
        dest_path.parent.mkdir(parents=True, exist_ok=True)
        
        if dest_path.exists():
            shutil.rmtree(dest_path)
            
        # Join the skills from the set to the local skills
        # We can symlink or copy. Copy is safer for now to avoid side effects? 
        # User said "grove_dir/worktreesetB/.gemini/skills".
        # Let's copy.
        shutil.copytree(src_path, dest_path)
        print(f"  Skills initialized from {src_path}")
