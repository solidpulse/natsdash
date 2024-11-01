#!/bin/bash

# Path to the local info.env file
INFO_ENV_PATH="./info.env"

# Extract the current version from info.env
CURRENT_VERSION=$(grep -E "^current_version=" "$INFO_ENV_PATH" | cut -d'=' -f2 | tr -d '"')

# Increment the patch version
IFS='.' read -r -a version_parts <<< "$CURRENT_VERSION"
((version_parts[2]++))
NEW_VERSION="${version_parts[0]}.${version_parts[1]}.${version_parts[2]}"

# Update the info.env file with the new version
sed -i "s/^current_version=.*/current_version=\"$NEW_VERSION\"/" "$INFO_ENV_PATH"

# Build the application with the new version
LDFLAGS="-ldflags '-X github.com/solidpulse/natsdash/ds.Version=$NEW_VERSION'"
go build -gcflags=all="-N -l" -o natsdash $LDFLAGS

# Run the application
./natsdash
