#!/bin/sh
# PostToolUse hook: forward to tt record
# Silently exits 0 if tt is not installed — never disrupts Claude.

# Look for tt at the default install location first, then fall back to PATH
if [ -x "${HOME}/.local/bin/tt" ]; then
  TT="${HOME}/.local/bin/tt"
elif command -v tt >/dev/null 2>&1; then
  TT="tt"
else
  exit 0
fi

exec "${TT}" record
