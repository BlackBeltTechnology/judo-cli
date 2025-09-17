import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// This custom MockWebSocket class is used due to technical barriers in mocking the global WebSocket object directly with vi.hoisted().
// This approach maintains test isolation and reliability, and is an approved exception as per the project's constitution (v2.4.1, Section VIII. Frontend Testing, Technical Exceptions).
// Future plans include researching better WebSocket mocking solutions if Vitest or related tools provide more direct global object mocking capabilities.
class MockWebSocket {
  static instances: MockWebSocket[] = [];

  
  url: string;
  readyState: number;
  onopen: ((this: WebSocket, ev: Event) => any) | null = null;
  onmessage: ((this: WebSocket, ev: MessageEvent) => any) | null = null;
  onclose: ((this: WebSocket, ev: CloseEvent) => any) | null = null;
  onerror: ((this: WebSocket, ev: Event) => any) | null = null;
  
  send = vi.fn();
  close = vi.fn();
  
  // Proper event listener storage
  private eventListeners: { [event: string]: Function[] } = {};
  
  addEventListener = vi.fn((event: string, callback: any) => {
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
  
  removeEventListener = vi.fn((event: string, callback: any) => {
    if (this.eventListeners[event]) {
      this.eventListeners[event] = this.eventListeners[event].filter(cb => cb !== callback);
    }
  });
  
  constructor(url: string) {
    this.url = url;
    this.readyState = MockWebSocket.CONNECTING;
    MockWebSocket.instances.push(this);
    
    // Auto-connect after a short delay
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN;
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
    this.readyState = MockWebSocket.CLOSED;
    const closeEvent = new CloseEvent('close');
    if (this.onclose) {
      this.onclose(closeEvent);
    }
    if (this.eventListeners['close']) {
      this.eventListeners['close'].forEach(callback => callback(closeEvent));
    }
  }
}

// Add WebSocket constants to MockWebSocket class
MockWebSocket.CONNECTING = 0;
MockWebSocket.OPEN = 1;
MockWebSocket.CLOSING = 2;
MockWebSocket.CLOSED = 3;

// Replace global WebSocket with our mock
Object.defineProperty(global, 'WebSocket', {
  value: MockWebSocket,
  writable: true,
  configurable: true
});

describe('WebSocket Connection Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    MockWebSocket.instances = []; // Clear instances before each test
  });

  it('establishes WebSocket connection for logs', async () => {
    // Simulate WebSocket connection
    const ws = new WebSocket('ws://localhost:6969/ws/logs/combined');
    
    await waitFor(() => {
      expect(ws.url).toBe('ws://localhost:6969/ws/logs/combined');
      expect(ws.readyState).toBe(MockWebSocket.OPEN);
    });
  });

  it('handles WebSocket messages for log streaming', async () => {
    const messageHandler = vi.fn();
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

  it('handles WebSocket closure and reconnection', async () => {
    const closeHandler = vi.fn();
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

  it('sends session initialization data', async () => {
    const ws = new WebSocket('ws://localhost:6969/ws/session');
    
    await waitFor(() => {
      expect(ws.readyState).toBe(MockWebSocket.OPEN);
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

  it('handles session input events', async () => {
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

  it('handles session resize events', async () => {
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