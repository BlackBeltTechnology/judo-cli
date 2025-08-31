# JUDO CLI Documentation

This directory contains the documentation for JUDO CLI, built with Jekyll and the custom JUDO theme.

## Documentation Structure

```
docs/
├── README.md           # This file
├── _config.yml         # Jekyll configuration
├── index.md           # Documentation home page
├── glossary.md        # Command glossary
├── Gemfile            # Ruby dependencies
├── _docs/             # Documentation pages
│   ├── getting-started.md
│   ├── commands.md
│   ├── configuration.md
│   ├── examples.md
│   └── api.md
└── .github/
    └── workflows/
        └── docs.yml   # GitHub Pages deployment
```

## Local Development

### Prerequisites

- Ruby 3.2+
- Bundler

### Setup

```bash
# Install dependencies
bundle install

# Serve locally with live reload
bundle exec jekyll serve --livereload

# Build static site
bundle exec jekyll build
```

The site will be available at `http://localhost:4000/judo-cli/`

### Theme Configuration

The documentation uses the custom JUDO theme from:
`BlackBeltTechnology/jekyll-theme-judo@develop`

Theme features:
- Responsive design
- Dark/light mode toggle
- Search functionality
- Navigation sidebar
- Code syntax highlighting
- Mobile-friendly layout

## GitHub Pages Deployment

Documentation is automatically deployed to GitHub Pages when:

1. Changes are pushed to `develop` or `main` branches
2. Files in the documentation paths are modified:
   - `docs/**`
   - `_docs/**` 
   - `_config.yml`
   - `index.md`

The deployment workflow:
1. Builds the Jekyll site
2. Deploys to GitHub Pages
3. Available at: `https://blackbelttechnology.github.io/judo-cli/`

## Content Guidelines

### Writing Style

- Use clear, concise language
- Include practical examples
- Structure content with proper headings
- Use code blocks for commands and configuration
- Include navigation between related sections

### Documentation Sections

#### Getting Started
- Installation instructions for all platforms
- Basic usage and first steps
- Prerequisites and setup

#### Commands Reference
- Complete command documentation
- All options and flags
- Usage examples for each command

#### Configuration
- Configuration file formats
- Environment-specific settings
- Property reference tables

#### Examples
- Real-world usage patterns
- Common workflows
- Troubleshooting scenarios

#### API Reference
- Technical specifications
- Configuration schemas
- Error codes and responses

### Markdown Conventions

```markdown
# Page Title (H1 - once per page)

## Section Title (H2)

### Subsection (H3)

#### Details (H4)

```bash
# Code blocks with language
judo build start
```

**Bold text** for emphasis
*Italic text* for terms
`inline code` for commands/properties

> **Note:** Use blockquotes for important notes

| Column 1 | Column 2 |
|----------|----------|
| Value 1  | Value 2  |
```

### Front Matter

Each documentation page should include front matter:

```yaml
---
title: "Page Title"
permalink: /docs/page-name/
excerpt: "Brief description for SEO and navigation"
---
```

## Theme Customization

### Navigation

Update navigation in `_config.yml`:

```yaml
navigation:
  - title: "Home"
    url: "/"
  - title: "Getting Started"
    url: "/docs/getting-started/"
  # Add more items...
```

### Site Configuration

Key configuration options in `_config.yml`:

```yaml
title: "JUDO CLI"
description: "Command-line tool for managing JUDO applications"
baseurl: "/judo-cli"
url: "https://blackbelttechnology.github.io"

remote_theme: BlackBeltTechnology/jekyll-theme-judo@develop
```

## Contributing to Documentation

1. **Fork the repository**
2. **Create a feature branch** for documentation changes
3. **Make your changes** following the content guidelines
4. **Test locally** with `bundle exec jekyll serve`
5. **Submit a pull request** with clear description of changes

### Common Tasks

**Add a new documentation page:**

1. Create new file in `_docs/` directory
2. Add proper front matter
3. Update navigation in `_config.yml` if needed
4. Cross-reference from other relevant pages

**Update existing content:**

1. Edit the relevant `.md` file
2. Test locally to ensure formatting is correct
3. Update any cross-references if needed

**Add new sections:**

1. Plan the information architecture
2. Create the necessary files
3. Update navigation and cross-references
4. Consider adding examples and use cases

## Troubleshooting

### Common Issues

**Bundle install fails:**
```bash
# Update Ruby and Bundler
gem update bundler
bundle install
```

**Jekyll serve fails:**
```bash
# Clear cache and reinstall
rm -rf _site .jekyll-cache
bundle clean --force
bundle install
```

**Theme not loading:**
```bash
# Check theme configuration in _config.yml
# Ensure internet connection for remote theme
bundle exec jekyll clean
bundle exec jekyll serve
```

**GitHub Pages build fails:**
- Check the Actions tab for build errors
- Verify all files have valid front matter
- Ensure no invalid Liquid syntax

### Local Development Tips

- Use `--livereload` for automatic browser refresh
- Use `--drafts` to include draft posts
- Use `--incremental` for faster builds during development
- Check `_site/` directory to see generated output

## Resources

- [Jekyll Documentation](https://jekyllrb.com/docs/)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Markdown Guide](https://www.markdownguide.org/)
- [JUDO Theme Repository](https://github.com/BlackBeltTechnology/jekyll-theme-judo)