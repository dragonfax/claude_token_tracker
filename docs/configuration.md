# Configuring tt with Claude Code

## Installation

### Option 1 — curl one-liner (no Go required)

```sh
curl -fsSL https://raw.githubusercontent.com/dragonfax/claude_token_tracker/main/install.sh | sh
```

Downloads the pre-built binary for your platform to `~/.local/bin/tt` and automatically adds the PostToolUse hook to `~/.claude/settings.json`. No further configuration needed.

### Option 2 — Go install (for contributors)

```bash
go install github.com/reshophq/token-tracker/cmd/tt@latest
# then wire the hook:
tt install-hook
```

`tt install-hook` adds the PostToolUse hook to `~/.claude/settings.json` using the absolute path of the installed binary (required for desktop GUI users).

### Manual build

```bash
go build -o ~/.local/bin/tt ./cmd/tt
tt install-hook
```

## Compatibility

| Environment | Supported | Notes |
|---|---|---|
| Claude Code (CLI) | Yes | Full support |
| Claude Code in the desktop GUI | Yes | Use absolute path in hook config (see below) |
| Claude.ai desktop chat app | No | No hooks system |

The Claude.ai desktop chat app is a separate product from Claude Code and does not have a hooks system. Only Claude Code (whether run from the terminal or embedded in the desktop GUI) supports hooks.

## Hook configuration

### Claude Code CLI

When running Claude Code from the terminal, `tt` just needs to be on your shell's `PATH`. Add the PostToolUse hook to a project's `.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "tt record"
          }
        ]
      }
    ]
  }
}
```

### Claude Code in the desktop GUI

Desktop apps on macOS launch with a minimal environment that does not source your shell configuration, so `tt` may not be on the PATH the desktop app uses. Use the **full absolute path** to the binary instead:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/YOUR_USERNAME/bin/tt record"
          }
        ]
      }
    ]
  }
}
```

Replace `/Users/YOUR_USERNAME/bin/tt` with the actual path where you installed the binary. Using the absolute path works correctly in both the CLI and desktop GUI, so it is the safest choice if you use both.

The `matcher` field being empty means the hook fires for every tool call.

## Global configuration

Once tested and stable, move the hook entry to `~/.claude/settings.json` to apply it across all projects and both the CLI and desktop GUI.

## Data storage

All data is stored in `~/.claude/token_tracker/token_tracker.db` (SQLite, WAL mode).
The directory and database are created automatically on the first `tt record` call.

## Viewing data

```bash
tt watch    # live feed of tool calls as they happen
tt log      # scrollable historical log (↑↓ or j/k to scroll, tab to change time window)
tt stats    # aggregate stats by tool (main context only)
```

## Error log

Parse and DB errors are written to the `errors` table in the same database.
View them in `tt log` by pressing `e` to toggle errors-only mode.
They also appear inline (in red) in `tt watch`.
