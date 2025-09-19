const WebSocket = require('ws');

// Test WebSocket connections for judo server
// This test should be run after building with build.sh and running from test-model directory

const testEndpoints = [
    'ws://localhost:6969/ws/logs/combined',
    'ws://localhost:6969/ws/logs/service/karaf'
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
        try {
            const message = JSON.parse(data.toString());
            console.log(`📨 Received from ${url}: ${message.line || data.toString()}`);
        } catch (e) {
            console.log(`📨 Received raw message from ${url}: ${data.toString()}`);
        }
    });
    
    ws.on('error', (error) => {
        console.log(`❌ Error with ${url}:`, error.message);
    });
    
    ws.on('close', (code, reason) => {
        console.log(`🔌 Connection closed to ${url}: code=${code}, reason=${reason.toString()}`);
    });
    
    // Set timeout to close connection after 10 seconds
    setTimeout(() => {
        if (ws.readyState === WebSocket.OPEN) {
            ws.close();
        }
    }, 10000);
}

// Test all endpoints
console.log('Starting WebSocket tests for judo server (port 6969)...');
console.log('Make sure the server is running with: ./judo server');
console.log('Run this from test-model directory after building with build.sh\n');

testEndpoints.forEach(testWebSocket);

// Keep process alive for 15 seconds
setTimeout(() => {
    console.log('Tests completed');
    process.exit(0);
}, 15000);