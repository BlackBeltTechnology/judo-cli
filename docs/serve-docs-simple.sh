#!/bin/bash

# Simple JUDO CLI Documentation Server (No LiveReload)
# Compatible with older Ruby versions

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to show Ruby error message
show_ruby_error() {
    echo "Error: Failed to install Jekyll dependencies!"
    echo "This is likely due to Ruby environment issues."
    echo ""
    echo "The documentation is available online at: https://judo.technology/"
    echo ""
    echo "For local development, consider:"
    echo "  1. Installing a proper Ruby version manager (rbenv, rvm, or asdf)"
    echo "  2. Installing a newer Ruby version (3.0+)"
    echo "  3. Using the online documentation instead"
    exit 1
}

echo -e "${BLUE}Starting JUDO CLI Documentation Server...${NC}"

# Check if we're in the right directory
if [[ ! -f "_config.yml" ]]; then
    echo "Error: _config.yml not found. Run this script from the project root."
    exit 1
fi

# Clean old builds
echo "Cleaning previous builds..."
rm -rf _site .jekyll-cache

# Install dependencies if needed
if [[ -f "Gemfile" ]] && ! bundle check &> /dev/null; then
    echo "Installing dependencies..."
    # Try with system bundle first, fallback to rbenv if available
    if command -v ~/.rbenv/shims/bundle &> /dev/null; then
        if ! ~/.rbenv/shims/bundle install --path vendor/bundle 2>/dev/null; then
            show_ruby_error
        fi
    elif command -v bundle &> /dev/null; then
        if ! bundle install --path vendor/bundle 2>/dev/null; then
            show_ruby_error
        fi
    else
        show_ruby_error
    fi
fi

# Start Jekyll server
echo -e "${GREEN}Starting Jekyll server...${NC}"
echo "Documentation will be available at: http://127.0.0.1:4000/"
echo "Press Ctrl+C to stop"
echo

# Start without livereload for maximum compatibility
# Use rbenv bundle if available, otherwise fallback to system
if command -v ~/.rbenv/shims/bundle &> /dev/null && [[ -f "Gemfile" ]]; then
    ~/.rbenv/shims/bundle exec jekyll serve --host 127.0.0.1 --port 4000 --incremental --watch
elif command -v bundle &> /dev/null && [[ -f "Gemfile" ]]; then
    bundle exec jekyll serve --host 127.0.0.1 --port 4000 --incremental --watch
else
    jekyll serve --host 127.0.0.1 --port 4000 --incremental --watch
fi