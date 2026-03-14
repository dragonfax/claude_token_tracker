#!/bin/sh
# Uninstall tt (Claude Code token tracker)
# Usage: sh uninstall.sh

set -e

INSTALL_DIR="${HOME}/.local/bin"
BINARY="tt"
BINARY_PATH="${INSTALL_DIR}/${BINARY}"

# Remove the PostToolUse hook from ~/.claude/settings.json
if command -v "${BINARY_PATH}" >/dev/null 2>&1; then
  echo "Removing PostToolUse hook from ~/.claude/settings.json..."
  "${BINARY_PATH}" uninstall-hook
elif command -v tt >/dev/null 2>&1; then
  echo "Removing PostToolUse hook from ~/.claude/settings.json..."
  tt uninstall-hook
else
  echo "tt binary not found — skipping hook removal"
  echo "You may need to manually remove the tt hook from ~/.claude/settings.json"
fi

# Remove the binary
if [ -f "${BINARY_PATH}" ]; then
  echo "Removing ${BINARY_PATH}..."
  rm -f "${BINARY_PATH}"
  echo "Removed ${BINARY_PATH}"
else
  echo "Binary not found at ${BINARY_PATH} — already removed or installed elsewhere"
fi

echo ""
echo "Done. Data at ~/.claude/token_tracker/token_tracker.db was NOT removed."
echo "To also remove the database: rm -rf ~/.claude/token_tracker"
