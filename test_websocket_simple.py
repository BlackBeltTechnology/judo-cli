#!/usr/bin/env python3
"""
Simple WebSocket test using built-in libraries
This test should be run after building with build.sh and running from test-model directory
"""
import socket
import ssl
import base64
import json

def test_websocket(url):
    """Test WebSocket connection using raw sockets"""
    try:
        # Parse URL
        if url.startswith('ws://'):
            host = url[5:].split('/')[0]
            port = 80
            path = '/' + '/'.join(url[5:].split('/')[1:])
        elif url.startswith('wss://'):
            host = url[6:].split('/')[0]
            port = 443
            path = '/' + '/'.join(url[6:].split('/')[1:])
        else:
            print(f"❌ Invalid URL: {url}")
            return
        
        # Create socket connection
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(5)
        
        if port == 443:
            context = ssl.create_default_context()
            sock = context.wrap_socket(sock, server_hostname=host)
        
        sock.connect((host, port))
        
        # WebSocket handshake
        key = base64.b64encode(b'judo-test-key').decode()
        handshake = (
            f"GET {path} HTTP/1.1\r\n"
            f"Host: {host}\r\n"
            "Upgrade: websocket\r\n"
            "Connection: Upgrade\r\n"
            f"Sec-WebSocket-Key: {key}\r\n"
            "Sec-WebSocket-Version: 13\r\n"
            "\r\n"
        )
        
        sock.send(handshake.encode())
        
        # Read response
        response = sock.recv(1024).decode()
        if "101 Switching Protocols" in response:
            print(f"✅ WebSocket handshake successful for {url}")
            print(f"   Response: {response.split('\\r\\n')[0]}")
        else:
            print(f"❌ WebSocket handshake failed for {url}")
            print(f"   Response: {response}")
            
        sock.close()
        
    except Exception as e:
        print(f"❌ Error testing {url}: {e}")

def main():
    print("Testing WebSocket connectivity for judo server...")
    print("Make sure the server is running with: ./judo server")
    print("Run this from test-model directory after building with build.sh\n")
    
    endpoints = [
        'ws://localhost:6969/ws/logs/combined',
        'ws://localhost:6969/ws/logs/service/karaf'
    ]
    
    for endpoint in endpoints:
        test_websocket(endpoint)
        print()

if __name__ == "__main__":
    main()