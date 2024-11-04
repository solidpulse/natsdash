#!/bin/bash

# Check if version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

VERSION=$1
BASE_URL="https://github.com/solidpulse/natsdash/releases/download/v${VERSION}"

# Determine OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture to the correct format
case $ARCH in
  x86_64)
    ARCH="x64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Construct the download URL
BINARY_NAME="natsdash-${OS}-${ARCH}"
DOWNLOAD_URL="${BASE_URL}/${BINARY_NAME}"

# Download the binary
echo "Downloading $DOWNLOAD_URL"
curl -L -o natsdash $DOWNLOAD_URL

# Make the binary executable
chmod +x natsdash

echo "natsdash version $VERSION has been downloaded and is ready to use."
