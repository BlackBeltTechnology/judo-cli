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

function App() {
  const [input, setInput] = useState('');
  const [commands, setCommands] = useState<Command[]>([]);
  const [status, setStatus] = useState<Status>({ status: 'unknown', timestamp: '' });
  const [logs, setLogs] = useState<string[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    fetchStatus();
    connectWebSocket();
    
    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);

  const connectWebSocket = () => {
    ws.current = new WebSocket('ws://localhost:8080/ws/logs');
    
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
      const response = await axios.get('http://localhost:8080/api/status');
      setStatus(response.data);
    } catch (error) {
      console.error('Failed to fetch status:', error);
    };
  };

  const handleStart = async () => {
    try {
      await axios.post('http://localhost:8080/api/actions/start');
      fetchStatus();
    } catch (error) {
      console.error('Failed to start:', error);
    };
  };

  const handleStop = async () => {
    try {
      await axios.post('http://localhost:8080/api/actions/stop');
      fetchStatus();
    } catch (error) {
      console.error('Failed to stop:', error);
    };
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    try {
      const response = await axios.post(`http://localhost:8080/api/commands/${encodeURIComponent(input)}`);
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
          <h2>Controls</h2>
          <div className="button-group">
            <button onClick={handleStart} className="btn btn-start">Start</button>
            <button onClick={handleStop} className="btn btn-stop">Stop</button>
            <button onClick={fetchStatus} className="btn btn-status">Refresh Status</button>
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