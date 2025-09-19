const WebSocket = require('ws');

// Test WebSocket connection to the frontend
const ws = new WebSocket('ws://localhost:6969/ws/logs/combined');

ws.on('open', function open() {
  console.log('Connected to WebSocket server');
  
  // Send a test message
  const testMessage = {
    ts: new Date().toISOString(),
    service: 'test',
    line: 'This is a test message from WebSocket client'
  };
  
  ws.send(JSON.stringify(testMessage));
  console.log('Sent test message:', testMessage);
});

ws.on('message', function message(data) {
  console.log('Received:', data.toString());
});

ws.on('close', function close() {
  console.log('WebSocket connection closed');
});

ws.on('error', function error(err) {
  console.error('WebSocket error:', err);
});

// Close after 5 seconds
setTimeout(() => {
  ws.close();
  process.exit(0);
}, 5000);