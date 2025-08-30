#!/bin/bash

set -e

# Version management script for JUDO CLI
# This script handles version increment and management for releases

VERSION_FILE="VERSION"
CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$CURRENT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if VERSION file exists, create if not
ensure_version_file() {
    if [[ ! -f "$PROJECT_ROOT/$VERSION_FILE" ]]; then
        log_warning "VERSION file not found, creating with initial version 0.1.0"
        echo "0.1.0" > "$PROJECT_ROOT/$VERSION_FILE"
    fi
}

# Get current version from VERSION file
get_current_version() {
    ensure_version_file
    cat "$PROJECT_ROOT/$VERSION_FILE" | tr -d '[:space:]'
}

# Increment version number
increment_version() {
    local version="$1"
    local part="$2" # major, minor, patch
    
    IFS='.' read -ra VERSION_PARTS <<< "$version"
    local major="${VERSION_PARTS[0]}"
    local minor="${VERSION_PARTS[1]}"
    local patch="${VERSION_PARTS[2]}"
    
    case "$part" in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch"|*)
            patch=$((patch + 1))
            ;;
    esac
    
    echo "${major}.${minor}.${patch}"
}

# Generate snapshot version with timestamp
generate_snapshot_version() {
    local base_version="$1"
    local timestamp=$(date +%Y%m%d%H%M%S)
    echo "${base_version}-snapshot-${timestamp}"
}

# Update VERSION file
update_version_file() {
    local new_version="$1"
    echo "$new_version" > "$PROJECT_ROOT/$VERSION_FILE"
    log_success "Updated VERSION file to: $new_version"
}

# Commit version change
commit_version_change() {
    local new_version="$1"
    local message="chore: bump version to $new_version"
    
    cd "$PROJECT_ROOT"
    git add "$VERSION_FILE"
    git commit -m "$message" || {
        log_warning "No changes to commit (version file unchanged)"
        return 0
    }
    log_success "Committed version change: $message"
}

# Get version for current context (tag vs branch)
get_build_version() {
    local current_version
    current_version=$(get_current_version)
    
    # Check if we're on a tag
    if git describe --exact-match --tags HEAD >/dev/null 2>&1; then
        local tag_name
        tag_name=$(git describe --exact-match --tags HEAD)
        log_info "Building from tag: $tag_name"
        echo "$tag_name" | sed 's/^v//'
    else
        # Check current branch
        local current_branch
        current_branch=$(git rev-parse --abbrev-ref HEAD)
        
        if [[ "$current_branch" == "develop" ]]; then
            local snapshot_version
            snapshot_version=$(generate_snapshot_version "$current_version")
            log_info "Building snapshot version: $snapshot_version"
            echo "$snapshot_version"
        else
            log_info "Building from branch $current_branch with version: $current_version"
            echo "$current_version"
        fi
    fi
}

# Main script logic
main() {
    local command="${1:-get}"
    
    case "$command" in
        "get")
            get_current_version
            ;;
        "build")
            get_build_version
            ;;
        "increment")
            local part="${2:-patch}"
            local current_version
            current_version=$(get_current_version)
            local new_version
            new_version=$(increment_version "$current_version" "$part")
            
            log_info "Incrementing $part version: $current_version -> $new_version"
            update_version_file "$new_version"
            
            if [[ "${3:-}" == "--commit" ]]; then
                commit_version_change "$new_version"
            fi
            
            echo "$new_version"
            ;;
        "set")
            local new_version="$2"
            if [[ -z "$new_version" ]]; then
                log_error "Version number required for 'set' command"
                exit 1
            fi
            
            log_info "Setting version to: $new_version"
            update_version_file "$new_version"
            
            if [[ "${3:-}" == "--commit" ]]; then
                commit_version_change "$new_version"
            fi
            
            echo "$new_version"
            ;;
        "snapshot")
            local current_version
            current_version=$(get_current_version)
            generate_snapshot_version "$current_version"
            ;;
        "help"|*)
            echo "Usage: $0 <command> [options]"
            echo ""
            echo "Commands:"
            echo "  get                     Get current version"
            echo "  build                   Get version for build (handles tags/branches)"
            echo "  increment [part]        Increment version (major|minor|patch, default: patch)"
            echo "  set <version>           Set specific version"
            echo "  snapshot                Generate snapshot version"
            echo "  help                    Show this help"
            echo ""
            echo "Options:"
            echo "  --commit                Commit version changes to git"
            echo ""
            echo "Examples:"
            echo "  $0 get                  # Get current version"
            echo "  $0 build                # Get build version (with snapshot for develop)"
            echo "  $0 increment patch      # Increment patch version"
            echo "  $0 increment minor --commit  # Increment minor and commit"
            echo "  $0 set 1.2.3 --commit  # Set version and commit"
            ;;
    esac
}

# Run main function with all arguments
main "$@"