import React, { useState, useEffect, useRef, useCallback } from "react";
import axios from "axios";
import { XTerm, useXTerm } from "react-xtermjs";
import { FitAddon } from "@xterm/addon-fit";

import "./App.css";

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

function App() {
  const [terminalSource, setTerminalSource] = useState<string>("combined");
  const [serviceStatus, setServiceStatus] = useState<{
    [key: string]: ServiceStatus;
  }>({});
  const [isServicePanelOpen, setIsServicePanelOpen] = useState(false);
  const [loadingServices, setLoadingServices] = useState<{
    [key: string]: boolean;
  }>({});
  const [isProjectInitialized, setIsProjectInitialized] = useState<
    boolean | null
  >(null);
  const [showInitModal, setShowInitModal] = useState(false);

  const { ref: terminalRef, instance: terminalInstance } = useXTerm();
  const fitAddon = useRef(new FitAddon());
  const terminalInstanceRef = useRef(terminalInstance);

  const terminalOptions = {
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
    theme: {
      background: "#0f1117",
      foreground: "#ffffff",
      cursor: "#ffffff",
      selection: "#ffffff40",
    },
    scrollback: 1000,
    allowTransparency: true,
    convertEol: true,
    cols: 120,
    rows: 40,
  };

  const logWs = useRef<WebSocket | null>(null);

  const getApiBaseUrl = useCallback(() => {
    const { protocol, hostname, port } = window.location;
    return `${protocol}//${hostname}:${port}`;
  }, []);

  const getWsBaseUrl = useCallback(() => {
    const { protocol, hostname, port } = window.location;
    const wsProtocol = protocol === "https:" ? "wss:" : "ws:";
    return `${wsProtocol}//${hostname}:${port}`;
  }, []);

  const connectLogWebSocket = useCallback(
    (source: string) => {
      if (logWs.current) {
        logWs.current.onclose = null;
        logWs.current.close();
      }

      const wsUrl =
        source === "combined"
          ? `${getWsBaseUrl()}/ws/logs/combined`
          : `${getWsBaseUrl()}/ws/logs/service/${source}`;

      const ws = new WebSocket(wsUrl);
      logWs.current = ws;

      ws.onopen = () => {
        // Connection established silently - no status message needed
        // Terminal instance check is handled by the initialization process
      };

      ws.onmessage = (event) => {
        try {
          const logMessage: LogMessage = JSON.parse(event.data);

          // Skip empty messages (heartbeats)
          if (!logMessage.line || logMessage.line.trim() === "") {
            return;
          }

          if (terminalInstanceRef.current) {
            // Preserve ANSI color codes but remove problematic control sequences
            const processedLine = logMessage.line;
            // .replace(/\x1b\[K/g, '') // Remove clear line codes (can cause display issues)
            // .replace(/\x1b\[2K/g, '') // Remove clear line codes
            // .replace(/\x1b\[0G/g, '') // Remove cursor position codes (can cause wrapping issues)
            // .replace(/\x1b\[\?.*[hl]/g, '') // Remove terminal mode codes
            // .replace(/\x1b\=.*[hl]/g, ''); // Remove additional terminal codes

            // For combined logs, don't add timestamp or service prefix since individual lines already have them
            const formattedMessage =
              logMessage.service === "combined"
                ? `${processedLine}\r\n`
                : `[${new Date(logMessage.ts).toLocaleTimeString()}] [${logMessage.service.toUpperCase()}] ${processedLine}\r\n`;
            
            console.log("Writing to terminal:", formattedMessage);
            terminalInstanceRef.current.write(formattedMessage);
          } else {
            console.log("Terminal instance not available for writing");
          }
          // Silently skip if terminal instance is not available - no warning needed
        } catch (error) {
          console.error("Error parsing log message:", error, event.data);
        }
      };

      ws.onclose = () => {
        // Connection closed silently - no status message needed
      };

      ws.onerror = (error) => {
        console.error("Log WebSocket error:", error);
      };
    },
    [getWsBaseUrl],
  );

  const checkProjectInitialized = useCallback(async () => {
    try {
      const response = await axios.get(
        `${getApiBaseUrl()}/api/project/init/status`,
      );
      setIsProjectInitialized(response.data.initialized);
      if (!response.data.initialized) {
        setShowInitModal(true);
      }
    } catch (error) {
      console.error("Failed to check project initialization status:", error);
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
        console.error("Failed to initialize project:", error);
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
      const response = await axios.get(
        `${getApiBaseUrl()}/api/services/status`,
      );
      const statuses = response.data;

      const statusMap: { [key: string]: ServiceStatus } = {};
      statuses.forEach((status: ServiceStatus) => {
        statusMap[status.service] = status;
      });

      setServiceStatus(statusMap);
    } catch (error) {
      console.error("Failed to fetch service statuses:", error);
      // Fallback to individual endpoints
      try {
        const [karaf, postgres, keycloak] = await Promise.all([
          axios.get(`${getApiBaseUrl()}/api/services/karaf/status`),
          axios.get(`${getApiBaseUrl()}/api/services/postgresql/status`),
          axios.get(`${getApiBaseUrl()}/api/services/keycloak/status`),
        ]);

        setServiceStatus({
          karaf: karaf.data,
          postgresql: postgres.data,
          keycloak: keycloak.data,
        });
      } catch (fallbackError) {
        console.error("Fallback status fetch also failed:", fallbackError);
      }
    }
  }, [getApiBaseUrl]);

  const handleServiceStart = async (service: string) => {
    setLoadingServices((prev) => ({ ...prev, [service]: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/start`);
      startStatusPolling(service);
    } catch (error) {
      console.error(`Failed to start ${service}:`, error);
      setLoadingServices((prev) => ({ ...prev, [service]: false }));
    }
  };

  const handleServiceStop = async (service: string) => {
    setLoadingServices((prev) => ({ ...prev, [service]: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/stop`);
      startStatusPolling(service);
    } catch (error) {
      console.error(`Failed to stop ${service}:`, error);
      setLoadingServices((prev) => ({ ...prev, [service]: false }));
    }
  };

  const handleAllServicesStart = async () => {
    setLoadingServices((prev) => ({ ...prev, all: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/start`);
      startStatusPolling("all");
    } catch (error) {
      console.error("Failed to start all services:", error);
      setLoadingServices((prev) => ({ ...prev, all: false }));
    }
  };

  const handleAllServicesStop = async () => {
    setLoadingServices((prev) => ({ ...prev, all: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/stop`);
      startStatusPolling("all");
    } catch (error) {
      console.error("Failed to stop all services:", error);
      setLoadingServices((prev) => ({ ...prev, all: false }));
    }
  };

  const startStatusPolling = (service: string) => {
    const pollInterval = setInterval(() => {
      fetchServiceStatuses();
    }, 2000);

    setTimeout(() => {
      clearInterval(pollInterval);
      setLoadingServices((prev) => ({ ...prev, [service]: false }));
    }, 30000);
  };

  useEffect(() => {
    // Handle window resize with debouncing
    let resizeTimeout: NodeJS.Timeout;
    const handleResize = () => {
      clearTimeout(resizeTimeout);
      resizeTimeout = setTimeout(() => {
        if (terminalInstance && fitAddon.current) {
          try {
            fitAddon.current.fit();
            // Ensure minimum dimensions after fitting
            if (terminalInstance.cols < 120) {
              terminalInstance.resize(120, terminalInstance.rows);
            }
            if (terminalInstance.rows < 24) {
              terminalInstance.resize(terminalInstance.cols, 24);
            }
            // Force a refresh to handle any rendering artifacts and wrapping issues
            terminalInstance.refresh(0, terminalInstance.rows - 1);
            
            // Additional check: if we're still at 80 columns, force resize
            if (terminalInstance.cols === 80) {
              console.warn("Terminal stuck at 80 columns, forcing resize to 120");
              terminalInstance.resize(120, terminalInstance.rows);
              terminalInstance.refresh(0, terminalInstance.rows - 1);
            }
          } catch (error) {
            console.warn("Could not fit terminal on resize:", error);
          }
        }
      }, 100);
    };

    window.addEventListener("resize", handleResize);

    // Fetch service statuses
    fetchServiceStatuses();

    // Check if project is initialized
    checkProjectInitialized();

    return () => {
      clearTimeout(resizeTimeout);
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  useEffect(() => {
    // Update terminal instance ref when it changes
    terminalInstanceRef.current = terminalInstance;

    // Initialize terminal when it becomes available
    if (terminalInstance && terminalRef.current) {
      console.log("Terminal instance available, initializing...");
      // Load addons
      terminalInstance.loadAddon(fitAddon.current);

      // Set up resize observer
      const resizeObserver = new ResizeObserver(() => {
        if (terminalInstance && fitAddon.current) {
          try {
            fitAddon.current.fit();
            // Ensure minimum dimensions after fitting
            if (terminalInstance.cols < 120) {
              terminalInstance.resize(120, terminalInstance.rows);
            }
            if (terminalInstance.rows < 24) {
              terminalInstance.resize(terminalInstance.cols, 24);
            }
            
            // Additional check: if we're still at 80 columns, force resize
            if (terminalInstance.cols === 80) {
              console.warn("Terminal stuck at 80 columns, forcing resize to 120");
              terminalInstance.resize(120, terminalInstance.rows);
              terminalInstance.refresh(0, terminalInstance.rows - 1);
            }
          } catch (error) {
            console.warn("Could not fit terminal on container resize:", error);
          }
        }
      });

      resizeObserver.observe(terminalRef.current);

      // Fit terminal to container with a small delay to ensure DOM is ready
      setTimeout(() => {
        try {
          console.log("Fitting terminal to container...");
          fitAddon.current.fit();
          console.log("Terminal fitted, cols:", terminalInstance.cols, "rows:", terminalInstance.rows);
          
          // Ensure minimum dimensions after fitting
          if (terminalInstance.cols < 120) {
            terminalInstance.resize(120, terminalInstance.rows);
            console.log("Resized terminal to 120 columns");
          }
          if (terminalInstance.rows < 24) {
            terminalInstance.resize(terminalInstance.cols, 24);
            console.log("Resized terminal to 24 rows");
          }
          
          // Force immediate refresh to handle any initial rendering issues
          terminalInstance.refresh(0, terminalInstance.rows - 1);
          console.log("Terminal refreshed with", terminalInstance.rows, "rows");
        } catch (error) {
          console.warn("Could not fit terminal:", error);
        }
      }, 150);

      // Connect to WebSocket after terminal is fully initialized with a small delay
      const connectTimeout = setTimeout(() => {
        console.log("Terminal initialized, connecting WebSocket...");
        connectLogWebSocket(terminalSource);
      }, 200);

      // Store the resize observer for cleanup
      return () => {
        resizeObserver.disconnect();
        clearTimeout(connectTimeout);
        if (logWs.current) {
          logWs.current.onclose = null;
          logWs.current.close();
          logWs.current = null;
        }
      };
    }
  }, [terminalInstance, terminalSource, connectLogWebSocket]);

  // WebSocket connection is now handled in the terminal initialization useEffect

  return (
    <div className={`App ${isServicePanelOpen ? "service-panel-open" : ""}`}>
      <header className="App-header">
        <button
          className="btn btn-service-panel"
          onClick={() => setIsServicePanelOpen(!isServicePanelOpen)}
        >
          {isServicePanelOpen ? "◀" : "▶"} Services
        </button>

        <h1>JUDO CLI Server</h1>

        <div className="terminal-controls">
          <span>Source: </span>
          <select
            value={terminalSource}
            onChange={(e) => setTerminalSource(e.target.value)}
            className="source-selector"
          >
            <option value="combined">Combined</option>
            <option value="karaf">Karaf</option>
            <option value="postgresql">PostgreSQL</option>
            <option value="keycloak">Keycloak</option>
          </select>


        </div>
      </header>

      <div className="main-content">
        <div className={`service-panel ${isServicePanelOpen ? "open" : ""}`}>
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
                  {loadingServices.all ? "Starting All..." : "Start All"}
                </button>
                <button
                  onClick={handleAllServicesStop}
                  className="btn btn-service-stop"
                  disabled={loadingServices.all}
                >
                  {loadingServices.all ? "Stopping All..." : "Stop All"}
                </button>
              </div>
            </div>

            {/* Individual service controls */}
            {Object.entries(serviceStatus).map(([service, status]) => (
              <div key={service} className="service-control">
                <span className="service-name">{service}</span>
                <span className={`service-status ${status.status}`}>
                  {status.status}
                </span>
                <div className="service-buttons">
                  <button
                    onClick={() => handleServiceStart(service)}
                    className="btn btn-service-start"
                    disabled={
                      status.status === "starting" ||
                      status.status === "running" ||
                      loadingServices[service] ||
                      loadingServices.all
                    }
                  >
                    {loadingServices[service] ? "Starting..." : "Start"}
                  </button>
                  <button
                    onClick={() => handleServiceStop(service)}
                    className="btn btn-service-stop"
                    disabled={
                      status.status === "stopping" ||
                      status.status === "stopped" ||
                      loadingServices[service] ||
                      loadingServices.all
                    }
                  >
                    {loadingServices[service] ? "Stopping..." : "Stop"}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="terminal-container">
          <XTerm
            ref={terminalRef}
            className={`terminal ${isProjectInitialized === false ? "disabled" : ""}`}
            options={terminalOptions}
          />
        </div>
      </div>

      {/* Project Initialization Modal */}
      {showInitModal && (
        <div className="modal-overlay">
          <div className="modal">
            <h2>Project Not Initialized</h2>
            <p>
              This directory does not appear to be a JUDO project. Would you
              like to initialize it?
            </p>
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
