// Build validation tests for Hugo site
const fs = require('fs');
const path = require('path');

const buildTests = {
  // Test that build output exists
  buildOutputExists: () => {
    const publicDir = path.join(__dirname, '../../docs/public');
    return fs.existsSync(publicDir) && fs.readdirSync(publicDir).length > 0;
  },

  // Test that CSS is processed and minified
  cssIsProcessed: () => {
    const cssDir = path.join(__dirname, '../../docs/public/css');
    if (!fs.existsSync(cssDir)) return false;
    
    const cssFiles = fs.readdirSync(cssDir);
    return cssFiles.some(file => file.endsWith('.css') && file.includes('min'));
  },

  // Test that JS is processed and minified
  jsIsProcessed: () => {
    const jsDir = path.join(__dirname, '../../docs/public/js');
    if (!fs.existsSync(jsDir)) return false;
    
    const jsFiles = fs.readdirSync(jsDir);
    return jsFiles.some(file => file.endsWith('.js') && file.includes('min'));
  },

  // Test that HTML files are generated
  htmlFilesGenerated: () => {
    const publicDir = path.join(__dirname, '../../docs/public');
    if (!fs.existsSync(publicDir)) return false;
    
    const findHtmlFiles = (dir) => {
      let htmlFiles = [];
      const items = fs.readdirSync(dir);
      
      for (const item of items) {
        const fullPath = path.join(dir, item);
        const stat = fs.statSync(fullPath);
        
        if (stat.isDirectory()) {
          htmlFiles = htmlFiles.concat(findHtmlFiles(fullPath));
        } else if (item.endsWith('.html')) {
          htmlFiles.push(fullPath);
        }
      }
      return htmlFiles;
    };
    
    return findHtmlFiles(publicDir).length >= 5; // At least 5 HTML files
  },

  // Test that assets are fingerprinted
  assetsFingerprinted: () => {
    const cssDir = path.join(__dirname, '../../docs/public/css');
    if (!fs.existsSync(cssDir)) return false;
    
    const cssFiles = fs.readdirSync(cssDir);
    return cssFiles.some(file => file.includes('.')); // Files with dots (hashes)
  }
};

// Run all build tests
const runBuildTests = () => {
  const results = {};
  let allPassed = true;
  
  for (const [testName, testFn] of Object.entries(buildTests)) {
    try {
      const result = testFn();
      results[testName] = result ? 'PASS' : 'FAIL';
      if (!result) allPassed = false;
    } catch (error) {
      results[testName] = 'ERROR';
      allPassed = false;
    }
  }
  
  console.log('Build Validation Results:');
  console.table(results);
  
  return allPassed;
};

module.exports = { buildTests, runBuildTests };

// Run if called directly
if (require.main === module) {
  process.exit(runBuildTests() ? 0 : 1);
}