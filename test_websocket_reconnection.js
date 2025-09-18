const WebSocket = require('ws');

// Test WebSocket reconnection and error handling
class WebSocketTester {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // 1 second initial delay
        this.maxReconnectDelay = 30000; // 30 seconds max delay
        this.isConnected = false;
        this.messagesReceived = 0;
        this.errorsEncountered = 0;
    }

    connect() {
        console.log(`Connecting to ${this.url}...`);
        
        this.ws = new WebSocket(this.url);
        
        this.ws.on('open', () => {
            console.log('✅ WebSocket connection established');
            this.isConnected = true;
            this.reconnectAttempts = 0;
            this.reconnectDelay = 1000;
        });

        this.ws.on('message', (data) => {
            this.messagesReceived++;
            const message = data.toString();
            
            // Parse JSON messages for structured logging
            try {
                const parsed = JSON.parse(message);
                console.log(`📨 Received: ${parsed.service} - ${parsed.line.substring(0, 50)}${parsed.line.length > 50 ? '...' : ''}`);
            } catch (e) {
                console.log(`📨 Received: ${message.substring(0, 100)}${message.length > 100 ? '...' : ''}`);
            }

            // Test reconnection by closing connection after some messages
            if (this.messagesReceived === 10) {
                console.log('🔄 Testing reconnection - closing connection...');
                this.ws.close();
            }
        });

        this.ws.on('error', (error) => {
            this.errorsEncountered++;
            console.log(`❌ WebSocket error: ${error.message}`);
            this.isConnected = false;
        });

        this.ws.on('close', (code, reason) => {
            console.log(`🔌 WebSocket closed: code=${code}, reason=${reason}`);
            this.isConnected = false;
            
            // Attempt reconnection if not manually closed and under max attempts
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
                this.attemptReconnection();
            } else {
                console.log('❌ Max reconnection attempts reached');
            }
        });

        // Set timeout to test connection stability
        setTimeout(() => {
            if (this.isConnected) {
                console.log('✅ Connection stable for 30 seconds');
            }
        }, 30000);
    }

    attemptReconnection() {
        this.reconnectAttempts++;
        const delay = Math.min(this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1), this.maxReconnectDelay);
        
        console.log(`🔄 Reconnection attempt ${this.reconnectAttempts} in ${delay}ms...`);
        
        setTimeout(() => {
            this.connect();
        }, delay);
    }

    close() {
        if (this.ws) {
            this.ws.close();
        }
    }

    getStats() {
        return {
            messagesReceived: this.messagesReceived,
            errorsEncountered: this.errorsEncountered,
            reconnectAttempts: this.reconnectAttempts,
            isConnected: this.isConnected
        };
    }
}

// Test different WebSocket endpoints
const testEndpoints = [
    'ws://localhost:6969/ws/logs/combined',
    'ws://localhost:6969/ws/logs/service/karaf',
    'ws://localhost:6969/ws/logs/service/postgresql',
    'ws://localhost:6969/ws/logs/service/keycloak',
    'ws://localhost:6969/ws/session'
];

// Test a specific endpoint
const endpointToTest = testEndpoints[0]; // combined logs
console.log(`Testing WebSocket endpoint: ${endpointToTest}`);

const tester = new WebSocketTester(endpointToTest);
tester.connect();

// Handle graceful shutdown
process.on('SIGINT', () => {
    console.log('\n🛑 Shutting down...');
    const stats = tester.getStats();
    console.log('📊 Final stats:', stats);
    tester.close();
    process.exit(0);
});

// Run for 2 minutes then exit
setTimeout(() => {
    console.log('\n⏰ Test completed (2 minutes elapsed)');
    const stats = tester.getStats();
    console.log('📊 Final stats:', stats);
    tester.close();
    process.exit(0);
}, 120000);