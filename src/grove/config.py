import json
import os
from pathlib import Path
from typing import Dict, List, Optional, Any

CONFIG_FILE = Path.home() / ".groverc"

class ConfigManager:
    def __init__(self):
        self.config_path = CONFIG_FILE
        self.config: Dict[str, Any] = self._load_config()

    def _load_config(self) -> Dict[str, Any]:
        """Load configuration from JSON file or return defaults."""
        if not self.config_path.exists():
            return self._default_config()
        
        try:
            with open(self.config_path, "r") as f:
                return json.load(f)
        except json.JSONDecodeError:
            return self._default_config()

    def _default_config(self) -> Dict[str, Any]:
        """Return default configuration structure."""
        return {
            "root_dir": str(Path.home() / "Documents" / "Grove"),
            "sets": {},
            "features": {}
        }

    def save(self):
        """Save current configuration to file."""
        with open(self.config_path, "w") as f:
            json.dump(self.config, f, indent=2)

    def get_root_dir(self) -> Path:
        """Get the root directory for Grove."""
        return Path(self.config.get("root_dir", str(Path.home() / "Documents" / "Grove")))

    def set_root_dir(self, path: str):
        self.config["root_dir"] = str(path)
        self.save()

    # --- Sets Management ---
    def get_sets(self) -> Dict[str, Any]:
        return self.config.get("sets", {})

    def get_set(self, name: str) -> Optional[Dict[str, Any]]:
        return self.config.get("sets", {}).get(name)

    def add_set(self, name: str, repos: List[str]):
        if "sets" not in self.config:
            self.config["sets"] = {}
        
        # Determine skills dir default location
        skills_dir = Path.home() / ".grove" / "skills" / name
        
        self.config["sets"][name] = {
            "repos": repos,
            "skills_dir": str(skills_dir)
        }
        self.save()

    def update_set(self, name: str, new_name: Optional[str] = None, 
                   add_repos: Optional[List[str]] = None, 
                   remove_repos: Optional[List[str]] = None):
        sets = self.config.get("sets", {})
        if name not in sets:
            raise ValueError(f"Set '{name}' not found.")

        set_data = sets[name]
        
        if add_repos:
            # Avoid duplicates
            current_repos = set(set_data["repos"])
            current_repos.update(add_repos)
            set_data["repos"] = list(current_repos)
        
        if remove_repos:
            set_data["repos"] = [r for r in set_data["repos"] if r not in remove_repos]

        if new_name and new_name != name:
            if new_name in sets:
                raise ValueError(f"Set '{new_name}' already exists.")
            sets[new_name] = set_data
            del sets[name]
            # Verify if we need to update features referencing this set? 
            # Implement logic to update features referencing 'name' to 'new_name' if needed
            # But usually we only rely on the current definition for NEW features.
            # Existing features already store their path and set name. 
            # Ideally we update references too.
            self._update_set_references(name, new_name)

        self.save()

    def remove_set(self, name: str):
        sets = self.config.get("sets", {})
        if name not in sets:
            raise ValueError(f"Set '{name}' not found.")
        
        # Check if any active features use this set
        features = self.config.get("features", {})
        for feat_name, feat_data in features.items():
            if feat_data.get("set") == name:
                raise ValueError(f"Cannot remove set '{name}' because it is used by active feature '{feat_name}'.")

        del sets[name]
        self.save()

    def _update_set_references(self, old_name: str, new_name: str):
        """Update features to point to the new set name."""
        features = self.config.get("features", {})
        for _, feat_data in features.items():
            if feat_data.get("set") == old_name:
                feat_data["set"] = new_name

    # --- Features Management ---
    def get_features(self) -> Dict[str, Any]:
        return self.config.get("features", {})

    def get_feature(self, name: str) -> Optional[Dict[str, Any]]:
        return self.config.get("features", {}).get(name)

    def add_feature(self, feature_name: str, set_name: str, path: str):
        if "features" not in self.config:
            self.config["features"] = {}
        
        self.config["features"][feature_name] = {
            "set": set_name,
            "path": str(path)
        }
        self.save()

    def remove_feature(self, feature_name: str):
        if "features" in self.config and feature_name in self.config["features"]:
            del self.config["features"][feature_name]
            self.save()
