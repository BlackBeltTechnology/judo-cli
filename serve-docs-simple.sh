#!/bin/bash

# Simple JUDO CLI Documentation Server (No LiveReload)
# Compatible with older Ruby versions

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    bundle install
fi

# Start Jekyll server
echo -e "${GREEN}Starting Jekyll server...${NC}"
echo "Documentation will be available at: http://127.0.0.1:4000/judo-cli/"
echo "Press Ctrl+C to stop"
echo

# Start without livereload for maximum compatibility
if command -v bundle &> /dev/null && [[ -f "Gemfile" ]]; then
    bundle exec jekyll serve --host 127.0.0.1 --port 4000 --incremental --watch
else
    jekyll serve --host 127.0.0.1 --port 4000 --incremental --watch
fi