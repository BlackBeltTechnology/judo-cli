// Browser functional tests using Playwright/Puppeteer pattern

const testScenarios = [
  {
    name: 'Theme Toggle Functionality',
    steps: [
      'Navigate to homepage',
      'Verify initial theme matches system preference',
      'Click theme toggle',
      'Verify theme changes and persists',
      'Reload page and verify theme persistence'
    ],
    expected: 'Theme toggle should work correctly and persist across page loads'
  },
  {
    name: 'Navigation Menu',
    steps: [
      'Navigate to homepage',
      'Verify all menu items are present',
      'Click each menu item and verify navigation',
      'Test mobile hamburger menu on small screens'
    ],
    expected: 'Navigation should work correctly on all screen sizes'
  },
  {
    name: 'Install Tabs Component',
    steps: [
      'Navigate to installation page',
      'Verify install tabs are present',
      'Click each OS tab and verify content changes',
      'Test copy-to-clipboard functionality'
    ],
    expected: 'Install tabs should switch content and copy functionality should work'
  },
  {
    name: 'Responsive Design',
    steps: [
      'Test homepage on desktop (1920x1080)',
      'Test homepage on tablet (768x1024)',
      'Test homepage on mobile (375x667)',
      'Verify layout adapts correctly for each breakpoint'
    ],
    expected: 'Site should be fully responsive across all screen sizes'
  }
];

module.exports = {
  baseUrl: 'http://localhost:1313',
  viewports: [
    { width: 1920, height: 1080, name: 'desktop' },
    { width: 768, height: 1024, name: 'tablet' },
    { width: 375, height: 667, name: 'mobile' }
  ],
  scenarios: testScenarios,
  timeout: 30000,
  headless: true
};