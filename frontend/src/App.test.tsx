import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import App from './App';
import { mockWebSocket } from './setupTests';
import axios from 'axios';

// Mock react-xtermjs
vi.mock('react-xtermjs', () => ({
  XTerm: vi.fn(() => <div data-testid="mock-xterm" />),
  useXTerm: vi.fn(() => ({
    ref: { current: null },
    instance: {
      write: vi.fn(),
      writeln: vi.fn(),
      loadAddon: vi.fn(),
      onData: vi.fn(),
      resize: vi.fn(),
      refresh: vi.fn(),
      cols: 120,
      rows: 40,
    },
    fitAddon: {
      fit: vi.fn(),
    },
  })),
}));

// Mock axios
vi.mock('axios');
const mockedAxios = axios as vi.Mocked<typeof axios>;

describe('App Component', () => {
  beforeEach(() => {
    // Reset mocks before each test
    mockWebSocket.mockClear();
    mockedAxios.get.mockClear();
    mockedAxios.post.mockClear();

    // Mock project initialization status to true by default
    mockedAxios.get.mockImplementation((url: string) => {
      if (url.includes('/api/project/init/status')) {
        return Promise.resolve({ data: { initialized: true } });
      }
      if (url.includes('/api/services/status')) {
        return Promise.resolve({ 
          data: [
            { service: 'karaf', status: 'running', timestamp: '2024-01-01T00:00:00Z' },
            { service: 'postgresql', status: 'stopped', timestamp: '2024-01-01T00:00:00Z' },
            { service: 'keycloak', status: 'running', timestamp: '2024-01-01T00:00:00Z' },
          ]
        });
      }
      // Default mock for any other GET requests
      return Promise.resolve({ data: {} });
    });

    // Mock successful POST requests
    mockedAxios.post.mockResolvedValue({ data: {} });
  });

  it('renders main application with correct elements', async () => {
    render(<App />);
    expect(screen.getByText(/JUDO CLI Server/i)).toBeInTheDocument();
    expect(screen.getByText('Services')).toBeInTheDocument();
    expect(screen.getByText('Source:')).toBeInTheDocument();

    // Wait for initial data fetches to complete
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/api/project/init/status')));
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/api/services/status')));
  });

  it('handles log source selection', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    const sourceSelector = screen.getByRole('combobox');
    fireEvent.change(sourceSelector, { target: { value: 'karaf' } });

    // Expect log WebSocket to be connected to karaf source
    await waitFor(() => expect(mockWebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/logs/service/karaf')));
  });

  it('handles service panel toggle', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    const servicePanelButton = screen.getByText('▶ Services');
    fireEvent.click(servicePanelButton);
    expect(screen.getByText('All Services')).toBeInTheDocument();

    fireEvent.click(servicePanelButton);
    // The service panel remains in DOM but should not have 'open' class
    const servicePanel = screen.getByText('All Services').closest('.service-panel');
    expect(servicePanel).not.toHaveClass('open');
  });

  it('displays service statuses correctly', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    // Open service panel to see service statuses
    const servicePanelButton = screen.getByText('▶ Services');
    fireEvent.click(servicePanelButton);

    // Check that each service is displayed with correct status
    expect(screen.getByText('karaf')).toBeInTheDocument();
    expect(screen.getByText('postgresql')).toBeInTheDocument();
    expect(screen.getByText('keycloak')).toBeInTheDocument();
    
    // Check statuses using more specific queries
    const karafStatus = screen.getByText('karaf').closest('.service-control')?.querySelector('.service-status');
    const postgresqlStatus = screen.getByText('postgresql').closest('.service-control')?.querySelector('.service-status');
    const keycloakStatus = screen.getByText('keycloak').closest('.service-control')?.querySelector('.service-status');
    
    expect(karafStatus).toHaveTextContent('running');
    expect(postgresqlStatus).toHaveTextContent('stopped');
    expect(keycloakStatus).toHaveTextContent('running');
  });

  it('handles service start/stop actions', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    // Open service panel
    const servicePanelButton = screen.getByText('▶ Services');
    fireEvent.click(servicePanelButton);

    // Click start button for postgresql (which is stopped)
    const postgresqlStartButton = screen.getByText('postgresql').closest('.service-control')?.querySelector('.btn-service-start');
    fireEvent.click(postgresqlStartButton!);

    await waitFor(() => expect(mockedAxios.post).toHaveBeenCalledWith(expect.stringContaining('/api/services/postgresql/start')));
  });

  it('handles all services start/stop actions', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    // Open service panel
    const servicePanelButton = screen.getByText('▶ Services');
    fireEvent.click(servicePanelButton);

    // Click start all button
    const startAllButton = screen.getByText('Start All');
    fireEvent.click(startAllButton);

    await waitFor(() => expect(mockedAxios.post).toHaveBeenCalledWith(expect.stringContaining('/api/services/start')));
  });

  it('shows project initialization modal when project not initialized', async () => {
    // Mock project not initialized
    mockedAxios.get.mockImplementation((url: string) => {
      if (url.includes('/api/project/init/status')) {
        return Promise.resolve({ data: { initialized: false } });
      }
      if (url.includes('/api/services/status')) {
        return Promise.resolve({ 
          data: [
            { service: 'karaf', status: 'stopped', timestamp: '2024-01-01T00:00:00Z' },
            { service: 'postgresql', status: 'stopped', timestamp: '2024-01-01T00:00:00Z' },
            { service: 'keycloak', status: 'stopped', timestamp: '2024-01-01T00:00:00Z' },
          ]
        });
      }
      return Promise.resolve({ data: {} });
    });

    render(<App />);
    
    await waitFor(() => expect(screen.getByText('Project Not Initialized')).toBeInTheDocument());
    expect(screen.getByText('Yes, Initialize')).toBeInTheDocument();
    expect(screen.getByText('No, Continue Anyway')).toBeInTheDocument();
  });

  it('handles WebSocket connection for combined logs', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    // Default should be combined logs
    await waitFor(() => expect(mockWebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/logs/combined')));
  });

  it('handles service status polling after service actions', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    // Open service panel
    const servicePanelButton = screen.getByText('▶ Services');
    fireEvent.click(servicePanelButton);

    // Click start button for postgresql
    const postgresqlStartButton = screen.getByText('postgresql').closest('.service-control')?.querySelector('.btn-service-start');
    fireEvent.click(postgresqlStartButton!);

    // Should trigger status polling (multiple calls to fetchServiceStatuses)
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalledTimes(3)); // initial + 2 polling calls
  });

  it('handles WebSocket reconnection on source change', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    // Change source multiple times
    const sourceSelector = screen.getByRole('combobox');
    fireEvent.change(sourceSelector, { target: { value: 'karaf' } });
    fireEvent.change(sourceSelector, { target: { value: 'postgresql' } });
    fireEvent.change(sourceSelector, { target: { value: 'combined' } });

    // Should have multiple WebSocket connections
    await waitFor(() => expect(mockWebSocket).toHaveBeenCalledTimes(4)); // initial + 3 changes
  });

  it('handles API errors gracefully', async () => {
    // Mock API failure
    mockedAxios.get.mockRejectedValue(new Error('API Error'));
    
    render(<App />);
    
    // Application should still render despite API errors
    expect(screen.getByText(/JUDO CLI Server/i)).toBeInTheDocument();
    
    // Service panel should be available
    const servicePanelButton = screen.getByText('▶ Services');
    expect(servicePanelButton).toBeInTheDocument();
  });

  it('disables service buttons during loading states', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    // Open service panel
    const servicePanelButton = screen.getByText('▶ Services');
    fireEvent.click(servicePanelButton);

    // Click start button for postgresql
    const postgresqlStartButton = screen.getByText('postgresql').closest('.service-control')?.querySelector('.btn-service-start');
    fireEvent.click(postgresqlStartButton!);

    // Button should be disabled during loading
    await waitFor(() => expect(postgresqlStartButton).toBeDisabled());
  });
});