// Visual regression test configuration
// This would integrate with tools like Percy, BackstopJS, or Playwright

module.exports = {
  // Test scenarios for visual regression
  scenarios: [
    {
      label: 'Homepage - Light Theme',
      url: 'http://localhost:1313',
      referenceUrl: './baseline/homepage-light.png',
      viewports: [
        { width: 1920, height: 1080, label: 'desktop' },
        { width: 768, height: 1024, label: 'tablet' },
        { width: 375, height: 667, label: 'mobile' }
      ],
      onBeforeScript: 'theme-light.js',
      misMatchThreshold: 0.1
    },
    {
      label: 'Homepage - Dark Theme',
      url: 'http://localhost:1313',
      referenceUrl: './baseline/homepage-dark.png',
      viewports: [
        { width: 1920, height: 1080, label: 'desktop' },
        { width: 768, height: 1024, label: 'tablet' },
        { width: 375, height: 667, label: 'mobile' }
      ],
      onBeforeScript: 'theme-dark.js',
      misMatchThreshold: 0.1
    },
    {
      label: 'Commands Page',
      url: 'http://localhost:1313/commands/',
      referenceUrl: './baseline/commands.png',
      viewports: [
        { width: 1920, height: 1080, label: 'desktop' }
      ],
      misMatchThreshold: 0.1
    }
  ],
  
  // Paths configuration
  paths: {
    bitmaps_reference: 'tests/visual/baseline',
    bitmaps_test: 'tests/visual/results',
    engine: 'puppeteer',
    html_report: 'tests/visual/reports',
    ci_report: 'tests/visual/ci-reports'
  },
  
  // Engine configuration
  engine: 'puppeteer',
  engineOptions: {
    args: ['--no-sandbox']
  },
  
  // Report configuration
  report: ['browser', 'json'],
  debug: false
};