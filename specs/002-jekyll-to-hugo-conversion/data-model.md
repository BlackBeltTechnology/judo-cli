# Data Model: Jekyll to Hugo Content Structure

## Content Entities

### Documentation Page
**Purpose**: Individual documentation content with metadata and content
**Attributes**:
- title: Page title (string)
- layout: Template assignment (string)
- permalink: URL structure (string)
- description: Meta description (string)
- nav_order: Navigation ordering (number)
- has_children: Section hierarchy (boolean)

### Navigation Structure
**Purpose**: Site menu configuration and hierarchy
**Attributes**:
- nav: Array of menu items with title and url
- sections: Content organization with ordering
- collections: Grouped content types (commands, configuration)

### Theme Configuration
**Purpose**: Design system and visual appearance
**Attributes**:
- color_scheme: Light/dark theme settings
- typography: Font families and scales
- spacing: Design token values
- breakpoints: Responsive design thresholds

### Interactive Component
**Purpose**: JavaScript-enhanced UI elements
**Attributes**:
- type: Component category (theme-toggle, install-tabs, copy-button)
- behavior: Interactive functionality specification
- dependencies: JavaScript and CSS requirements
- accessibility: ARIA labels and keyboard navigation

## Content Relationships
- Navigation → Pages: One-to-many (menu contains multiple pages)
- Theme → Components: One-to-many (theme styles multiple components)
- Page → Layout: One-to-one (each page uses one layout template)

## Migration Mapping
- Jekyll front matter → Hugo front matter (direct conversion)
- Jekyll layouts → Hugo templates (syntax conversion)
- Jekyll includes → Hugo partials (syntax conversion)
- Jekyll Sass → Hugo asset pipeline (processing conversion)
- Jekyll plugins → Hugo built-in features (functional replacement)
