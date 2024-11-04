#!/bin/bash

# Path to the local info.env file
INFO_ENV_PATH="./info.env"

# Extract the current version from info.env
CURRENT_VERSION=$(grep -E "^current_version=" "$INFO_ENV_PATH" | cut -d'=' -f2 | tr -d '"')

#if FGOOS contains ubuntu, then set FGOOS to linux
#if FGOOS contains windows, then set FGOOS to windows
#if FGOOS contains mac, then set FGOOS to darwin
if [[ $FGOOS == *"ubuntu"* ]]; then
    FGOOS="linux"
elif [[ $FGOOS == *"windows"* ]]; then
    FGOOS="windows"
elif [[ $FGOOS == *"mac"* ]]; then
    FGOOS="darwin"
fi


# Build the application with the new version and respect GOOS and GOARCH
go build -gcflags=all="-N -l" -o natsdash-${FGOOS}-${FGOARCH} -ldflags "-X github.com/solidpulse/natsdash/ds.Version=$CURRENT_VERSION"
