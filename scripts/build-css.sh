#!/bin/bash

# Build Tailwind CSS using standalone CLI
# Usage: ./scripts/build-css.sh [--watch]

set -e

INPUT="./web/static/css/input.css"
OUTPUT="./web/static/css/output.css"

# Check if tailwindcss is installed
if ! command -v tailwindcss &> /dev/null; then
    echo "Error: tailwindcss CLI not found"
    echo "Install it with: make install-tailwind"
    echo "Or download from: https://github.com/tailwindlabs/tailwindcss/releases/latest"
    exit 1
fi

if [ "$1" = "--watch" ]; then
    echo "Starting Tailwind CSS in watch mode..."
    tailwindcss -i "$INPUT" -o "$OUTPUT" --watch
else
    echo "Building Tailwind CSS..."
    tailwindcss -i "$INPUT" -o "$OUTPUT" --minify
    echo "✓ CSS built successfully at $OUTPUT"
fi
