# Design System Contract

## Color Scheme
**Primary Accent**: #eb5a29 (orange-red)
**Primary Hover**: #d44a1f (darker orange-red)

### Light Theme
- Background: #ffffff
- Foreground: #111827
- Muted: #6b7280
- Border: #e5e7eb
- Surface: #f9fafb
- Hero Background: #fff8f0

### Dark Theme
- Background: #111827
- Foreground: #f9fafb
- Muted: #9ca3af
- Border: #374151
- Surface: #1f2937
- Hero Background: #1f2937

## Typography
**Sans Font**: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif
**Mono Font**: "JetBrains Mono", "Fira Code", "SFMono-Regular", "Roboto Mono", Consolas, "Liberation Mono", Menlo, monospace

### Type Scale
- H1: clamp(2.5rem, 5vw + 1rem, 3rem)
- H2: clamp(2rem, 3vw + 1rem, 2rem)
- H3: clamp(1.2rem, 2vw + 0.5rem, 1.6rem)
- H4: 1.25rem
- Body: 1.125rem
- Small: 0.875rem
- XS: 0.75rem

## Spacing
- 1: 0.25rem
- 2: 0.5rem
- 3: 0.75rem
- 4: 1rem
- 5: 1.25rem
- 6: 1.5rem
- 8: 2rem
- 12: 3rem
- 16: 4rem
- 20: 5rem

## Components
### Buttons
- Primary: accent background, white text, hover elevation
- Secondary: transparent, border, hover background

### Cards
- Background: surface color
- Border: border color
- Shadow: subtle elevation
- Hover: elevation increase and border accent

### Navigation
- Sticky header with backdrop blur
- Mobile hamburger menu
- Theme toggle switch

## Responsive Breakpoints
- Mobile: < 768px
- Tablet: 768px - 1024px
- Desktop: > 1024px

## Animation
- Theme transition: 0.3s ease
- Button hover: 0.2s ease
- Card hover: 0.3s ease
- Hero elements: floating animation
