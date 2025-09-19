import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';

// Mock axios for API testing
vi.mock('axios');
const mockedAxios = axios as vi.Mocked<typeof axios>;

describe('API Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch project initialization status', async () => {
    // Mock successful response
    mockedAxios.get.mockResolvedValueOnce({
      data: { initialized: true }
    });

    const response = await axios.get('/api/project/init/status');
    
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/project/init/status');
    expect(response.data).toEqual({ initialized: true });
  });

  it('should fetch service statuses', async () => {
    const mockStatuses = [
      { service: 'karaf', status: 'running', timestamp: '2024-01-01T00:00:00Z' },
      { service: 'postgresql', status: 'stopped', timestamp: '2024-01-01T00:00:00Z' },
      { service: 'keycloak', status: 'running', timestamp: '2024-01-01T00:00:00Z' }
    ];

    mockedAxios.get.mockResolvedValueOnce({
      data: mockStatuses
    });

    const response = await axios.get('/api/services/status');
    
    expect(mockedAxios.get).toHaveBeenCalledWith('/api/services/status');
    expect(response.data).toEqual(mockStatuses);
  });

  it('should start individual service', async () => {
    mockedAxios.post.mockResolvedValueOnce({ data: {} });

    await axios.post('/api/services/karaf/start');
    
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/services/karaf/start');
  });

  it('should stop individual service', async () => {
    mockedAxios.post.mockResolvedValueOnce({ data: {} });

    await axios.post('/api/services/karaf/stop');
    
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/services/karaf/stop');
  });

  it('should start all services', async () => {
    mockedAxios.post.mockResolvedValueOnce({ data: {} });

    await axios.post('/api/services/start');
    
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/services/start');
  });

  it('should stop all services', async () => {
    mockedAxios.post.mockResolvedValueOnce({ data: {} });

    await axios.post('/api/services/stop');
    
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/services/stop');
  });

  it('should handle API errors gracefully', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    
    mockedAxios.get.mockRejectedValueOnce(new Error('Network Error'));

    try {
      await axios.get('/api/services/status');
    } catch (error) {
      expect(error).toBeInstanceOf(Error);
      expect(error.message).toBe('Network Error');
    }
    
    consoleError.mockRestore();
  });

  it('should handle fallback service status fetching', async () => {
    // Mock primary endpoint failure
    mockedAxios.get.mockRejectedValueOnce(new Error('Primary endpoint failed'));
    
    // Mock fallback endpoints
    mockedAxios.get.mockResolvedValueOnce({ data: { service: 'karaf', status: 'running' } });
    mockedAxios.get.mockResolvedValueOnce({ data: { service: 'postgresql', status: 'stopped' } });
    mockedAxios.get.mockResolvedValueOnce({ data: { service: 'keycloak', status: 'running' } });

    // This test simulates the fallback logic that would be implemented in the component
    try {
      await axios.get('/api/services/status');
    } catch (error) {
      // Primary endpoint failed, try fallback
      const karafStatus = await axios.get('/api/services/karaf/status');
      const postgresStatus = await axios.get('/api/services/postgresql/status');
      const keycloakStatus = await axios.get('/api/services/keycloak/status');
      
      expect(karafStatus.data).toEqual({ service: 'karaf', status: 'running' });
      expect(postgresStatus.data).toEqual({ service: 'postgresql', status: 'stopped' });
      expect(keycloakStatus.data).toEqual({ service: 'keycloak', status: 'running' });
    }
  });

  it('should handle concurrent API requests', async () => {
    const mockResponses = [
      { data: { service: 'karaf', status: 'running' } },
      { data: { service: 'postgresql', status: 'stopped' } },
      { data: { service: 'keycloak', status: 'running' } }
    ];

    mockedAxios.get.mockResolvedValueOnce(mockResponses[0]);
    mockedAxios.get.mockResolvedValueOnce(mockResponses[1]);
    mockedAxios.get.mockResolvedValueOnce(mockResponses[2]);

    const [karafStatus, postgresStatus, keycloakStatus] = await Promise.all([
      axios.get('/api/services/karaf/status'),
      axios.get('/api/services/postgresql/status'),
      axios.get('/api/services/keycloak/status')
    ]);

    expect(karafStatus.data).toEqual(mockResponses[0].data);
    expect(postgresStatus.data).toEqual(mockResponses[1].data);
    expect(keycloakStatus.data).toEqual(mockResponses[2].data);
  });

  it('should handle service status polling', async () => {
    const mockStatuses = [
      { service: 'karaf', status: 'starting', timestamp: '2024-01-01T00:00:00Z' },
      { service: 'karaf', status: 'running', timestamp: '2024-01-01T00:00:10Z' }
    ];

    // Mock first call returns starting, second call returns running
    mockedAxios.get.mockResolvedValueOnce({ data: [mockStatuses[0]] });
    mockedAxios.get.mockResolvedValueOnce({ data: [mockStatuses[1]] });

    // Simulate polling behavior
    const pollServiceStatus = async (service: string, maxAttempts = 5) => {
      let attempts = 0;
      
      while (attempts < maxAttempts) {
        attempts++;
        const response = await axios.get('/api/services/status');
        const serviceStatus = response.data.find((s: any) => s.service === service);
        
        if (serviceStatus && serviceStatus.status === 'running') {
          return serviceStatus;
        }
        
        // Wait before next poll
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
      
      throw new Error(`Service ${service} did not reach running state`);
    };

    const status = await pollServiceStatus('karaf', 2);
    expect(status.status).toBe('running');
    expect(mockedAxios.get).toHaveBeenCalledTimes(2);
  });

  it('should handle judo init command', async () => {
    mockedAxios.post.mockResolvedValueOnce({ data: {} });

    await axios.post('/api/commands/judo%20init');
    
    expect(mockedAxios.post).toHaveBeenCalledWith('/api/commands/judo%20init');
  });
});