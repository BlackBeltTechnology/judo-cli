const WebSocket = require('ws');

// Test terminal dimensions by sending a resize message
const ws = new WebSocket('ws://localhost:6969/ws/session');

ws.on('open', () => {
  console.log('✅ Connected to session WebSocket');
  
  // Send a resize message to test terminal dimensions
  const resizeMsg = {
    type: 'resize',
    cols: 120,
    rows: 24
  };
  
  ws.send(JSON.stringify(resizeMsg));
  console.log('📤 Sent resize message:', resizeMsg);
  
  // Send a test message
  const testMsg = {
    type: 'input',
    data: 'echo "Testing terminal dimensions"\r'
  };
  
  setTimeout(() => {
    ws.send(JSON.stringify(testMsg));
    console.log('📤 Sent test message:', testMsg);
  }, 1000);
});

ws.on('message', (data) => {
  const message = data.toString();
  console.log('📨 Received:', message);
});

ws.on('close', (code, reason) => {
  console.log(`🔌 Connection closed: code=${code}, reason=${reason}`);
});

ws.on('error', (error) => {
  console.error('❌ WebSocket error:', error);
});