#!/bin/sh
set -e

REPO="yanonymousV2/sage"
BIN="sage"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux"  ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)         ARCH="amd64" ;;
  arm64 | aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

ASSET="sage_${OS}_${ARCH}"

# Fetch latest release tag
echo "→ fetching latest release..."
TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\(.*\)".*/\1/')"

if [ -z "$TAG" ]; then
  echo "Could not determine latest release. Check https://github.com/${REPO}/releases"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}.tar.gz"

echo "→ downloading sage ${TAG} (${OS}/${ARCH})..."
TMP="$(mktemp -d)"
curl -fsSL "$URL" | tar -xz -C "$TMP"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BIN" "$INSTALL_DIR/$BIN"
else
  echo "→ need sudo to install to $INSTALL_DIR"
  sudo mv "$TMP/$BIN" "$INSTALL_DIR/$BIN"
fi

rm -rf "$TMP"

echo "✓ sage ${TAG} installed to ${INSTALL_DIR}/${BIN}"
echo ""
echo "Get started:"
echo "  sage \"how is my system doing\""
