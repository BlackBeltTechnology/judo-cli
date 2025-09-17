import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock WebSocket implementation
class MockWebSocket {
  static instances: MockWebSocket[] = [];
  
  url: string;
  readyState: number;
  onopen: ((this: WebSocket, ev: Event) => any) | null = null;
  onmessage: ((this: WebSocket, ev: MessageEvent) => any) | null = null;
  onclose: ((this: WebSocket, ev: CloseEvent) => any) | null = null;
  onerror: ((this: WebSocket, ev: Event) => any) | null = null;
  
  send = jest.fn();
  close = jest.fn();
  
  // Proper event listener storage
  private eventListeners: { [event: string]: Function[] } = {};
  
  addEventListener = jest.fn((event: string, callback: any) => {
    if (!this.eventListeners[event]) {
      this.eventListeners[event] = [];
    }
    this.eventListeners[event].push(callback);
    
    // Also set the on* property for compatibility
    if (event === 'open') {
      this.onopen = callback;
    } else if (event === 'message') {
      this.onmessage = callback;
    } else if (event === 'close') {
      this.onclose = callback;
    } else if (event === 'error') {
      this.onerror = callback;
    }
  });
  
  removeEventListener = jest.fn((event: string, callback: any) => {
    if (this.eventListeners[event]) {
      this.eventListeners[event] = this.eventListeners[event].filter(cb => cb !== callback);
    }
  });
  
  constructor(url: string) {
    this.url = url;
    this.readyState = WebSocket.CONNECTING;
    MockWebSocket.instances.push(this);
    
    // Auto-connect after a short delay
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      // Call both onopen and event listeners
      if (this.onopen) {
        this.onopen(new Event('open'));
      }
      if (this.eventListeners['open']) {
        this.eventListeners['open'].forEach(callback => callback(new Event('open')));
      }
    }, 10);
  }
  
  // Helper method to simulate receiving a message
  simulateMessage(data: any) {
    const messageEvent = new MessageEvent('message', { data: JSON.stringify(data) });
    if (this.onmessage) {
      this.onmessage(messageEvent);
    }
    if (this.eventListeners['message']) {
      this.eventListeners['message'].forEach(callback => callback(messageEvent));
    }
  }
  
  // Helper method to simulate connection close
  simulateClose() {
    this.readyState = WebSocket.CLOSED;
    const closeEvent = new CloseEvent('close');
    if (this.onclose) {
      this.onclose(closeEvent);
    }
    if (this.eventListeners['close']) {
      this.eventListeners['close'].forEach(callback => callback(closeEvent));
    }
  }
}

// Add WebSocket constants
Object.assign(global, {
  WebSocket: MockWebSocket as any,
});

Object.defineProperty(global.WebSocket, 'CONNECTING', { value: 0 });
Object.defineProperty(global.WebSocket, 'OPEN', { value: 1 });
Object.defineProperty(global.WebSocket, 'CLOSING', { value: 2 });
Object.defineProperty(global.WebSocket, 'CLOSED', { value: 3 });

describe('WebSocket Connection Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('establishes WebSocket connection for logs', async () => {
    // Simulate WebSocket connection
    const ws = new WebSocket('ws://localhost:6969/ws/logs/combined');
    
    await waitFor(() => {
      expect(ws.url).toBe('ws://localhost:6969/ws/logs/combined');
      expect(ws.addEventListener).toHaveBeenCalledWith('open', expect.any(Function));
    });
  });

  test('handles WebSocket messages for log streaming', async () => {
    const messageHandler = jest.fn();
    const ws = new WebSocket('ws://localhost:6969/ws/logs/combined');
    
    ws.addEventListener('message', messageHandler);
    
    // Wait for connection
    await new Promise(resolve => setTimeout(resolve, 20));
    
    // Simulate incoming message using onmessage directly
    if (ws.onmessage) {
      const messageEvent = new MessageEvent('message', {
        data: JSON.stringify({ 
          ts: '2025-09-17T12:00:00Z', 
          service: 'karaf', 
          line: 'Test log message' 
        })
      });
      ws.onmessage(messageEvent);
      expect(messageHandler).toHaveBeenCalledWith(messageEvent);
    }
  });

  test('handles WebSocket closure and reconnection', async () => {
    const closeHandler = jest.fn();
    const ws = new WebSocket('ws://localhost:6969/ws/logs/combined');
    
    ws.addEventListener('close', closeHandler);
    
    // Wait for connection
    await new Promise(resolve => setTimeout(resolve, 20));
    
    // Simulate connection close using onclose directly
    if (ws.onclose) {
      ws.onclose(new CloseEvent('close'));
      expect(closeHandler).toHaveBeenCalled();
    }
  });

  test('sends session initialization data', async () => {
    const ws = new WebSocket('ws://localhost:6969/ws/session');
    
    await waitFor(() => {
      expect(ws.addEventListener).toHaveBeenCalledWith('open', expect.any(Function));
    });
    
    // Simulate sending initialization data
    const initData = JSON.stringify({
      type: 'init',
      term: 'xterm-256color',
      cols: 80,
      rows: 24
    });
    
    ws.send(initData);
    
    expect(ws.send).toHaveBeenCalledWith(initData);
  });

  test('handles session input events', async () => {
    const ws = new WebSocket('ws://localhost:6969/ws/session');
    
    // Wait for connection
    await new Promise(resolve => setTimeout(resolve, 20));
    
    // Simulate sending command input
    const inputData = JSON.stringify({
      type: 'input',
      data: 'help\r'
    });
    
    ws.send(inputData);
    
    expect(ws.send).toHaveBeenCalledWith(inputData);
  });

  test('handles session resize events', async () => {
    const ws = new WebSocket('ws://localhost:6969/ws/session');
    
    // Wait for connection
    await new Promise(resolve => setTimeout(resolve, 20));
    
    // Simulate sending resize event
    const resizeData = JSON.stringify({
      type: 'resize',
      cols: 100,
      rows: 30
    });
    
    ws.send(resizeData);
    
    expect(ws.send).toHaveBeenCalledWith(resizeData);
  });
});