const WebSocket = require('ws');

// Test WebSocket connection to the server
const ws = new WebSocket('ws://localhost:6969/ws/logs/combined');

ws.on('open', function open() {
  console.log('✅ WebSocket connection established');
});

ws.on('message', function message(data) {
  console.log('📨 Received message:', data.toString());
});

ws.on('error', function error(err) {
  console.error('❌ WebSocket error:', err);
});

ws.on('close', function close() {
  console.log('🔌 WebSocket connection closed');
});

// Close after 5 seconds
setTimeout(() => {
  ws.close();
  process.exit(0);
}, 5000);