const WebSocket = require('ws');

// Test WebSocket connections
const testEndpoints = [
    'ws://localhost:6974/ws/logs/combined',
    'ws://localhost:6974/ws/logs/service/karaf'
];

function testWebSocket(url) {
    console.log(`Testing WebSocket: ${url}`);
    
    const ws = new WebSocket(url);
    
    ws.on('open', () => {
        console.log(`✅ Connected to ${url}`);
        // Send a ping to keep connection alive
        ws.ping();
    });
    
    ws.on('message', (data) => {
        console.log(`📨 Received message from ${url}: ${data.toString()}`);
    });
    
    ws.on('error', (error) => {
        console.log(`❌ Error with ${url}:`, error.message);
    });
    
    ws.on('close', (code, reason) => {
        console.log(`🔌 Connection closed to ${url}: code=${code}, reason=${reason}`);
    });
    
    // Set timeout to close connection after 5 seconds
    setTimeout(() => {
        if (ws.readyState === WebSocket.OPEN) {
            ws.close();
        }
    }, 5000);
}

// Test all endpoints
console.log('Starting WebSocket tests...');
testEndpoints.forEach(testWebSocket);

// Keep process alive for 10 seconds
setTimeout(() => {
    console.log('Tests completed');
    process.exit(0);
}, 10000);