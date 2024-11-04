#!/bin/bash

# Check if version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

# Fetch the current version from the remote info.env file
INFO_URL="https://raw.githubusercontent.com/solidpulse/natsdash/refs/heads/master/info.env"
VERSION=$(curl -s $INFO_URL | grep -oP 'current_version=\K[0-9.]+')

if [ -z "$VERSION" ]; then
  echo "Failed to fetch the current version from $INFO_URL"
  exit 1
fi
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
