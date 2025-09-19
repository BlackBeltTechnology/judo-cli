// This script helps check browser console for terminal errors
console.log('🔍 Checking browser console for terminal issues...');

// Check if xterm.js is loaded
if (typeof window.Terminal === 'function') {
  console.log('✅ xterm.js is loaded');
} else {
  console.error('❌ xterm.js not found');
}

// Check if fit addon is available
if (typeof window.FitAddon === 'function') {
  console.log('✅ FitAddon is available');
} else {
  console.error('❌ FitAddon not found');
}

// Check if WebSocket is working
if (typeof window.WebSocket === 'function') {
  console.log('✅ WebSocket API available');
} else {
  console.error('❌ WebSocket API not available');
}

// Simulate terminal initialization to check for errors
try {
  const term = new window.Terminal();
  const fitAddon = new window.FitAddon();
  term.loadAddon(fitAddon);
  console.log('✅ Terminal and FitAddon can be instantiated');
  
  // Try to fit to container
  const container = document.getElementById('terminal');
  if (container) {
    console.log('✅ Terminal container found');
    term.open(container);
    
    // Test minimum dimensions
    const dimensions = fitAddon.proposeDimensions();
    console.log('📏 Proposed dimensions:', dimensions);
    
    if (dimensions) {
      const minCols = Math.max(120, dimensions.cols || 80);
      const minRows = Math.max(24, dimensions.rows || 24);
      console.log('📏 Minimum dimensions enforced:', { cols: minCols, rows: minRows });
    }
    
    // Write test message
    term.write('Hello from browser console test\r\n');
    console.log('📝 Test message written to terminal');
  } else {
    console.error('❌ Terminal container not found');
  }
} catch (error) {
  console.error('❌ Terminal initialization error:', error);
}