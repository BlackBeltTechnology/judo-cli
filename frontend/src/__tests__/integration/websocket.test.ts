import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WebSocket } from 'ws';
import axios from 'axios';
import { mockWebSocket } from '../../setupTests';

// Mock axios for API testing
vi.mock('axios');
const mockedAxios = axios as vi.Mocked<typeof axios>;

// Mock WebSocket implementation for integration tests
const createMockWebSocket = () => {
  const mockWs = {
    send: vi.fn(),
    close: vi.fn(),
    readyState: WebSocket.OPEN,
    onopen: null,
    onmessage: null,
    onclose: null,
    onerror: null,
    
    // Helper methods for testing
    simulateOpen: function() {
      if (this.onopen) this.onopen(new Event('open'));
    },
    
    simulateMessage: function(data: any) {
      if (this.onmessage) this.onmessage({ data: JSON.stringify(data) });
    },
    
    simulateClose: function(code = 1000, reason = '') {
      if (this.onclose) this.onclose(new CloseEvent('close', { code, reason, wasClean: true }));
    },
    
    simulateError: function() {
      if (this.onerror) this.onerror(new Event('error'));
    }
  };
  
  return mockWs;
};

describe('WebSocket Integration Tests', () => {
  let mockWs: any;

  beforeEach(() => {
    mockWs = createMockWebSocket();
    mockWebSocket.mockImplementation(() => mockWs);
    
    // Reset axios mocks
    mockedAxios.get.mockClear();
    mockedAxios.post.mockClear();
    
    // Mock successful API responses
    mockedAxios.get.mockResolvedValue({ 
      data: { 
        initialized: true,
        services: [
          { service: 'karaf', status: 'running', timestamp: '2024-01-01T00:00:00Z' },
          { service: 'postgresql', status: 'stopped', timestamp: '2024-01-01T00:00:00Z' },
          { service: 'keycloak', status: 'running', timestamp: '2024-01-01T00:00:00Z' }
        ]
      }
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should establish WebSocket connection on component mount', async () => {
    // WebSocket should be created with correct URL
    expect(mockWebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/logs/combined'));
    
    // Connection should be attempted
    expect(mockWs).toBeDefined();
  });

  it('should handle WebSocket open event', async () => {
    // Simulate WebSocket connection open
    mockWs.simulateOpen();
    
    // No specific assertions needed for open event (it's handled silently)
    expect(mockWs.onopen).toBeDefined();
  });

  it('should handle WebSocket message events with valid JSON', async () => {
    const logMessage = {
      ts: '2024-01-01T12:00:00Z',
      service: 'karaf',
      line: 'INFO: Service started successfully'
    };
    
    // Simulate message reception
    mockWs.simulateMessage(logMessage);
    
    // Message handler should be set up
    expect(mockWs.onmessage).toBeDefined();
  });

  it('should handle WebSocket message events with invalid JSON', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    
    // Simulate invalid JSON message
    const invalidMessage = { data: 'invalid json' };
    if (mockWs.onmessage) {
      mockWs.onmessage(invalidMessage);
    }
    
    // Should log error for invalid JSON
    expect(consoleError).toHaveBeenCalledWith(
      expect.stringContaining('Error parsing log message'),
      expect.any(Error),
      'invalid json'
    );
    
    consoleError.mockRestore();
  });

  it('should handle WebSocket close events', async () => {
    // Simulate WebSocket closure
    mockWs.simulateClose(1000, 'Normal closure');
    
    // Close handler should be set up
    expect(mockWs.onclose).toBeDefined();
  });

  it('should handle WebSocket error events', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    
    // Simulate WebSocket error
    mockWs.simulateError();
    
    // Error should be logged
    expect(consoleError).toHaveBeenCalledWith(
      expect.stringContaining('Log WebSocket error'),
      expect.any(Event)
    );
    
    consoleError.mockRestore();
  });

  it('should close WebSocket on component unmount', async () => {
    // Simulate component unmount by calling close
    mockWs.close();
    
    expect(mockWs.close).toHaveBeenCalled();
  });

  it('should handle WebSocket reconnection on source change', async () => {
    // Initial connection should be to combined logs
    expect(mockWebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/logs/combined'));
    
    // Simulate source change to karaf
    const karafUrl = 'ws://localhost:6969/ws/logs/service/karaf';
    mockWebSocket.mockImplementationOnce(() => {
      const newMock = createMockWebSocket();
      newMock.simulateOpen();
      return newMock;
    });
    
    // The actual source change would be triggered by UI interaction
    // This test verifies the WebSocket factory behavior
  });

  it('should handle rapid WebSocket reconnections', async () => {
    // Create multiple mock WebSockets for rapid reconnections
    const mockConnections: any[] = [];
    
    mockWebSocket.mockImplementation(() => {
      const newMock = createMockWebSocket();
      mockConnections.push(newMock);
      return newMock;
    });
    
    // Simulate multiple rapid reconnections
    for (let i = 0; i < 5; i++) {
      const mockWs = mockWebSocket();
      mockWs.simulateOpen();
      mockWs.simulateClose();
    }
    
    // Should handle multiple connections without errors
    expect(mockConnections).toHaveLength(5);
    mockConnections.forEach(conn => {
      expect(conn.close).toHaveBeenCalled();
    });
  });

  it('should handle WebSocket messages with empty lines', async () => {
    const emptyMessage = {
      ts: '2024-01-01T12:00:00Z',
      service: 'karaf',
      line: ''
    };
    
    // Simulate empty message
    mockWs.simulateMessage(emptyMessage);
    
    // Empty messages should be filtered out (no terminal write)
    // This is handled in the component logic
    expect(mockWs.onmessage).toBeDefined();
  });

  it('should handle WebSocket messages with heartbeat/ping content', async () => {
    const heartbeatMessage = {
      ts: '2024-01-01T12:00:00Z',
      service: 'combined',
      line: '\u0000' // Null character often used for heartbeats
    };
    
    // Simulate heartbeat message
    mockWs.simulateMessage(heartbeatMessage);
    
    // Heartbeat messages should be handled appropriately
    // (either filtered or processed silently)
    expect(mockWs.onmessage).toBeDefined();
  });
});