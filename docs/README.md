# JUDO CLI Documentation

This directory contains the documentation for JUDO CLI, built with Hugo static site generator.

## Documentation Structure

```
docs/
├── README.md           # This file
├── hugo.toml          # Hugo configuration
├── content/           # Documentation content
│   ├── _index.md      # Home page
│   ├── commands/      # Command documentation
│   ├── configuration/ # Configuration guides
│   ├── getting-started/ # Getting started guides
│   ├── examples/      # Usage examples
│   └── api/           # API reference
├── layouts/           # HTML templates
│   ├── _default/      # Default layouts
│   └── partials/      # Reusable components
├── assets/            # CSS, JS, images
│   ├── css/
│   ├── js/
│   └── img/
├── static/            # Static files
└── .github/
    └── workflows/
        └── hugo.yml   # GitHub Pages deployment
```

## Local Development

### Prerequisites

- Hugo Extended v0.150.0+

### Setup

```bash
# Serve locally with live reload
hugo server

# Build static site
hugo --minify
```

The site will be available at `http://localhost:1313/`

### Theme Features

- Responsive design
- Dark/light mode toggle
- Navigation menu
- Code syntax highlighting
- Mobile-friendly layout
- Fast build times

## GitHub Pages Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to the main branch.

The deployment workflow:
1. Builds the Hugo site
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
description: "Brief description for SEO and navigation"
---
```

## Theme Customization

### Navigation

Update navigation in `hugo.toml`:

```toml
[menu]
  [[menu.main]]
    name = "Home"
    url = "/"
    weight = 100
  [[menu.main]]
    name = "Getting Started"
    url = "/getting-started/"
    weight = 200
```

### Site Configuration

Key configuration options in `hugo.toml`:

```toml
baseURL = 'https://judo.technology/'
title = 'JUDO CLI'

[params]
  description = 'Command-line tool for managing JUDO applications'
  author = 'BlackBelt Technology'
```

## Contributing to Documentation

1. **Fork the repository**
2. **Create a feature branch** for documentation changes
3. **Make your changes** following the content guidelines
4. **Test locally** with `hugo server`
5. **Submit a pull request** with clear description of changes

### Common Tasks

**Add a new documentation page:**

1. Create new file in `content/` directory
2. Add proper front matter
3. Update navigation in `hugo.toml` if needed
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

## Local Development Tips

- Use `hugo server` for automatic browser refresh
- Use `hugo --minify` for production builds
- Check `public/` directory to see generated output

## Resources

- [Hugo Documentation](https://gohugo.io/documentation/)
- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [Markdown Guide](https://www.markdownguide.org/)