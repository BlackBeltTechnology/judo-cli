#!/usr/bin/env python3
"""
WebSocket reconnection and error handling test for JUDO CLI server
"""

import asyncio
import websockets
import json
import time
import signal
import sys

class WebSocketTester:
    def __init__(self, url):
        self.url = url
        self.websocket = None
        self.reconnect_attempts = 0
        self.max_reconnect_attempts = 5
        self.reconnect_delay = 1  # 1 second initial delay
        self.max_reconnect_delay = 30  # 30 seconds max delay
        self.is_connected = False
        self.messages_received = 0
        self.errors_encountered = 0
        self.should_reconnect = True

    async def connect(self):
        print(f"Connecting to {self.url}...")
        
        try:
            self.websocket = await websockets.connect(self.url, ping_interval=20, ping_timeout=10)
            print("✅ WebSocket connection established")
            self.is_connected = True
            self.reconnect_attempts = 0
            self.reconnect_delay = 1
            
            # Start listening for messages
            await self.listen()
            
        except Exception as e:
            self.errors_encountered += 1
            print(f"❌ Connection failed: {e}")
            self.is_connected = False
            await self.attempt_reconnection()

    async def listen(self):
        try:
            async for message in self.websocket:
                self.messages_received += 1
                
                # Parse JSON messages
                try:
                    parsed = json.loads(message)
                    line_preview = parsed.get('line', '')[:50]
                    if len(parsed.get('line', '')) > 50:
                        line_preview += '...'
                    print(f"📨 Received: {parsed.get('service', 'unknown')} - {line_preview}")
                except json.JSONDecodeError:
                    print(f"📨 Received: {message[:100]}{'...' if len(message) > 100 else ''}")

                # Test reconnection after some messages
                if self.messages_received == 10:
                    print("🔄 Testing reconnection - closing connection...")
                    await self.websocket.close()
                    break
                    
        except websockets.exceptions.ConnectionClosed:
            print("🔌 WebSocket connection closed")
            self.is_connected = False
            await self.attempt_reconnection()
        except Exception as e:
            print(f"❌ Error in listen: {e}")
            self.is_connected = False
            await self.attempt_reconnection()

    async def attempt_reconnection(self):
        if not self.should_reconnect or self.reconnect_attempts >= self.max_reconnect_attempts:
            print("❌ Max reconnection attempts reached or reconnection disabled")
            return

        self.reconnect_attempts += 1
        delay = min(self.reconnect_delay * (2 ** (self.reconnect_attempts - 1)), self.max_reconnect_delay)
        
        print(f"🔄 Reconnection attempt {self.reconnect_attempts} in {delay}s...")
        
        await asyncio.sleep(delay)
        await self.connect()

    async def close(self):
        self.should_reconnect = False
        if self.websocket:
            await self.websocket.close()

    def get_stats(self):
        return {
            'messages_received': self.messages_received,
            'errors_encountered': self.errors_encountered,
            'reconnect_attempts': self.reconnect_attempts,
            'is_connected': self.is_connected
        }

async def main():
    # Test endpoints
    test_endpoints = [
        'ws://localhost:6969/ws/logs/combined',
        'ws://localhost:6969/ws/logs/service/karaf',
        'ws://localhost:6969/ws/logs/service/postgresql',
        'ws://localhost:6969/ws/logs/service/keycloak',
        'ws://localhost:6969/ws/session'
    ]
    
    endpoint = test_endpoints[0]  # combined logs
    print(f"Testing WebSocket endpoint: {endpoint}")
    
    tester = WebSocketTester(endpoint)
    
    # Handle graceful shutdown
    def signal_handler(sig, frame):
        print("\n🛑 Shutting down...")
        asyncio.create_task(tester.close())
        stats = tester.get_stats()
        print(f"📊 Final stats: {stats}")
        sys.exit(0)
    
    signal.signal(signal.SIGINT, signal_handler)
    
    # Start connection
    await tester.connect()
    
    # Run for 2 minutes
    await asyncio.sleep(120)
    
    print("\n⏰ Test completed (2 minutes elapsed)")
    stats = tester.get_stats()
    print(f"📊 Final stats: {stats}")
    await tester.close()

if __name__ == "__main__":
    asyncio.run(main())