#!/usr/bin/env node
// Script to run build validation tests
const { runBuildTests } = require('./build-validation');

console.log('Running build validation tests...');
const success = runBuildTests();

if (success) {
  console.log('✅ All build validation tests passed!');
  process.exit(0);
} else {
  console.log('❌ Some build validation tests failed');
  process.exit(1);
}