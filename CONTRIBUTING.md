# Contributing

## Prerequisites

- Go 1.21+
- [GoReleaser](https://goreleaser.com/install/) (for cutting releases)

## Building

```sh
go build ./cmd/tt
```

Or install to `~/.local/bin`:

```sh
go build -o ~/.local/bin/tt ./cmd/tt
```

## Running tests

```sh
go test ./...
```

The TUI package uses golden-file snapshot tests in `internal/tui/testdata/`. If your changes intentionally alter TUI output, update the snapshots by running:

```sh
UPDATE_SNAPSHOTS=true go test ./internal/tui/...
```

## Running the hook locally

To test `tt record` end-to-end without a live Claude session, pipe a sample PostToolUse payload directly:

```sh
echo '{"session_id":"test","tool_name":"Read","tool_use_id":"abc","response":{"type":"text","text":"hello"}}' | tt record
tt log
```

## Manual hook setup

If you build from source and want to wire the hook yourself:

```sh
go build -o ~/.local/bin/tt ./cmd/tt
tt install-hook      # adds PostToolUse hook to ~/.claude/settings.json
```

To remove it:

```sh
tt uninstall-hook    # removes tt entries from ~/.claude/settings.json PostToolUse hooks
```

## Database

The SQLite database lives at `~/.claude/token_tracker/token_tracker.db`. To inspect it directly:

```sh
sqlite3 ~/.claude/token_tracker/token_tracker.db
```

Schema is defined in `internal/db/schema.go`.

## Cutting a release

Releases are built with [GoReleaser](https://goreleaser.com) and published to GitHub Releases. The release assets are:

- `tt_<version>_darwin_amd64.tar.gz`
- `tt_<version>_darwin_arm64.tar.gz`
- `tt_<version>_linux_amd64.tar.gz`
- `tt_<version>_linux_arm64.tar.gz`
- `checksums.txt` (sha256)

### Steps

1. Make sure your working tree is clean and all changes are merged to `main`.

2. Create and push an annotated tag:

   ```sh
   git tag -a v0.x.y -m "Release v0.x.y"
   git push origin v0.x.y
   ```

3. Run GoReleaser:

   ```sh
   goreleaser release --clean
   ```

   GoReleaser reads `.goreleaser.yml`, builds all platform binaries (CGO disabled), packages them as `.tar.gz` archives, generates `checksums.txt`, and publishes everything to the GitHub release.

   You need a `GITHUB_TOKEN` environment variable with `repo` scope:

   ```sh
   GITHUB_TOKEN=<your-token> goreleaser release --clean
   ```

4. Verify the release on GitHub. The `install.sh` one-liner fetches the `latest` release via the GitHub API, so the new version will be picked up automatically.

### Snapshot builds (no tag required)

To test the release pipeline locally without publishing:

```sh
goreleaser release --snapshot --clean
```

Artifacts are written to `dist/`. The snapshot version is named `<last-tag>-next`.

## Project layout

```
cmd/tt/         CLI entry point and hook install/uninstall logic
internal/db/    SQLite schema and query helpers
internal/hook/  PostToolUse JSON parser
internal/tui/   Bubble Tea TUI views (watch, log, stats)
docs/           Configuration and setup documentation
install.sh      One-liner install script (downloads release binary)
uninstall.sh    Reverses install.sh
.goreleaser.yml GoReleaser build configuration
```
