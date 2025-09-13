// BrowserStack configuration for cross-browser testing
module.exports = {
  webdriver: {
    timeout: 180000,
    host: 'hub.browserstack.com',
    port: 80
  },

  browsers: [
    {
      browser: 'Chrome',
      os: 'Windows',
      os_version: '10',
      browser_version: 'latest',
      name: 'Windows Chrome'
    },
    {
      browser: 'Firefox',
      os: 'Windows',
      os_version: '10',
      browser_version: 'latest',
      name: 'Windows Firefox'
    },
    {
      browser: 'Safari',
      os: 'OS X',
      os_version: 'Ventura',
      browser_version: '16.0',
      name: 'macOS Safari'
    },
    {
      browser: 'Edge',
      os: 'Windows',
      os_version: '10',
      browser_version: 'latest',
      name: 'Windows Edge'
    },
    {
      device: 'iPhone 12 Pro',
      os: 'ios',
      os_version: '14',
      real_mobile: true,
      name: 'iPhone 12 Safari'
    },
    {
      device: 'Samsung Galaxy S22',
      os: 'android',
      os_version: '12.0',
      real_mobile: true,
      name: 'Galaxy S22 Chrome'
    }
  ],

  test_settings: {
    default: {
      desiredCapabilities: {
        'browserstack.user': process.env.BROWSERSTACK_USERNAME,
        'browserstack.key': process.env.BROWSERSTACK_ACCESS_KEY,
        'browserstack.debug': true,
        'browserstack.console': 'verbose',
        'browserstack.networkLogs': true,
        'resolution': '1920x1080'
      }
    }
  },

  test_path: 'tests/browser/',
  output_folder: 'tests/compatibility/reports'
};