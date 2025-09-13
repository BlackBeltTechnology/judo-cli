// Cross-browser compatibility test configuration
module.exports = {
  // Target browsers for testing
  browsers: [
    {
      name: 'Chrome',
      version: 'latest',
      platform: 'Windows 10',
      features: ['es6', 'css-variables', 'flexbox', 'grid']
    },
    {
      name: 'Firefox',
      version: 'latest',
      platform: 'Windows 10',
      features: ['es6', 'css-variables', 'flexbox', 'grid']
    },
    {
      name: 'Safari',
      version: '16',
      platform: 'macOS Ventura',
      features: ['es6', 'css-variables', 'flexbox', 'grid']
    },
    {
      name: 'Edge',
      version: 'latest',
      platform: 'Windows 10',
      features: ['es6', 'css-variables', 'flexbox', 'grid']
    },
    {
      name: 'Mobile Chrome',
      version: 'latest',
      platform: 'Android',
      features: ['es6', 'css-variables', 'flexbox', 'touch']
    },
    {
      name: 'Mobile Safari',
      version: '16',
      platform: 'iOS',
      features: ['es6', 'css-variables', 'flexbox', 'touch']
    }
  ],

  // Test scenarios for cross-browser compatibility
  testScenarios: [
    {
      name: 'Basic Rendering',
      description: 'Verify site renders without errors in all browsers',
      checks: [
        'No JavaScript errors in console',
        'No CSS parsing errors',
        'All images load correctly',
        'Fonts load correctly'
      ]
    },
    {
      name: 'Theme Functionality',
      description: 'Verify theme switching works across browsers',
      checks: [
        'Theme toggle button is functional',
        'Theme changes apply correctly',
        'Theme preference persists across page loads',
        'System theme detection works'
      ]
    },
    {
      name: 'Responsive Design',
      description: 'Verify responsive behavior across browsers and devices',
      checks: [
        'Layout adapts to different screen sizes',
        'Mobile navigation works correctly',
        'Touch interactions work on mobile devices',
        'No horizontal scrolling on mobile'
      ]
    },
    {
      name: 'Interactive Components',
      description: 'Verify all interactive elements work across browsers',
      checks: [
        'Install tabs switch content correctly',
        'Copy-to-clipboard functionality works',
        'Navigation menus work correctly',
        'Form elements are accessible'
      ]
    }
  ],

  // CSS Feature support matrix
  cssFeatures: {
    'css-variables': {
      required: true,
      browsers: {
        chrome: 49,
        firefox: 31,
        safari: 9.1,
        edge: 15,
        ios_saf: 9.3,
        android: 50
      }
    },
    'flexbox': {
      required: true,
      browsers: {
        chrome: 29,
        firefox: 28,
        safari: 9,
        edge: 12,
        ios_saf: 9,
        android: 4.4
      }
    }
  },

  // JavaScript Feature support matrix
  jsFeatures: {
    'es6': {
      required: true,
      browsers: {
        chrome: 51,
        firefox: 54,
        safari: 10,
        edge: 14,
        ios_saf: 10,
        android: 52
      }
    },
    'localStorage': {
      required: true,
      browsers: {
        chrome: 4,
        firefox: 3.5,
        safari: 4,
        edge: 12,
        ios_saf: 3.2,
        android: 2.1
      }
    }
  }
};