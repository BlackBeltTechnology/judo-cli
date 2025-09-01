#!/bin/bash

# JUDO CLI Documentation Development Server
# This script build jekyll site

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCS_DIR="$SCRIPT_DIR"
PORT=4000
HOST="127.0.0.1"
LIVERELOAD_PORT=35729

# Functions
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

# Function to show Ruby error message
show_ruby_error() {
    log_error "Failed to install Jekyll dependencies!"
    log_error "This is likely due to Ruby environment or native compilation issues."
    log_warning "The documentation is available online at: https://judo.technology/"
    log_info "For local development, you may need to:"
    log_info "  1. Install a proper Ruby version manager (rbenv, rvm, or asdf)"
    log_info "  2. Install a newer Ruby version (3.0+)"
    log_info "  3. Ensure Xcode command line tools are properly configured"
    log_info "  4. Or use the online documentation instead"
    exit 1
}

# Check if we're in the right directory
check_directory() {
    if [[ ! -f "$DOCS_DIR/_config.yml" ]]; then
        log_error "Jekyll configuration file (_config.yml) not found!"
        log_error "Make sure you're running this script from the project root directory."
        exit 1
    fi
    
    if [[ ! -f "$DOCS_DIR/Gemfile" ]]; then
        log_error "Gemfile not found!"
        log_error "Make sure Jekyll dependencies are set up properly."
        exit 1
    fi
}

# Check if Ruby and Bundler are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check for Ruby - try rbenv first, then system
    if command -v "$HOME/.rbenv/shims/ruby" &> /dev/null; then
        RUBY_CMD="$HOME/.rbenv/shims/ruby"
    elif command -v ruby &> /dev/null; then
        RUBY_CMD="ruby"
    else
        log_error "Ruby is not installed!"
        log_error "Please install Ruby 3.2+ to run Jekyll."
        log_error "Visit: https://www.ruby-lang.org/en/documentation/installation/"
        exit 1
    fi
    
    # Check for Bundler - try rbenv first, then system
    if command -v "$HOME/.rbenv/shims/bundle" &> /dev/null; then
        BUNDLE_CMD="$HOME/.rbenv/shims/bundle"
    elif command -v bundle &> /dev/null; then
        BUNDLE_CMD="bundle"
    else
        log_error "Bundler is not installed!"
        log_error "Install it with: gem install bundler"
        exit 1
    fi
    
    RUBY_VERSION=$($RUBY_CMD -v | cut -d' ' -f2 | cut -d'.' -f1-2)
    log_success "Ruby $RUBY_VERSION detected"
    log_success "Bundler $($BUNDLE_CMD -v | cut -d' ' -f3) detected"
    
    # Check Ruby version compatibility
    if [[ $(echo "$RUBY_VERSION < 2.6" | bc -l 2>/dev/null || echo 0) -eq 1 ]]; then
        log_warning "Ruby $RUBY_VERSION detected. Ruby 2.6+ recommended for best compatibility."
        log_info "Consider upgrading Ruby or some features may not work properly."
    fi
}

# Install Jekyll dependencies
install_dependencies() {
    log_info "Installing Jekyll dependencies..."
    
    cd "$DOCS_DIR"
    
    # Check for Bundler - try rbenv first, then system
    if command -v "$HOME/.rbenv/shims/bundle" &> /dev/null; then
        BUNDLE_CMD="$HOME/.rbenv/shims/bundle"
    elif command -v bundle &> /dev/null; then
        BUNDLE_CMD="bundle"
    else
        log_error "Bundler is not installed!"
        log_error "Install it with: gem install bundler"
        exit 1
    fi
    
    if ! $BUNDLE_CMD check &> /dev/null; then
        log_info "Installing gems..."
        if ! $BUNDLE_CMD install --path vendor/bundle 2>/dev/null; then
            show_ruby_error
        fi
        log_success "Dependencies installed successfully"
    else
        log_success "All dependencies are already installed"
    fi
}


# Clean Jekyll cache
clean_jekyll() {
    log_info "Cleaning Jekyll cache..."
    cd "$DOCS_DIR"
    
    if [[ -d "_site" ]]; then
        rm -rf _site
        log_success "Removed _site directory"
    fi
    
    if [[ -d ".jekyll-cache" ]]; then
        rm -rf .jekyll-cache
        log_success "Removed .jekyll-cache directory"
    fi
}

# Start Jekyll server
build_jekyll() {
    log_info "Build jekyll bundle"
    echo

    cd "$DOCS_DIR"

    # Use rbenv bundle if available, otherwise fallback to system
    if command -v "$HOME/.rbenv/shims/bundle" &> /dev/null; then
        "$HOME/.rbenv/shims/bundle" exec jekyll build --baseurl "" --trace 2>/dev/null
    else
        bundle exec jekyll build --baseurl "" --trace
    fi
}


# Parse command line arguments
CLEAN_CACHE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -c|--clean)
            CLEAN_CACHE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            log_info "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Main execution
main() {
    log_info "JUDO CLI Documentation Builder"
    echo
    
    # Run checks and setup
    check_directory
    check_dependencies
    install_dependencies

    # Clean cache if requested
    if [[ "$CLEAN_CACHE" == true ]]; then
        clean_jekyll
    fi

    # Build
    build_jekyll
}

# Run main function
main "$@"