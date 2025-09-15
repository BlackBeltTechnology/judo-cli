import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import './App.css';

interface Command {
  command: string;
  output: string;
  success: boolean;
}

interface Status {
  status: string;
  timestamp: string;
}

interface ServiceStatus {
  service: string;
  status: string;
  timestamp: string;
}

function App() {
  const [input, setInput] = useState('');
  const [commands, setCommands] = useState<Command[]>([]);
  const [status, setStatus] = useState<Status>({ status: 'unknown', timestamp: '' });
  const [serviceStatus, setServiceStatus] = useState<{[key: string]: ServiceStatus}>({});
  const [logs, setLogs] = useState<string[]>([]);
  const [logFilter, setLogFilter] = useState<string>('all');
  const [isConnected, setIsConnected] = useState(false);
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    fetchStatus();
    fetchServiceStatuses();
    connectWebSocket();
    
    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);

  const getApiBaseUrl = () => {
    const { protocol, hostname, port } = window.location;
    return `${protocol}//${hostname}:${port}`;
  };

  const connectWebSocket = () => {
    const { protocol, hostname, port } = window.location;
    const wsProtocol = protocol === 'https:' ? 'wss:' : 'ws:';
    ws.current = new WebSocket(`${wsProtocol}//${hostname}:${port}/ws/logs`);
    
    ws.current.onopen = () => {
      console.log('WebSocket connected');
      setIsConnected(true);
    };
    
    ws.current.onmessage = (event) => {
      setLogs(prev => [...prev.slice(-100), event.data]);
    };
    
    ws.current.onclose = () => {
      console.log('WebSocket disconnected');
      setIsConnected(false);
      // Attempt to reconnect after 2 seconds
      setTimeout(connectWebSocket, 2000);
    };
    
    ws.current.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  };

  const fetchStatus = async () => {
    try {
      const response = await axios.get(`${getApiBaseUrl()}/api/status`);
      setStatus(response.data);
    } catch (error) {
      console.error('Failed to fetch status:', error);
    }
  };

  const fetchServiceStatuses = async () => {
    try {
      const [karaf, postgres, keycloak] = await Promise.all([
        axios.get(`${getApiBaseUrl()}/api/services/karaf/status`),
        axios.get(`${getApiBaseUrl()}/api/services/postgresql/status`),
        axios.get(`${getApiBaseUrl()}/api/services/keycloak/status`)
      ]);
      
      setServiceStatus({
        karaf: karaf.data,
        postgresql: postgres.data,
        keycloak: keycloak.data
      });
    } catch (error) {
      console.error('Failed to fetch service statuses:', error);
    }
  };

  const handleStart = async () => {
    try {
      await axios.post(`${getApiBaseUrl()}/api/actions/start`);
      fetchStatus();
    } catch (error) {
      console.error('Failed to start:', error);
    };
  };

  const handleStop = async () => {
    try {
      await axios.post(`${getApiBaseUrl()}/api/actions/stop`);
      fetchStatus();
    } catch (error) {
      console.error('Failed to stop:', error);
    }
  };

  const handleServiceStart = async (service: string) => {
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/start`);
      fetchServiceStatuses();
    } catch (error) {
      console.error(`Failed to start ${service}:`, error);
    }
  };

  const handleServiceStop = async (service: string) => {
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/stop`);
      fetchServiceStatuses();
    } catch (error) {
      console.error(`Failed to stop ${service}:`, error);
    }
  };

  const handleServiceStatus = async (service: string) => {
    try {
      await axios.get(`${getApiBaseUrl()}/api/services/${service}/status`);
      fetchServiceStatuses();
    } catch (error) {
      console.error(`Failed to get ${service} status:`, error);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    try {
      const response = await axios.post(`${getApiBaseUrl()}/api/commands/${encodeURIComponent(input)}`);
      setCommands(prev => [...prev, response.data]);
      setInput('');
    } catch (error) {
      console.error('Command failed:', error);
      setCommands(prev => [...prev, {
        command: input,
        output: 'Error executing command',
        success: false
      }]);
      setInput('');
    };
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>JUDO CLI Server</h1>
        <div className="status-bar">
          <span>Status: {status.status}</span>
          <span>WebSocket: {isConnected ? 'Connected' : 'Disconnected'}</span>
          <span>Last updated: {new Date(status.timestamp).toLocaleTimeString()}</span>
        </div>
      </header>

      <div className="main-content">
        <div className="control-panel">
          <h2>Global Controls</h2>
          <div className="button-group">
            <button onClick={handleStart} className="btn btn-start">Start All</button>
            <button onClick={handleStop} className="btn btn-stop">Stop All</button>
            <button onClick={fetchStatus} className="btn btn-status">Refresh Status</button>
          </div>
          
          <h3>Individual Services</h3>
          <div className="service-controls">
            {Object.entries(serviceStatus).map(([service, status]) => (
              <div key={service} className="service-control">
                <span className="service-name">{service}</span>
                <span className={`service-status ${status.status}`}>{status.status}</span>
                <div className="service-buttons">
                  <button 
                    onClick={() => handleServiceStart(service)}
                    className="btn btn-service-start"
                    disabled={status.status === 'starting' || status.status === 'running'}
                  >
                    Start
                  </button>
                  <button 
                    onClick={() => handleServiceStop(service)}
                    className="btn btn-service-stop"
                    disabled={status.status === 'stopping' || status.status === 'stopped'}
                  >
                    Stop
                  </button>
                  <button 
                    onClick={() => handleServiceStatus(service)}
                    className="btn btn-service-status"
                  >
                    Refresh
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="command-section">
          <h2>Command Input</h2>
          <form onSubmit={handleSubmit} className="command-form">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Enter JUDO CLI command..."
              className="command-input"
            />
            <button type="submit" className="btn btn-execute">Execute</button>
          </form>

          <div className="command-output">
            <h3>Command Results</h3>
            {commands.map((cmd, index) => (
              <div key={index} className={`command-result ${cmd.success ? 'success' : 'error'}`}>
                <strong>$ {cmd.command}</strong>
                <pre>{cmd.output}</pre>
              </div>
            ))}
          </div>
        </div>

        <div className="log-section">
          <h2>Real-time Logs</h2>
          <div className="log-filter">
            <span>Filter: </span>
            <button 
              className={logFilter === 'all' ? 'btn btn-filter active' : 'btn btn-filter'}
              onClick={() => setLogFilter('all')}
            >
              All
            </button>
            <button 
              className={logFilter === 'karaf' ? 'btn btn-filter active' : 'btn btn-filter'}
              onClick={() => setLogFilter('karaf')}
            >
              Karaf
            </button>
            <button 
              className={logFilter === 'postgresql' ? 'btn btn-filter active' : 'btn btn-filter'}
              onClick={() => setLogFilter('postgresql')}
            >
              PostgreSQL
            </button>
            <button 
              className={logFilter === 'keycloak' ? 'btn btn-filter active' : 'btn btn-filter'}
              onClick={() => setLogFilter('keycloak')}
            >
              Keycloak
            </button>
          </div>
          <div className="log-viewer">
            {logs.map((log, index) => (
              <div key={index} className="log-entry">
                {log}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;