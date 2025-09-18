#!/bin/bash

# Build script for judo-cli - compiles frontend and backend with embedded frontend assets

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Building judo-cli ===${NC}"

# Get the project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
BACKEND_DIR="$PROJECT_ROOT"

# Check if we're in the right directory
if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
    echo -e "${RED}Error: Not in the judo-cli project root directory${NC}"
    exit 1
fi

echo -e "${YELLOW}Building frontend React app...${NC}"

# Build frontend
cd "$FRONTEND_DIR"
if [ ! -f "package.json" ]; then
    echo -e "${RED}Error: Frontend directory not found at $FRONTEND_DIR${NC}"
    exit 1
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    npm install
fi

# Build the React app
npm run build

# Check if build was successful
if [ ! -d "build" ]; then
    echo -e "${RED}Error: Frontend build failed - no build directory created${NC}"
    exit 1
fi

echo -e "${GREEN}Frontend built successfully!${NC}"

# Copy build assets to embedded assets directory for Go embedding
echo -e "${YELLOW}Preparing embedded assets...${NC}"
cd "$PROJECT_ROOT"

# Create assets directory if it doesn't exist
mkdir -p internal/server/assets

# Copy frontend build to assets directory
rm -rf internal/server/assets/*
cp -r frontend/build/* internal/server/assets/

echo -e "${GREEN}Assets prepared for embedding!${NC}"

echo -e "${YELLOW}Building Go backend with embedded frontend...${NC}"

# Build the Go binary
cd "$BACKEND_DIR"

# Get version information
VERSION=$(scripts/version.sh get 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Build the Go binary with embedded frontend assets
echo -e "${YELLOW}Building Go binary with version: $VERSION, commit: $COMMIT${NC}"

go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE -X main.builtBy=build.sh" -o judo ./cmd/judo

# Check if build was successful
if [ ! -f "judo" ]; then
    echo -e "${RED}Error: Go build failed - no binary created${NC}"
    exit 1
fi

# Make the binary executable
chmod +x judo

echo -e "${GREEN}Backend built successfully!${NC}"

# Run tests to ensure everything works
echo -e "${YELLOW}Running tests...${NC}"
if go test ./... -cover; then
    echo -e "${GREEN}All tests passed!${NC}"
else
    echo -e "${RED}Tests failed!${NC}"
    exit 1
fi

echo -e "${BLUE}=== Build complete! ===${NC}"
echo -e "${GREEN}Binary created: $(pwd)/judo${NC}"
echo -e "${GREEN}Version: $VERSION${NC}"
echo -e "${GREEN}Commit: $COMMIT${NC}"
echo -e "${GREEN}Build date: $DATE${NC}"