#!/bin/bash
set -e

REPO="snowztech/vikusha"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

BINARY="vikusha-${OS}-${ARCH}"

echo "Fetching latest release..."
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
  echo "No releases found. Check https://github.com/${REPO}/releases"
  exit 1
fi

echo "Downloading vikusha ${LATEST} (${OS}/${ARCH})..."
if ! curl -fsSL "https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}" -o /tmp/vikusha; then
  echo "Download failed. Binary may not exist for ${OS}/${ARCH}."
  exit 1
fi
chmod +x /tmp/vikusha

if [ -w /usr/local/bin ]; then
  mv /tmp/vikusha /usr/local/bin/vikusha
  echo "Installed to /usr/local/bin/vikusha"
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  mv /tmp/vikusha "$INSTALL_DIR/vikusha"
  echo "Installed to $INSTALL_DIR/vikusha"
  if ! echo "$PATH" | grep -qF "$INSTALL_DIR"; then
    SHELL_RC="$HOME/.bashrc"
    [ -f "$HOME/.zshrc" ] && SHELL_RC="$HOME/.zshrc"
    touch "$SHELL_RC"
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC"
    echo "Run 'source $SHELL_RC' to update your PATH."
  fi
fi

echo ""
echo "Done. Try:"
echo "  vikusha version"
echo "  vikusha chat character.yaml"
