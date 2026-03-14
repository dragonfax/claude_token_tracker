# tt — Claude Code Token Tracker

`tt` tracks the size of tool and subagent responses flowing into the main context while Claude Code is running. This helps you identify which tools and subagents are the worst context-bloat offenders, so you can make better decisions about how you use them.

## Quick Install

**Option 1 — curl one-liner** (recommended — installs binary and configures hook):

```sh
curl -fsSL https://raw.githubusercontent.com/dragonfax/claude_token_tracker/main/install.sh | sh
```

Installs `tt` to `~/.local/bin/tt` and adds the PostToolUse hook to `~/.claude/settings.json`.

**Option 2 — Go install** (for contributors or manual setups):

```sh
go install github.com/reshophq/token-tracker/cmd/tt@latest
tt install-hook   # adds the PostToolUse hook to ~/.claude/settings.json
```

See [docs/configuration.md](docs/configuration.md) for full setup details.

## How it works

Claude Code's `PostToolUse` hook fires after every tool call. `tt record` is invoked as that hook, reads the response payload from stdin, and records its byte size to a local SQLite database. Because only responses entering the **main context** are meaningful for context budgeting, `tt` distinguishes between:

- **Main context calls** — tool calls made directly by the main agent
- **Subagent calls** — tool calls made inside a subagent (tracked but de-emphasized, since those responses are summarized before reaching the main context)
- **Agent tool calls** — when the main agent invokes a subagent, the subagent's final output is what enters the main context, and that is recorded as a main context call

## Uninstall

```sh
curl -fsSL https://raw.githubusercontent.com/dragonfax/claude_token_tracker/main/uninstall.sh | sh
```

Removes the PostToolUse hook from `~/.claude/settings.json` and deletes the `tt` binary from `~/.local/bin`. The database at `~/.claude/token_tracker/token_tracker.db` is left in place — remove it manually if desired:

```sh
rm -rf ~/.claude/token_tracker
```

You can also remove just the hook without uninstalling the binary:

```sh
tt uninstall-hook
```

## Commands

```bash
tt record         # called by the PostToolUse hook — do not run manually
tt watch          # live feed of tool calls across all sessions
tt log            # scrollable historical log with filtering
tt stats          # aggregate stats by tool name (main context only)
tt install-hook   # add PostToolUse hook to ~/.claude/settings.json
tt uninstall-hook # remove PostToolUse hook from ~/.claude/settings.json
```

### tt watch

Live tail of all tool calls as they happen, across all concurrent sessions. Subagent tool calls are hidden by default (press `s` to show them dimmed).

### tt log

Scrollable historical log. Press `tab` to cycle time windows (24h / 48h / 7d / all), `s` to toggle subagent rows, `e` for errors-only, `j`/`k` or `↑`/`↓` to scroll, `g`/`G` for top/bottom.

### tt stats

Aggregate totals and averages per tool, main context only. Press `tab` to cycle time windows (24h / 48h / 7d). Auto-refreshes every 5 seconds.

## Data storage

All data is written to `~/.claude/token_tracker/token_tracker.db` (SQLite, WAL mode). The file is created automatically on first use. Multiple concurrent hook invocations (parallel subagents, multiple sessions) are safe.

Errors encountered by `tt record` are written to an `errors` table in the same database and never surfaced to Claude Code.

## Configuration

See [docs/configuration.md](docs/configuration.md) for installation and hook setup instructions, including notes on using `tt` with Claude Code in the desktop GUI.
