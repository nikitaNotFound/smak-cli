# Smak CLI

A command-line tool for easier git interaction with interactive interfaces.

## Features

- **Interactive Branch Management** (`smak b`): Browse, select, and delete branches with an intuitive interface
- **Commit Browser** (`smak c`): Navigate commits with full diff viewing
- **Commit Amend** (`smak c am`): Stage all changes and amend to latest commit with optional push
- **Git Repository Integration**: Works with any git repository

## Installation

### Build from Source

1. Clone the repository:
   ```bash
   git clone <your-repo-url>
   cd smak-cli
   ```

2. Build the binary:
   ```bash
   go build -o smak
   ```

3. Install globally (optional):
   ```bash
   # On macOS/Linux - copy to a directory in your PATH
   sudo cp smak /usr/local/bin/
   
   # Or add to your shell profile
   export PATH=$PATH:/path/to/smak-cli
   ```

### Using Go Install

If you have Go installed, you can install directly:

```bash
go install github.com/nikitaNotFound/smak-cli@latest
```

## Usage

Make sure you're in a git repository before using any commands.

### Commands

- `smak b` - Interactive branch browser and manager
- `smak c` - Interactive commit browser
- `smak c am` - Stage all changes and amend to latest commit
- `smak help` - Show help information

### Branch Management (`smak b`)

- Navigate with arrow keys
- Press `d` to select/deselect branches for deletion (shown in red)
- Press `Enter` to confirm deletion of selected branches
- Press `q` to quit

### Commit Browser (`smak c`)

- Navigate commits with arrow keys
- Press `Enter` to view full commit details and diff
- In diff view:
  - Use arrow keys or `j`/`k` to scroll
  - `Page Up`/`Page Down` for faster navigation
  - `Escape` to return to commit list
- Press `q` to quit

### Commit Amend (`smak c am`)

Quickly stage all unstaged changes and amend them to the latest commit with the same message.

**Basic usage:**
```bash
smak c am
```

**With automatic push:**
```bash
smak c am -p
# or
smak c am --push
```

**Options:**
- `-p, --push` - Push the amended commit to origin with force after amending

This command is useful for quickly incorporating additional changes into your most recent commit without having to manually stage files and run git commands.

## Requirements

- Git repository
- Go 1.21+ (for building from source)

## Contributing

Feel free to submit issues and pull requests.