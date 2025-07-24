# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Smak CLI is an interactive Git tool built in Go using the Cobra CLI framework and Bubble Tea for terminal UIs. It provides two main features:
- Interactive branch management (`smak b`)
- Interactive commit browsing with diff viewer (`smak c`)

## Development Commands

### Building and Running
```bash
# Build the binary
make build
# or
go build -o smak .

# Run directly
go run . b    # branches
go run . c    # commits

# Install locally for testing
make install-user  # installs to ~/bin
```

### Testing and Quality
```bash
# Run tests
make test
go test ./...

# Run tests with race detection
make test-race

# Format code
make fmt
go fmt ./...

# Vet code
make vet
go vet ./...

# All checks before commit
make check  # runs fmt, vet, test
```

### Other Useful Commands
```bash
# Clean build artifacts
make clean

# Tidy dependencies
make tidy
go mod tidy

# Cross-platform builds
make build-all
```

## Architecture

### Project Structure
- `main.go` - Entry point that calls cmd.Execute()
- `cmd/` - Cobra command definitions and TUI implementations
  - `root.go` - Root command setup and git repo validation
  - `branches.go` - Branch management TUI with Bubble Tea
  - `commits.go` - Commit browser TUI with diff viewer
  - `help.go` - Help command
- `internal/` - Core git operations
  - `git.go` - Git command wrappers and data structures

### Key Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - Terminal UI framework
- `github.com/charmbracelet/bubbles` - Pre-built TUI components (list, viewport)
- `github.com/charmbracelet/lipgloss` - Terminal styling

### TUI Architecture
Both branch and commit browsers follow the Bubble Tea pattern:
- Model holds state (list data, UI state, selections)
- Update handles messages (key presses, window resize)
- View renders the current state
- Custom delegates handle item rendering with special styling

### Git Integration
All git operations are in `internal/git.go`:
- Uses `exec.Command` to call git CLI
- Parses git output with custom formats
- Branch operations: list, checkout, delete (with fallback to force delete)
- Commit operations: log, show with diff

### Data Models
- `Branch` struct: name, commit info, ahead/behind counts
- `Commit` struct: hash, message, date, author
- Both sorted by date (most recent first)

## Development Notes

- All git operations require being in a git repository (checked by `checkGitRepo()`)
- TUIs use alt screen mode for full-screen experience
- Branch deletion supports bulk selection with confirmation
- Commit viewer includes scrollable diff with syntax highlighting from git
- Error handling preserves user experience (logs errors, continues operation)
- Window resize is handled dynamically for responsive UI