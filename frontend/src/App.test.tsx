import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import App from './App';
import { mockWebSocket } from './setupTests';
import axios from 'axios';


// Mock react-xtermjs
vi.mock('react-xtermjs', () => ({
  XTerm: vi.fn(() => <div data-testid="mock-xterm" />),
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
            { service: 'karaf', status: 'running', timestamp: '' },
            { service: 'postgresql', status: 'stopped', timestamp: '' },
            { service: 'keycloak', status: 'running', timestamp: '' },
          ]
        });
      }
      // Default mock for any other GET requests
      return Promise.resolve({ data: {} });
    });
  });

  it('renders main application with correct terminal labels', async () => {
    render(<App />);
    expect(screen.getByText(/JUDO CLI Server/i)).toBeInTheDocument();
    expect(screen.getByText('Logs')).toBeInTheDocument();
    expect(screen.getByText('JUDO Terminal')).toBeInTheDocument();
    expect(screen.getByText('Services')).toBeInTheDocument();

    // Wait for initial data fetches to complete
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/api/project/init/status')));
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalledWith(expect.stringContaining('/api/services/status')));
  });

  it('switches between terminal tabs', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    const judoTerminalTab = screen.getByText('JUDO Terminal');
    fireEvent.click(judoTerminalTab);

    // Expect session WebSocket to be connected
    await waitFor(() => expect(mockWebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/session')));

    const logsTab = screen.getByText('Logs');
    fireEvent.click(logsTab);

    // Expect log WebSocket to be connected
    await waitFor(() => expect(mockWebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/logs/combined')));
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

  it('displays connection status', async () => {
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

  it('handles log source selection', async () => {
    render(<App />);
    await waitFor(() => expect(mockedAxios.get).toHaveBeenCalled());

    const logsTab = screen.getByText('Logs');
    fireEvent.click(logsTab);

    const sourceSelector = screen.getByRole('combobox');
    fireEvent.change(sourceSelector, { target: { value: 'karaf' } });

    // Expect log WebSocket to be connected to karaf source
    await waitFor(() => expect(mockWebSocket).toHaveBeenCalledWith(expect.stringContaining('/ws/logs/service/karaf')));
  });
});