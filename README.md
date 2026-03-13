# difr

A platform-independent, local code review tool that visualizes Git diffs in a GitHub-style UI with commenting and Claude Code integration.

![Overview](.readme/difr_overview.jpg)

## Features

- **GitHub-style Diff Viewer** — Split / Unified view with syntax highlighting
- **Comments** — File-level and line-level comments with Markdown / JSON export
- **Claude Code Integration** — Real-time WebSocket chat and automated code review
- **File Browser** — Browse all tracked files with syntax highlighting
- **Single Binary** — Frontend embedded via Go embed, no separate install required

## Installation

### GitHub Releases

Download the latest binary for your platform from [GitHub Releases](https://github.com/shimasan0x00/difr/releases):

| Platform | File |
|----------|------|
| Linux (x64) | `difr-linux-amd64` |
| Linux (ARM64) | `difr-linux-arm64` |
| macOS (Intel) | `difr-darwin-amd64` |
| macOS (Apple Silicon) | `difr-darwin-arm64` |
| Windows (x64) | `difr-windows-amd64.exe` |

On Linux / macOS, make the binary executable after downloading:

```bash
chmod +x difr-*
```

### go install

```bash
go install github.com/shimasan0x00/difr/cmd/difr@latest
```

### Build from Source

Prerequisites: Go 1.25+, Node.js 22+, [Task](https://taskfile.dev/)

```bash
git clone https://github.com/shimasan0x00/difr.git
cd difr
task install
task build
# ./difr binary is generated
```

## Usage

```
difr [flags] [commit | from to | staged | working]
```

| Command | Description |
|---------|-------------|
| `difr` | Latest commit diff (`HEAD~1..HEAD`) |
| `difr <commit>` | Diff of a specific commit (`<commit>~1..<commit>`) |
| `difr <base> <compare>` | Diff between two refs (branches, tags, or commits) |
| `difr staged` | Staged changes (`git diff --cached`) |
| `difr working` | Unstaged working tree changes |
| `git diff \| difr` | Read diff from stdin pipe |

### Examples

```bash
# Compare two branches
difr main feature/new-api

# Review staged changes before committing
difr staged

# View a specific commit
difr abc1234

# Pipe from git diff with custom options
git diff --ignore-all-space main | difr
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port, -p` | `3333` | Server port |
| `--host` | `127.0.0.1` | Bind address |
| `--mode, -m` | `split` | Display mode (`split` / `unified`) |
| `--no-open` | `false` | Don't open browser automatically |
| `--no-claude` | `false` | Disable Claude Code integration |
| `--watch, -w` | `false` | Watch for file changes (experimental) |
| `--claude-timeout` | `5m` | Timeout for Claude CLI operations (e.g. `10m`, `300s`) |

## Tech Stack

| Area | Technology |
|------|------------|
| Backend | Go 1.25 / Chi v5 / cobra |
| Frontend | React 19 / TypeScript 5.9 / Vite 7 |
| Styling | Tailwind CSS v4 |
| State Management | Zustand v5 |
| Syntax Highlighting | Shiki (github-dark theme) |
| WebSocket | coder/websocket |
| Testing | testify / Vitest / React Testing Library / Playwright |

## Development

Prerequisites: Go 1.25+, Node.js 22+, [Task](https://taskfile.dev/)

```bash
task install           # Install dependencies (go mod tidy + npm install)

# Development
task dev               # Start Go (:3333) + Vite (:5173) dev servers

# Testing
task test              # All tests (Go + Frontend)
task test:backend      # Go tests only (-race enabled)
task test:frontend     # Vitest only
task test:e2e          # Playwright E2E tests
task test:coverage     # Tests with coverage report

# Quality
task lint              # Run all linters (go vet + eslint)

# Build
task build             # Production binary (single binary with embedded frontend)
task clean             # Clean build artifacts
```

## License

[MIT](LICENSE)
