const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:6969/ws/session');

ws.on('open', function open() {
  console.log('Connected to WebSocket');
  
  // Send a test message after connection
  setTimeout(() => {
    const message = JSON.stringify({ type: 'input', data: 'help\n' });
    ws.send(message);
    console.log('Sent help command');
  }, 1000);
});

ws.on('message', function message(data) {
  console.log('Received:', data.toString());
});

ws.on('error', function error(err) {
  console.error('WebSocket error:', err);
});

ws.on('close', function close() {
  console.log('WebSocket connection closed');
});