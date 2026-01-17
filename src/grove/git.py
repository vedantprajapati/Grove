import subprocess
import os
from pathlib import Path
from typing import List, Optional

class GitError(Exception):
    pass

def run_git(args: List[str], cwd: Optional[Path] = None, check: bool = True) -> str:
    """Run a git command."""
    try:
        result = subprocess.run(
            ["git"] + args,
            cwd=cwd,
            check=check,
            capture_output=True,
            text=True
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        raise GitError(f"Git command failed: {' '.join(args)}\nStderr: {e.stderr}")

def get_repo_name_from_url(url: str) -> str:
    """Extract repository name from URL (e.g., 'git@github.com:foo/bar.git' -> 'bar')."""
    name = url.split("/")[-1]
    if name.endswith(".git"):
        name = name[:-4]
    return name

def ensure_bare_repo(url: str, cache_dir: Path) -> Path:
    """Ensure a bare clone of the repository exists in the cache."""
    repo_name = get_repo_name_from_url(url)
    bare_repo_path = cache_dir / repo_name
    
    if not bare_repo_path.exists():
        if not cache_dir.exists():
            cache_dir.mkdir(parents=True)
            
        print(f"Cloning {url} to {bare_repo_path}...")
        run_git(["clone", "--bare", url, str(bare_repo_path)])
    
    return bare_repo_path

def create_worktree(bare_repo_path: Path, branch_name: str, target_path: Path):
    """Create a worktree from the bare repo."""
    # Ensure parent dir exists
    if not target_path.parent.exists():
        target_path.parent.mkdir(parents=True)
        
    # Check if branch already exists in the bare repo
    # If it does, we just checkout. If not, we create -b.
    # For now, let's assume we always want to create a new branch or checkout existing.
    
    # Simple strategy: try to add with -b, if fails, try without -b (assuming branch exists)
    try:
        run_git(
            ["worktree", "add", "-b", branch_name, str(target_path), branch_name],
            cwd=bare_repo_path
        )
    except GitError:
        # Retry without -b (might exist remotely or locally)
        # Check if branch exists
        try:
             run_git(
                ["worktree", "add", str(target_path), branch_name],
                cwd=bare_repo_path
            )
        except GitError as e:
             # If that failed, maybe we need to fetch?
             # But for now, just raise
             raise e

def remove_worktree(target_path: Path):
    """Remove a worktree. 
    Note: 'git worktree remove' should be run from the worktree itself or the main repo.
    Since we are in a bare repo setup, we should usually run 'git worktree remove' inside the bare repo 
    pointing to the target path, OR just delete the folder and prune.
    
    The cleanest way with bare repos is:
    git worktree remove --force <path>
    """
    # We need to find which bare repo this belongs to, or just delete the folder and let `git worktree prune` cleanup eventually?
    # Actually, we should try to be clean.
    # But since we have multiple repos inside one feature folder, we iterate them?
    # This function is per-repo.
    pass # Managed at higher level or we need bare_repo_path here.

def prune_worktrees(bare_repo_path: Path):
    run_git(["worktree", "prune"], cwd=bare_repo_path)
