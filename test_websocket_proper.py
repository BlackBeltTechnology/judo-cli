#!/usr/bin/env python3
"""
Test WebSocket connections for judo server
This test should be run after building with build.sh and running from test-model directory
"""
import asyncio
import websockets
import json
import sys

async def test_combined_logs():
    """Test combined logs WebSocket endpoint"""
    try:
        async with websockets.connect('ws://localhost:6969/ws/logs/combined') as ws:
            print("✅ Connected to combined logs WebSocket")
            
            # Wait for initial connection message
            message = await ws.recv()
            data = json.loads(message)
            print(f"📨 Received: {data.get('line', message)}")
            
            # Wait for a few log messages
            for i in range(5):
                try:
                    message = await asyncio.wait_for(ws.recv(), timeout=2.0)
                    data = json.loads(message)
                    print(f"📨 Log message: {data.get('line', message)}")
                except asyncio.TimeoutError:
                    print("⏰ Timeout waiting for log messages")
                    break
                    
    except Exception as e:
        print(f"❌ Error with combined logs: {e}")

async def test_karaf_logs():
    """Test Karaf service logs WebSocket endpoint"""
    try:
        async with websockets.connect('ws://localhost:6969/ws/logs/service/karaf') as ws:
            print("✅ Connected to Karaf logs WebSocket")
            
            # Wait for initial connection message
            message = await ws.recv()
            data = json.loads(message)
            print(f"📨 Received: {data.get('line', message)}")
            
            # Wait for a few log messages
            for i in range(5):
                try:
                    message = await asyncio.wait_for(ws.recv(), timeout=2.0)
                    data = json.loads(message)
                    print(f"📨 Karaf log: {data.get('line', message)}")
                except asyncio.TimeoutError:
                    print("⏰ Timeout waiting for Karaf log messages")
                    break
                    
    except Exception as e:
        print(f"❌ Error with Karaf logs: {e}")

async def main():
    print("Starting WebSocket tests for judo server (port 6969)...")
    print("Make sure the server is running with: ./judo server")
    print("Run this from test-model directory after building with build.sh\n")
    
    await asyncio.gather(
        test_combined_logs(),
        test_karaf_logs()
    )

if __name__ == "__main__":
    asyncio.run(main())