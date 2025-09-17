import React, { useState, useEffect, useRef, useCallback } from 'react';
import axios from 'axios';
import { XTerm } from 'react-xtermjs';


import './App.css';

interface ServiceStatus {
  service: string;
  status: string;
  timestamp: string;
}

interface LogMessage {
  ts: string;
  service: string;
  line: string;
}

interface SessionMessage {
  type: string;
  data?: string;
  state?: string;
  exitCode?: number;
  cols?: number;
  rows?: number;
  action?: string;
  welcome?: string;
}

function App() {
  const [activeTerminal, setActiveTerminal] = useState<'A' | 'B'>('A');

  const [terminalASource, setTerminalASource] = useState<string>('combined');
  const [serviceStatus, setServiceStatus] = useState<{[key: string]: ServiceStatus}>({});
  const [isServicePanelOpen, setIsServicePanelOpen] = useState(false);
  const [loadingServices, setLoadingServices] = useState<{[key: string]: boolean}>({});
  const [isProjectInitialized, setIsProjectInitialized] = useState<boolean | null>(null);
  const [showInitModal, setShowInitModal] = useState(false);
  
  const terminalARef = useRef<XTerm | null>(null);
  const terminalBRef = useRef<XTerm | null>(null);
  
  const logWs = useRef<WebSocket | null>(null);
  const sessionWs = useRef<WebSocket | null>(null);
  const isSessionRunning = useRef(false);
  const inputBufferRef = useRef<string>('');

  const getApiBaseUrl = useCallback(() => {
    const { protocol, hostname, port } = window.location;
    return `${protocol}//${hostname}:${port}`;
  }, []);

  const getWsBaseUrl = useCallback(() => {
    const { protocol, hostname, port } = window.location;
    const wsProtocol = protocol === 'https:' ? 'wss:' : 'ws:';
    return `${wsProtocol}//${hostname}:${port}`;
  }, []);


  const connectLogWebSocket = useCallback((source: string) => {
    if (logWs.current) {
      logWs.current.onclose = null;
      logWs.current.close();
    }

    const wsUrl = source === 'combined' 
      ? `${getWsBaseUrl()}/ws/logs/combined`
      : `${getWsBaseUrl()}/ws/logs/service/${source}`;

    const ws = new WebSocket(wsUrl);
    logWs.current = ws;
    
    ws.onopen = () => {
      console.log('Log WebSocket connected');
      if (terminalARef.current?.terminal) {
        terminalARef.current.terminal.write('\r\n\x1b[32m✓ Connected to log stream\x1b[0m\r\n');
      }
    };
    
    ws.onmessage = (event) => {
      try {
        const logMessage: LogMessage = JSON.parse(event.data);
        if (terminalARef.current?.terminal) {
          const serviceColor = {
            karaf: '\x1b[35m',
            postgresql: '\x1b[36m',
            keycloak: '\x1b[33m',
            combined: '\x1b[37m'
          }[logMessage.service] || '\x1b[37m';
          
          terminalARef.current.terminal.write(
            `${serviceColor}[${logMessage.service.toUpperCase()}]\x1b[0m ${logMessage.line}\r\n`
          );
        }
      } catch (error) {
        console.error('Error parsing log message:', error);
      }
    };
    
      
    ws.onclose = () => {
      console.log('Log WebSocket disconnected');
      if (terminalARef.current?.terminal) {
        terminalARef.current.terminal.write('\r\n\x1b[31m✗ Log stream disconnected\x1b[0m\r\n');
      }
    };
    
    ws.onerror = (error) => {
      console.error('Log WebSocket error:', error);
    };
  }, [getWsBaseUrl]);

  const connectSessionWebSocket = useCallback(() => {
    if (sessionWs.current) {
      sessionWs.current.onclose = null;
      sessionWs.current.close();
    }

    const ws = new WebSocket(`${getWsBaseUrl()}/ws/session`);
    sessionWs.current = ws;
    
    ws.onopen = () => {
      console.log('Session WebSocket connected');
      isSessionRunning.current = true;
      // Send initial terminal size
      if (terminalBRef.current?.terminal) {
        ws.send(JSON.stringify({ type: 'resize', cols: terminalBRef.current.terminal.cols, rows: terminalBRef.current.terminal.rows }));
      }
    };
    
    ws.onmessage = (event) => {
      try {
        const message: SessionMessage = JSON.parse(event.data);
        if (terminalBRef.current?.terminal) {
          switch (message.type) {
            case 'handshake':
              if (message.welcome) {
                terminalBRef.current.terminal.write(message.welcome);
              }
              break;
            case 'output':
              terminalBRef.current.terminal.write(message.data || '');
              break;
            case 'status':
              if (message.state === 'exited') {
                terminalBRef.current.terminal.write(`\r\n\x1b[31mSession exited with code ${message.exitCode}\x1b[0m\r\n`);
                isSessionRunning.current = false;
              }
              break;
            case 'prompt':
              terminalBRef.current.terminal.write(message.data || '');
              break;
          }
        }
      } catch (error) {
        console.error('Error parsing session message:', error);
      }
    };
    
    ws.onclose = () => {
      console.log('Session WebSocket disconnected');
      if (isSessionRunning.current) {
        isSessionRunning.current = false;
        if (terminalBRef.current?.terminal) {
          terminalBRef.current.terminal.write('\r\n\x1b[31m✗ Session disconnected. Reconnecting...\x1b[0m\r\n');
        }
        setTimeout(connectSessionWebSocket, 2000);
      } else {
        if (terminalBRef.current?.terminal) {
          terminalBRef.current.terminal.write('\r\n\x1b[31m✗ Session disconnected\x1b[0m\r\n');
        }
      }
    };
    
    ws.onerror = (error) => {
      console.error('Session WebSocket error:', error);
    };
  }, [getWsBaseUrl]);

  const handleTerminalBInput = (data: string) => {
    if (sessionWs.current && sessionWs.current.readyState === WebSocket.OPEN) {
      const message: SessionMessage = {
        type: 'input',
        data: data
      };
      sessionWs.current.send(JSON.stringify(message));
    }
  };

  const checkProjectInitialized = useCallback(async () => {
    try {
      const response = await axios.get(`${getApiBaseUrl()}/api/project/init/status`);
      setIsProjectInitialized(response.data.initialized);
      if (!response.data.initialized) {
        setShowInitModal(true);
      }
    } catch (error) {
      console.error('Failed to check project initialization status:', error);
      setIsProjectInitialized(false); // Assume not initialized if check fails
      setShowInitModal(true);
    }
  }, [getApiBaseUrl]);

  const handleProjectInit = async (initialize: boolean) => {
    setShowInitModal(false);
    if (initialize) {
      // Initialize project
      try {
        await axios.post(`${getApiBaseUrl()}/api/commands/judo%20init`);
        // Recheck initialization status
        await checkProjectInitialized();
      } catch (error) {
        console.error('Failed to initialize project:', error);
        // Still allow access but show warning
        setIsProjectInitialized(false);
      }
    } else {
      // User declined initialization, allow access but disable certain features
      setIsProjectInitialized(false);
    }
  };

  const fetchServiceStatuses = useCallback(async () => {
    try {
      // Use concurrent status endpoint for better performance
      const response = await axios.get(`${getApiBaseUrl()}/api/services/status`);
      const statuses = response.data;
      
      const statusMap: {[key: string]: ServiceStatus} = {};
      statuses.forEach((status: ServiceStatus) => {
        statusMap[status.service] = status;
      });
      
      setServiceStatus(statusMap);
    } catch (error) {
      console.error('Failed to fetch service statuses:', error);
      // Fallback to individual endpoints
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
      } catch (fallbackError) {
        console.error('Fallback status fetch also failed:', fallbackError);
      }
    }
  }, [getApiBaseUrl]);

  const handleServiceStart = async (service: string) => {
    setLoadingServices(prev => ({ ...prev, [service]: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/start`);
      startStatusPolling(service);
    } catch (error) {
      console.error(`Failed to start ${service}:`, error);
      setLoadingServices(prev => ({ ...prev, [service]: false }));
    }
  };

  const handleServiceStop = async (service: string) => {
    setLoadingServices(prev => ({ ...prev, [service]: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/stop`);
      startStatusPolling(service);
    } catch (error) {
      console.error(`Failed to stop ${service}:`, error);
      setLoadingServices(prev => ({ ...prev, [service]: false }));
    }
  };

  const handleAllServicesStart = async () => {
    setLoadingServices(prev => ({ ...prev, all: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/start`);
      startStatusPolling('all');
    } catch (error) {
      console.error('Failed to start all services:', error);
      setLoadingServices(prev => ({ ...prev, all: false }));
    }
  };

  const handleAllServicesStop = async () => {
    setLoadingServices(prev => ({ ...prev, all: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/stop`);
      startStatusPolling('all');
    } catch (error) {
      console.error('Failed to stop all services:', error);
      setLoadingServices(prev => ({ ...prev, all: false }));
    }
  };

  const startStatusPolling = (service: string) => {
    const pollInterval = setInterval(() => {
      fetchServiceStatuses();
    }, 2000);
    
    setTimeout(() => {
      clearInterval(pollInterval);
      setLoadingServices(prev => ({ ...prev, [service]: false }));
    }, 30000);
  };

  useEffect(() => {
    // Connect to log WebSocket
    connectLogWebSocket(terminalASource);

    // Fetch service statuses
    fetchServiceStatuses();

    // Check if project is initialized
    checkProjectInitialized();

    return () => {
      if (logWs.current) {
        logWs.current.onclose = null;
        logWs.current.close();
      }
      if (sessionWs.current) {
        sessionWs.current.onclose = null;
        sessionWs.current.close();
      }
    };
  }, []);

  useEffect(() => {
    // Connect to session WebSocket when Terminal B becomes active
    if (activeTerminal === 'B' && !isSessionRunning.current) {
      connectSessionWebSocket();
    }
  }, [activeTerminal, connectSessionWebSocket]);

  useEffect(() => {
    // Reconnect log WebSocket when source changes
    connectLogWebSocket(terminalASource);
  }, [terminalASource, connectLogWebSocket]);

  return (
    <div className={`App ${isServicePanelOpen ? 'service-panel-open' : ''}`}>
      <header className="App-header">
        <button 
          className="btn btn-service-panel"
          onClick={() => setIsServicePanelOpen(!isServicePanelOpen)}
        >
          {isServicePanelOpen ? '◀' : '▶'} Services
        </button>
        
        <h1>JUDO CLI Server</h1>
        
        <div className="terminal-switcher">
          <button 
            className={activeTerminal === 'A' ? 'btn btn-terminal active' : 'btn btn-terminal'}
            onClick={() => setActiveTerminal('A')}
          >
            Logs
          </button>
          <button 
            className={activeTerminal === 'B' ? 'btn btn-terminal active' : 'btn btn-terminal'}
            onClick={() => setActiveTerminal('B')}
          >
            JUDO Terminal
          </button>
        </div>
        
        {activeTerminal === 'A' && (
          <div className="terminal-a-controls">
            <span>Source: </span>
            <select 
              value={terminalASource} 
              onChange={(e) => setTerminalASource(e.target.value)}
              className="source-selector"
            >
              <option value="combined">Combined</option>
              <option value="karaf">Karaf</option>
              <option value="postgresql">PostgreSQL</option>
              <option value="keycloak">Keycloak</option>
            </select>
          </div>
        )}
      </header>

      <div className="main-content">
        <div className={`service-panel ${isServicePanelOpen ? 'open' : ''}`}>
          <h2>Services</h2>
          <div className="service-controls">
            {/* Parallel controls */}
            <div className="service-control parallel-controls">
              <span className="service-name">All Services</span>
              <div className="service-buttons">
                <button 
                  onClick={handleAllServicesStart}
                  className="btn btn-service-start"
                  disabled={loadingServices.all}
                >
                  {loadingServices.all ? 'Starting All...' : 'Start All'}
                </button>
                <button 
                  onClick={handleAllServicesStop}
                  className="btn btn-service-stop"
                  disabled={loadingServices.all}
                >
                  {loadingServices.all ? 'Stopping All...' : 'Stop All'}
                </button>
              </div>
            </div>
            
            {/* Individual service controls */}
            {Object.entries(serviceStatus).map(([service, status]) => (
              <div key={service} className="service-control">
                <span className="service-name">{service}</span>
                <span className={`service-status ${status.status}`}>{status.status}</span>
                <div className="service-buttons">
                  <button 
                    onClick={() => handleServiceStart(service)}
                    className="btn btn-service-start"
                    disabled={status.status === 'starting' || status.status === 'running' || loadingServices[service] || loadingServices.all}
                  >
                    {loadingServices[service] ? 'Starting...' : 'Start'}
                  </button>
                  <button 
                    onClick={() => handleServiceStop(service)}
                    className="btn btn-service-stop"
                    disabled={status.status === 'stopping' || status.status === 'stopped' || loadingServices[service] || loadingServices.all}
                  >
                    {loadingServices[service] ? 'Stopping...' : 'Stop'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="terminal-container">
          <XTerm 
            ref={terminalARef} 
            className={`terminal terminal-a ${activeTerminal === 'A' ? 'active' : 'hidden'} ${isProjectInitialized === false ? 'disabled' : ''}`}
          />
          <XTerm 
            ref={terminalBRef} 
            className={`terminal terminal-b ${activeTerminal === 'B' ? 'active' : 'hidden'} ${isProjectInitialized === false ? 'disabled' : ''}`}
            onData={handleTerminalBInput}
          />
        </div>
      </div>

      {/* Project Initialization Modal */}
      {showInitModal && (
        <div className="modal-overlay">
          <div className="modal">
            <h2>Project Not Initialized</h2>
            <p>This directory does not appear to be a JUDO project. Would you like to initialize it?</p>
            <div className="modal-buttons">
              <button 
                className="btn btn-service-start"
                onClick={() => handleProjectInit(true)}
              >
                Yes, Initialize
              </button>
              <button 
                className="btn btn-service-stop"
                onClick={() => handleProjectInit(false)}
              >
                No, Continue Anyway
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;