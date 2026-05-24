#!/bin/sh
set -e

BIN="sage"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.sage"

if [ -f "$INSTALL_DIR/$BIN" ]; then
  if [ -w "$INSTALL_DIR" ]; then
    rm "$INSTALL_DIR/$BIN"
  else
    sudo rm "$INSTALL_DIR/$BIN"
  fi
  echo "✓ removed $INSTALL_DIR/$BIN"
else
  echo "sage not found in $INSTALL_DIR"
fi

if [ -d "$CONFIG_DIR" ]; then
  printf "Remove config and history at %s? [y/N] " "$CONFIG_DIR"
  read -r answer
  case "$answer" in
    y|Y)
      rm -rf "$CONFIG_DIR"
      echo "✓ removed $CONFIG_DIR"
      ;;
    *)
      echo "  kept $CONFIG_DIR"
      ;;
  esac
fi

echo "done."
