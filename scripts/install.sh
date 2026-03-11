#!/bin/sh
set -e

REPO="guerrieroriccardo/CIPTr"
BINARY="ciptr-cli"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ "$OS" != "linux" ]; then
    echo "This installer supports Linux only. For Windows, download from GitHub Releases."
    exit 1
fi

LATEST=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
    echo "Failed to fetch latest version."
    exit 1
fi

URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}_${OS}_${ARCH}.tar.gz"
echo "Downloading ${BINARY} ${LATEST} for ${OS}/${ARCH}..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -sSfL "$URL" -o "$TMP/archive.tar.gz"
tar -xzf "$TMP/archive.tar.gz" -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"
echo "Installed ${BINARY} ${LATEST} to ${INSTALL_DIR}/${BINARY}"
