#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Create build directory if it doesn't exist
if [ ! -d "build" ]; then
  mkdir -p build
fi

# Copy appicon.png to build directory
cp appicon.png build/

wails build