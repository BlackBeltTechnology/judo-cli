# Tasks: Convert Jekyll Documentation to Hugo

**Input**: Design documents from `/specs/002-jekyll-to-hugo-conversion/`
**Prerequisites**: Implementation plan and research findings

## Phase 3.1: Setup
- [x] T001 [P] Install Hugo CLI and verify version compatibility in `/docs/`
- [x] T002 [P] Initialize Hugo site structure in `/docs/` with `hugo new site . --force`
- [x] T003 [P] Configure Hugo base settings in `/docs/config.toml` (title, baseURL, theme)
- [x] T004 [P] Set up asset pipeline configuration in `/docs/assets/` for Sass and JS
- [x] T005 [P] Create build scripts in `/docs/package.json` for development and production

## Phase 3.2: Testing & Validation Setup ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These validation steps MUST be established before migration**
- [x] T006 [P] Create visual regression test baseline from current Jekyll site in `/tests/visual/baseline/`
- [x] T007 [P] Set up browser testing configuration in `/tests/browser/` for functional testing
- [x] T008 [P] Create build validation tests in `/tests/build/` for Hugo compilation
- [x] T009 [P] Establish cross-browser testing matrix in `/tests/compatibility/`

## Phase 3.3: Content Migration
- [x] T010 [P] Convert all Markdown files from Jekyll to Hugo front matter in `/docs/content/`
- [x] T011 [P] Migrate navigation structure and menu configuration in `/docs/config.toml`
- [x] T012 [P] Set up content sections and collections in `/docs/content/` (commands, configuration, etc.)
- [x] T013 [P] Configure permalinks and URL structure in `/docs/config.toml` to match Jekyll
- [x] T014 [P] Migrate static assets (images, fonts) to `/docs/static/` preserving paths
- [x] T014a [P] Fix missing API/examples/session pages by moving from `/docs/` to `/docs/content/`

## Phase 3.4: Layout & Template Conversion
- [x] T015 [P] Convert Jekyll layouts to Hugo templates in `/docs/layouts/_default/`
- [x] T016 [P] Recreate header partial with theme toggle in `/docs/layouts/partials/header.html`
- [x] T017 [P] Recreate footer partial in `/docs/layouts/partials/footer.html`
- [x] T018 [P] Convert home layout with hero section in `/docs/layouts/_default/home.html`
- [x] T019 [P] Create page layout for documentation content in `/docs/layouts/_default/single.html`
- [x] T020 [P] Implement install tabs component as Hugo shortcode in `/docs/layouts/shortcodes/install-tabs.html`

## Phase 3.5: Design System Migration
- [x] T021 [P] Convert Sass variables to CSS custom properties in `/docs/assets/css/_variables.scss`
- [x] T022 [P] Migrate base styles and typography in `/docs/assets/css/_base.scss`
- [x] T023 [P] Recreate component styles in `/docs/assets/css/_components.scss`
- [x] T024 [P] Implement theme switching system in `/docs/assets/css/_theme.scss`
- [x] T025 [P] Set up Sass compilation pipeline in `/docs/assets/config.yml`

## Phase 3.6: Interactive Components
- [x] T026 [P] Migrate theme toggle JavaScript functionality in `/docs/assets/js/theme.js`
- [x] T027 [P] Recreate install tabs interactive behavior in `/docs/assets/js/install-tabs.js`
- [x] T028 [P] Implement copy-to-clipboard functionality in `/docs/assets/js/copy.js`
- [x] T029 [P] Add mobile navigation toggle in `/docs/assets/js/navigation.js`
- [x] T030 [P] Set up JavaScript bundling and optimization in `/docs/assets/config.yml`

## Phase 3.7: Build & Deployment
- [x] T031 [P] Configure Hugo build settings for production in `/docs/config.toml`
- [x] T032 [P] Set up GitHub Actions workflow for Hugo deployment in `.github/workflows/hugo.yml`
- [x] T033 [P] Create Netlify configuration in `/docs/netlify.toml` for alternative deployment
- [x] T034 [P] Configure build caching and incremental builds in `/docs/config.toml`
- [x] T035 [P] Set up asset fingerprinting and CDN configuration in `/docs/config.toml`

## Phase 3.8: Testing & Validation
- [x] T036 [P] Run visual regression tests against baseline in `/tests/visual/`
- [x] T037 [P] Execute functional browser tests for all interactive components in `/tests/browser/`
- [x] T038 [P] Validate build output and file structure in `/tests/build/`
- [x] T039 [P] Test cross-browser compatibility in `/tests/compatibility/`
- [x] T040 [P] Verify all URLs and permalinks work correctly in `/tests/urls/`

## Phase 3.9: Polish & Optimization
- [x] T041 [P] Optimize CSS and JavaScript bundle sizes in `/docs/assets/`
- [x] T042 [P] Implement lazy loading for images and components in `/docs/layouts/partials/`
- [x] T043 [P] Add performance monitoring and analytics in `/docs/layouts/partials/head.html`
- [x] T044 [P] Create 404 error page in `/docs/layouts/404.html`
- [x] T045 [P] Set up RSS feed and sitemap generation in `/docs/config.toml`

## Dependencies
- Setup (T001-T005) before all other phases
- Testing setup (T006-T009) before content migration (T010-T014)
- Content migration (T010-T014) before layout conversion (T015-T020)
- Layout conversion (T015-T020) before design system (T021-T025)
- Design system (T021-T025) before interactive components (T026-T030)
- All implementation before build/deployment (T031-T035)
- Build/deployment before testing/validation (T036-T040)
- Testing/validation before polish (T041-T045)

## Parallel Execution Examples
```
# Launch content migration tasks together:
Task: "Convert all Markdown files from Jekyll to Hugo front matter in /docs/content/"
Task: "Migrate navigation structure and menu configuration in /docs/config.toml"
Task: "Set up content sections and collections in /docs/content/"
Task: "Configure permalinks and URL structure in /docs/config.toml"

# Launch layout conversion tasks together:
Task: "Convert Jekyll layouts to Hugo templates in /docs/layouts/_default/"
Task: "Recreate header partial with theme toggle in /docs/layouts/partials/header.html"
Task: "Recreate footer partial in /docs/layouts/partials/footer.html"
Task: "Convert home layout with hero section in /docs/layouts/_default/home.html"
```

## Notes
- [P] tasks can run in parallel (different files, no dependencies)
- Follow TDD approach: testing setup before implementation
- Commit after each major phase completion
- Validate visual design consistency at each stage
- Maintain identical user experience throughout migration
