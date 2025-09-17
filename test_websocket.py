#!/usr/bin/env python3
import websockets
import asyncio
import json

async def test_session():
    try:
        async with websockets.connect('ws://localhost:6969/ws/session') as ws:
            print("Connected to WebSocket")
            
            # Wait for handshake message
            message = await ws.recv()
            print(f"Received: {message}")
            
            # Send a test message
            test_msg = json.dumps({"type": "input", "data": "help\n"})
            await ws.send(test_msg)
            print("Sent help command")
            
            # Wait for response
            response = await ws.recv()
            print(f"Response: {response}")
            
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    asyncio.run(test_session())