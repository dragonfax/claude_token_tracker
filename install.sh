#!/bin/sh
# Install tt (Claude Code token tracker) from GitHub Releases
# Usage: curl -fsSL https://raw.githubusercontent.com/dragonfax/claude_token_tracker/main/install.sh | sh

set -e

REPO="dragonfax/claude_token_tracker"
INSTALL_DIR="${HOME}/.local/bin"
BINARY="tt"

# Detect OS and arch
OS="$(uname -s)"
ARCH="$(uname -m)"

case "${OS}" in
  Linux)  OS_NAME="linux" ;;
  Darwin) OS_NAME="darwin" ;;
  *)
    echo "Unsupported OS: ${OS}" >&2
    exit 1
    ;;
esac

case "${ARCH}" in
  x86_64)  ARCH_NAME="amd64" ;;
  aarch64|arm64) ARCH_NAME="arm64" ;;
  *)
    echo "Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

# Fetch latest release version
echo "Fetching latest release..."
LATEST_VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"

if [ -z "${LATEST_VERSION}" ]; then
  echo "Failed to fetch latest release version" >&2
  exit 1
fi

echo "Latest version: ${LATEST_VERSION}"

ARCHIVE="tt_${LATEST_VERSION#v}_${OS_NAME}_${ARCH_NAME}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE}"
CHECKSUMS_URL="${BASE_URL}/checksums.txt"

# Create temp dir
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

# Download archive and checksums
echo "Downloading ${ARCHIVE}..."
curl -fsSL -o "${TMP_DIR}/${ARCHIVE}" "${ARCHIVE_URL}"
curl -fsSL -o "${TMP_DIR}/checksums.txt" "${CHECKSUMS_URL}"

# Verify checksum
echo "Verifying checksum..."
cd "${TMP_DIR}"
if command -v sha256sum >/dev/null 2>&1; then
  grep "${ARCHIVE}" checksums.txt | sha256sum -c -
elif command -v shasum >/dev/null 2>&1; then
  grep "${ARCHIVE}" checksums.txt | shasum -a 256 -c -
else
  echo "Warning: no sha256 tool found, skipping checksum verification" >&2
fi

# Extract binary
tar -xzf "${ARCHIVE}"

# Install
mkdir -p "${INSTALL_DIR}"
mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

# Configure the PostToolUse hook
echo "Configuring Claude Code hook..."
"${INSTALL_DIR}/${BINARY}" install-hook

# PATH hint
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Note: ${INSTALL_DIR} is not in your PATH."
    echo "Add the following to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "  export PATH=\"\${HOME}/.local/bin:\${PATH}\""
    echo ""
    ;;
esac

echo ""
echo "Done! tt is installed. Run 'tt watch' to see live token usage."
