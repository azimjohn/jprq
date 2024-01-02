#!/bin/bash

set -e

URL_PREFIX="https://github.com/azimjohn/jprq/releases/download/2.3"
INSTALL_DIR="/usr/local/bin"

case "$(uname -sm)" in
  "Darwin x86_64") FILENAME="jprq-darwin-amd64" ;;
  "Darwin arm64") FILENAME="jprq-darwin-arm64" ;;
  "Linux x86_64") FILENAME="jprq-linux-amd64" ;;
  "Linux i686") FILENAME="jprq-linux-386" ;;
  "Linux armv7l") FILENAME="jprq-linux-arm" ;;
  "Linux aarch64") FILENAME="jprq-linux-arm64" ;;
  *) echo "Unsupported architecture: $(uname -sm)" >&2; exit 1 ;;
esac

echo "Downloading $FILENAME from github releases"
if ! curl -sSLf "$URL_PREFIX/$FILENAME" -o "$INSTALL_DIR/jprq"; then
  echo "Failed to write to $INSTALL_DIR; try with sudo" >&2
  exit 1
fi

if ! chmod +x "$INSTALL_DIR/jprq"; then
  echo "Failed to set executable permission on $INSTALL_DIR/jprq" >&2
  exit 1
fi

echo "jprq is successfully installed"
