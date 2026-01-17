import typer
from grove.config import ConfigManager

app = typer.Typer(
    name="grove",
    help="Grove: A CLI tool to manage git worktrees across multiple repositories."
)

config = ConfigManager()

import typer
from typing import List, Optional
from pathlib import Path
from grove.manager import GroveManager

app = typer.Typer(
    name="grove",
    help="Grove: A CLI tool to manage git worktrees across multiple repositories.",
    no_args_is_help=True
)

manager = GroveManager()

# --- Top Level Commands ---

@app.command("list")
def list_items():
    """List all defined sets and active features."""
    sets = manager.config.get_sets()
    features = manager.config.get_features()
    
    typer.echo(f"Root Dir: {manager.config.get_root_dir()}")
    
    typer.echo("\nSets:")
    for name, data in sets.items():
        repos_count = len(data['repos'])
        skills_dir = data.get('skills_dir', 'N/A')
        typer.echo(f"  - {name} ({repos_count} repos)")
        typer.echo(f"    Skills: {skills_dir}")

    typer.echo("\nActive Features:")
    for name, data in features.items():
        path_uri = f"file:///{Path(data['path']).as_posix()}"
        typer.echo(f"  - {name} (Set: {data['set']})")
        typer.echo(f"    Path: \033[4m{path_uri}\033[0m")

@app.command("switch")
def switch(feature: str):
    """Switch to a feature workspace. 
    Output the path so the shell can cd into it.
    Usage: cd $(gr switch my-feature)
    """
    feature_data = manager.config.get_feature(feature)
    if not feature_data:
        typer.echo(f"Error: Feature '{feature}' not found.", err=True)
        raise typer.Exit(code=1)
    
    path = feature_data["path"]
    typer.echo(path)

# --- Add Group ---
add_app = typer.Typer(help="Add sets or features.")
app.add_typer(add_app, name="add")

@add_app.command("set")
def add_set_cmd(name: str, repos: List[str]):
    """Define a new set with a list of repo URLs."""
    try:
        manager.add_set(name, repos)
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)

@add_app.command("feature")
def add_feature_cmd(set_name: str, feature_name: str):
    """Create a new feature (worktree) from a set."""
    # Note: Typer argument parsing might get confused by "add [set] [feature]" vs "add feature [set] [feature]"
    # User requested `gr add [set] [feature]`. 
    # To support strictly `gr add set ...` and `gr add set-name feature-name`, we can use a callback or single add command.
    # But clean CLI design suggests explicit subcommands.
    # Let's try to support the user's requested syntax using a single `add` command with logic?
    # No, let's stick to `gr add set` and `gr add [set] [feature]` via a smart main command?
    # Typer doesn't easily support mixed subcommands and arguments for the parent.
    # So `gr add` is the group.
    # `gr add set ...` works.
    # `gr add my-set my-feature` requires `my-set` to be a command? No.
    # We can have a command named `feature` so it's `gr add feature [set] [feature]`.
    # AND/OR we can alias.
    # Let's implement `gr add feature` first.
    try:
        manager.create_feature(set_name, feature_name)
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)

# Hack to support `gr add [set] [feature]` if possible, but Typer makes it hard without ambiguity.
# We will tell user to use `gr add feature [set] [feature]` or we can try to make a "smart" add?
# For now, adhering to standard `gr add feature`.

# --- Remove Group ---
remove_app = typer.Typer(help="Remove sets or features.")
app.add_typer(remove_app, name="remove")

@remove_app.command("set")
def remove_set_cmd(name: str):
    """Remove a set definition."""
    try:
        manager.remove_set(name)
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)

@remove_app.command("feature")
def remove_feature_cmd(feature_name: str):
    """Remove a feature and its worktrees."""
    try:
        manager.remove_feature(feature_name)
    except Exception as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)

# Shortcuts at root level for convenience?
# User asked for `wts remove [feature]`. 
# So `gr remove featurename`.
# This conflicts with `gr remove set`.
# Verify if we can overload `gr remove`.
@app.command("remove")
def smart_remove(ctx: typer.Context, name: str):
    """
    Remove a feature (shortcut).
    To remove a set, use 'gr remove set ...'.
    """
    # If the user typed `gr remove set`, Typer routes to remove_app. 
    # If they typed `gr remove my-feature`, it might come here?
    # But `remove` is registered as a Typer group. You can't have both a command and a group with the same name usually.
    # Strategy: `gr remove` is ONLY a group.
    # We add a command `feature` to it.
    # BUT we can set `invoke_without_command=True` on the callback? No, that runs before subcommand.
    # We will stick to `gr remove feature [name]` for now to be safe, easier to implement.
    # Or we can make `gr remove` a command that inspects args. simpler to keep explicit.
    typer.echo("Please specify what to remove: 'gr remove feature [name]' or 'gr remove set [name]'.", err=True)



if __name__ == "__main__":
    app()
