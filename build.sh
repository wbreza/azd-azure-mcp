#!/bin/bash

# Change to the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# For each directory in the script's directory, run 'azd x build --cwd <path>'
for dir in */ ; do
    if [ -d "$dir" ]; then
        echo "Building in $dir..."
        azd x build --cwd "$dir"
    fi
done