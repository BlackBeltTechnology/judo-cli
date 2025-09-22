const WebSocket = require('ws');

// Test WebSocket connections
async function testWebSocket(url, name) {
  console.log(`Testing ${name} WebSocket: ${url}`);
  
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(url);
    
    ws.on('open', () => {
      console.log(`✓ ${name} WebSocket connected successfully`);
      ws.close();
      resolve(true);
    });
    
    ws.on('message', (data) => {
      console.log(`📨 ${name} received:`, data.toString());
    });
    
    ws.on('error', (error) => {
      console.log(`✗ ${name} WebSocket error:`, error.message);
      reject(error);
    });
    
    ws.on('close', (code, reason) => {
      console.log(`🔌 ${name} WebSocket closed:`, code, reason.toString());
    });
    
    // Timeout after 5 seconds
    setTimeout(() => {
      if (ws.readyState !== WebSocket.OPEN) {
        console.log(`⏰ ${name} WebSocket connection timeout`);
        ws.close();
        reject(new Error('Connection timeout'));
      }
    }, 5000);
  });
}

async function testAll() {
  try {
    console.log('Testing WebSocket connections to localhost:6969...\n');
    
    await testWebSocket('ws://localhost:6969/ws/logs/combined', 'Logs Combined');
    console.log('');
    
    await testWebSocket('ws://localhost:6969/ws/session', 'Session');
    console.log('');
    
    console.log('✅ All WebSocket tests passed!');
    process.exit(0);
  } catch (error) {
    console.log('❌ WebSocket test failed:', error.message);
    process.exit(1);
  }
}

testAll();