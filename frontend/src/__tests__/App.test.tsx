import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from '../App';

// Mock WebSocket with proper event handling
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
    
    // Simulate connection after a short delay
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

// Mock fetch
global.fetch = jest.fn() as any;

describe('App Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (global.fetch as jest.Mock).mockResolvedValue({
      json: () => Promise.resolve({ initialized: true }),
    });
  });

  test('renders main application with correct terminal labels', () => {
    render(<App />);
    
    expect(screen.getByText('Logs')).toBeInTheDocument();
    expect(screen.getByText('JUDO Terminal')).toBeInTheDocument();
  });

  test('switches between terminal tabs', async () => {
    render(<App />);
    
    const judoTerminalTab = screen.getByText('JUDO Terminal');
    fireEvent.click(judoTerminalTab);
    
    await waitFor(() => {
      // Should create a session WebSocket connection
      const sessionWs = MockWebSocket.instances.find(ws => ws.url.includes('session'));
      expect(sessionWs).toBeDefined();
    });

    const logsTab = screen.getByText('Logs');
    fireEvent.click(logsTab);
    
    await waitFor(() => {
      // Logs terminal should be active
      expect(screen.getByText('Source:')).toBeInTheDocument();
    });
  });

  test('handles service panel toggle', async () => {
    render(<App />);
    
    const serviceToggle = screen.getByText('▶ Services');
    fireEvent.click(serviceToggle);
    
    await waitFor(() => {
      expect(screen.getByText('Services')).toBeInTheDocument();
    });

    fireEvent.click(serviceToggle);
    
    await waitFor(() => {
      expect(screen.queryByText('Services')).not.toBeInTheDocument();
    });
  });

  test('displays connection status', async () => {
    render(<App />);
    
    await waitFor(() => {
      expect(screen.getByText('✓ Connected to log stream')).toBeInTheDocument();
    });

    // Simulate disconnection
    const wsInstance = MockWebSocket.instances[0];
    if (wsInstance && wsInstance.onclose) {
      wsInstance.onclose(new CloseEvent('close'));
      await waitFor(() => {
        expect(screen.getByText('✗ Log stream disconnected')).toBeInTheDocument();
      });
    }
  });

  test('handles log source selection', async () => {
    render(<App />);
    
    // Wait for initial connection
    await waitFor(() => {
      expect(screen.getByText('✓ Connected to log stream')).toBeInTheDocument();
    });
    
    // Find and change the source selector
    const sourceSelect = screen.getByDisplayValue('combined');
    fireEvent.change(sourceSelect, { target: { value: 'karaf' } });
    
    await waitFor(() => {
      // Should create a new WebSocket connection for karaf logs
      const karafWs = MockWebSocket.instances.find(ws => ws.url.includes('karaf'));
      expect(karafWs).toBeDefined();
    });
  });
});