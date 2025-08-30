#!/bin/bash

# JUDO CLI Documentation Development Server
# This script starts Jekyll with livereload for local documentation development

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
    
    if ! command -v ruby &> /dev/null; then
        log_error "Ruby is not installed!"
        log_error "Please install Ruby 3.2+ to run Jekyll."
        log_error "Visit: https://www.ruby-lang.org/en/documentation/installation/"
        exit 1
    fi
    
    if ! command -v bundle &> /dev/null; then
        log_error "Bundler is not installed!"
        log_error "Install it with: gem install bundler"
        exit 1
    fi
    
    RUBY_VERSION=$(ruby -v | cut -d' ' -f2 | cut -d'.' -f1-2)
    log_success "Ruby $RUBY_VERSION detected"
    log_success "Bundler $(bundle -v | cut -d' ' -f3) detected"
    
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
    
    if ! bundle check &> /dev/null; then
        log_info "Installing gems..."
        if ! bundle install --path vendor/bundle 2>/dev/null; then
            log_error "Failed to install Jekyll dependencies!"
            log_error "This is likely due to Ruby environment or native compilation issues."
            log_warning "The documentation is available online at: https://judo.technology/"
            log_info "For local development, you may need to:"
            log_info "  1. Install a proper Ruby version manager (rbenv, rvm, or asdf)"
            log_info "  2. Install a newer Ruby version (3.0+)"
            log_info "  3. Ensure Xcode command line tools are properly configured"
            log_info "  4. Or use the online documentation instead"
            exit 1
        fi
        log_success "Dependencies installed successfully"
    else
        log_success "All dependencies are already installed"
    fi
}

# Check if ports are available
check_ports() {
    log_info "Checking if ports are available..."
    
    if lsof -i :$PORT &> /dev/null; then
        log_warning "Port $PORT is already in use!"
        log_info "Trying to find an available port..."
        
        # Find next available port
        for ((port=$PORT; port<=4010; port++)); do
            if ! lsof -i :$port &> /dev/null; then
                PORT=$port
                log_info "Using port $port instead"
                break
            fi
        done
    fi
    
    if lsof -i :$LIVERELOAD_PORT &> /dev/null; then
        log_warning "LiveReload port $LIVERELOAD_PORT is already in use!"
        log_info "LiveReload may not work properly"
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
start_jekyll() {
    log_info "Starting Jekyll development server..."
    log_info "Site will be available at: http://$HOST:$PORT/judo-cli/"
    log_info "LiveReload will watch for changes automatically"
    log_info "Press Ctrl+C to stop the server"
    echo
    
    cd "$DOCS_DIR"
    
    # Start Jekyll with livereload (using built-in feature)
    bundle exec jekyll serve \
        --host "$HOST" \
        --port "$PORT" \
        --livereload \
        --livereload-port "$LIVERELOAD_PORT" \
        --incremental \
        --watch \
        --open-url 2>/dev/null || \
    bundle exec jekyll serve \
        --host "$HOST" \
        --port "$PORT" \
        --incremental \
        --watch
}

# Cleanup on exit
cleanup() {
    echo
    log_info "Shutting down Jekyll server..."
    log_success "Development server stopped"
}

# Show help
show_help() {
    cat << EOF
JUDO CLI Documentation Development Server

Usage: $0 [OPTIONS]

Options:
    -h, --help          Show this help message
    -c, --clean         Clean Jekyll cache before starting
    -p, --port PORT     Use specific port (default: 4000)
    --host HOST         Use specific host (default: 127.0.0.1)
    --no-open          Don't automatically open browser

Examples:
    $0                  Start development server with defaults
    $0 --clean          Clean cache and start server
    $0 --port 3000      Start server on port 3000
    $0 --host 0.0.0.0   Start server accessible from network

The documentation will be available at:
    http://HOST:PORT/judo-cli/

EOF
}

# Parse command line arguments
CLEAN_CACHE=false
OPEN_BROWSER=true

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
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        --host)
            HOST="$2"
            shift 2
            ;;
        --no-open)
            OPEN_BROWSER=false
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
    log_info "JUDO CLI Documentation Development Server"
    echo
    
    # Setup trap for cleanup
    trap cleanup EXIT INT TERM
    
    # Run checks and setup
    check_directory
    check_dependencies
    install_dependencies
    check_ports
    
    # Clean cache if requested
    if [[ "$CLEAN_CACHE" == true ]]; then
        clean_jekyll
    fi
    
    # Modify Jekyll args based on options
    JEKYLL_ARGS="--host $HOST --port $PORT --livereload --livereload-port $LIVERELOAD_PORT --incremental --watch"
    
    if [[ "$OPEN_BROWSER" == true ]]; then
        JEKYLL_ARGS="$JEKYLL_ARGS --open-url"
    fi
    
    # Start the server
    start_jekyll
}

# Run main function
main "$@"