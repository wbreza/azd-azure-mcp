#!/bin/bash

# Change to the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# For each directory in the script's directory, run 'azd x watch --cwd <path>' in the background
declare -a pids
for dir in */ ; do
    if [ -d "$dir" ]; then
        echo "Building in $dir..."
        azd x watch --cwd "$dir" &
        pids+=("$!")
    fi
done

# Wait for all background jobs to finish
wait