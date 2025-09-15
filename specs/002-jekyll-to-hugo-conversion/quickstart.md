# Quickstart: Hugo Migration Guide

## Installation
```bash
# Install Hugo extended version (for Sass support)
brew install hugo  # macOS
# or
choco install hugo-extended  # Windows
# or see https://gohugo.io/installation/

# Verify installation
hugo version
```

## Setup
```bash
# Navigate to docs directory
cd /Users/robson/Project/judo-cli/docs

# Initialize Hugo site (if starting fresh)
hugo new site . --force

# Install dependencies for asset processing
npm init -y
npm install -D @fullhuman/postcss-purgecss postcss-cli
```

## Content Migration
1. Convert Markdown files from Jekyll to Hugo front matter
2. Update permalinks in `config.toml` to match Jekyll structure
3. Set up content sections and collections
4. Migrate static assets to appropriate directories

## Development
```bash
# Start development server
hugo server -D

# Build for production
hugo --minify

# Test build output
hugo --buildDrafts --buildFuture
```

## Deployment
### GitHub Pages
```yaml
# .github/workflows/hugo.yml
name: Deploy Hugo site to Pages
on:
  push:
    branches: [master]
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - run: go install github.com/gohugoio/hugo@latest
      - run: hugo --minify
      - uses: peaceiris/actions-gh-pages@v3
```

### Netlify
```toml
# netlify.toml
[build]
  publish = "public"
  command = "hugo --gc --minify"

[build.environment]
  HUGO_VERSION = "0.150.0"
  HUGO_ENABLEGITINFO = "true"

[context.production.environment]
  HUGO_ENV = "production"
```

## Testing
```bash
# Run visual regression tests
npm run test:visual

# Execute browser tests
npm run test:browser

# Validate build output
npm run test:build
```

## Troubleshooting
- **Build errors**: Check Hugo version compatibility
- **Missing styles**: Verify Sass compilation configuration
- **Broken links**: Validate permalink structure
- **JavaScript issues**: Check asset pipeline configuration
- **Missing pages**: Ensure content files are in `/docs/content/` directory, not root `/docs/`
