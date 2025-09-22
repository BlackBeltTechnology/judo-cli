import React, { useState, useEffect, useRef, useCallback } from "react";
import axios from "axios";
import { XTerm, useXTerm } from "react-xtermjs";
import { FitAddon } from "@xterm/addon-fit";

import { useTerminalMode } from "./contexts/TerminalModeContext";
import TerminalSwitch from "./components/TerminalSwitch";
import TerminalHeader from "./components/TerminalHeader";
import { useTerminalMode as useTerminalModeHook } from "./hooks/useTerminalMode";
import { useCommandHistory } from "./hooks/useCommandHistory";
import { preserveTerminalState } from "./services/sessionState";

import "./App.css";
import ServicePanel from "./components/ServicePanel";
import ProjectInitModal from "./components/ProjectInitModal";
import TerminalContainer from "./components/TerminalContainer";
import AppHeader from "./components/AppHeader";

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
  const { mode, setMode } = useTerminalMode();
  const [terminalSource, setTerminalSource] = useState<string>("combined");
  const [serviceStatus, setServiceStatus] = useState<{
    [key: string]: ServiceStatus;
  }>({});
  const [isServicePanelOpen, setIsServicePanelOpen] = useState(false);
  const [loadingServices, setLoadingServices] = useState<{
    [key: string]: boolean;
  }>({});
  const [isProjectInitialized, setIsProjectInitialized] = useState<boolean | null>(null);
  const [showInitModal, setShowInitModal] = useState(false);

  const { ref: terminalARef, instance: terminalAInstance } = useXTerm();
  const { ref: terminalBRef, instance: terminalBInstance } = useXTerm();
  const fitAddonA = useRef(new FitAddon());
  const fitAddonB = useRef(new FitAddon());
  const terminalAInstanceRef = useRef(terminalAInstance);
  const terminalBInstanceRef = useRef(terminalBInstance);
  const [isTerminalAReady, setIsTerminalAReady] = useState(false);
  const [isTerminalBReady, setIsTerminalBReady] = useState(false);

  const fitTerminal = (terminal: any, fitAddon: FitAddon) => {
    try {
      fitAddon.fit();
      if (terminal.cols < 120) {
        terminal.resize(120, terminal.rows);
      }
      if (terminal.rows < 40) {
        terminal.resize(terminal.cols, 40);
      }
      terminal.refresh(0, terminal.rows - 1);
    } catch (error) {
      console.error("Error fitting terminal:", error);
    }
  };

  useEffect(() => {
    terminalAInstanceRef.current = terminalAInstance;
    const wasReady = isTerminalAReady;
    const isReady = !!terminalAInstance;
    setIsTerminalAReady(isReady);

    if (terminalAInstance) {
      terminalAInstance.loadAddon(fitAddonA.current);
      setTimeout(() => fitTerminal(terminalAInstance, fitAddonA.current), 100);
    }
  }, [terminalAInstance, isTerminalAReady]);

  useEffect(() => {
    terminalBInstanceRef.current = terminalBInstance;
    const wasReady = isTerminalBReady;
    const isReady = !!terminalBInstance;
    setIsTerminalBReady(isReady);

    if (terminalBInstance) {
      terminalBInstance.loadAddon(fitAddonB.current);
      setTimeout(() => {
        fitTerminal(terminalBInstance, fitAddonB.current);
        if (!wasReady && isReady && mode === "judo-terminal") {
          console.log("Terminal B fully ready, connecting session WebSocket");
          connectSessionWebSocket();
        }
      }, 100);
    }
  }, [terminalBInstance, isTerminalBReady, mode]);

  const sessionWs = useRef<WebSocket | null>(null);

  const { handleModeChange } = useTerminalModeHook({
    onModeChange: (newMode, oldMode) => {
      if (oldMode === "judo-terminal" && terminalBInstanceRef.current) {
        preserveTerminalState(terminalBInstanceRef.current, oldMode);
      } else if (oldMode === "logs" && terminalAInstanceRef.current) {
        preserveTerminalState(terminalAInstanceRef.current, oldMode);
      }

      if (oldMode === "logs" && logWs.current) {
        logWs.current.onclose = null;
        logWs.current.close();
        logWs.current = null;
      } else if (oldMode === "judo-terminal" && sessionWs.current) {
        sessionWs.current.onclose = null;
        sessionWs.current.close();
        sessionWs.current = null;
      }

      if (newMode === "logs") {
        connectLogWebSocket(terminalSource);
      } else if (newMode === "judo-terminal") {
        connectSessionWebSocket();
      }
    },
    onTerminalClear: (mode) => {
      if (mode === "logs" && terminalAInstanceRef.current) {
        terminalAInstanceRef.current.clear();
      } else if (mode === "judo-terminal" && terminalBInstanceRef.current) {
        terminalBInstanceRef.current.clear();
      }
    },
  });

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
  const [connectionState, setConnectionState] = useState<{ logs: string; session: string }>({ logs: "disconnected", session: "disconnected" });
  const isConnectingRef = useRef<{ logs: boolean; session: boolean }>({ logs: false, session: false });
  const connectionAttemptRef = useRef<{ logs: number; session: number }>({ logs: 0, session: 0 });

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
        logWs.current = null;
      }

      if (!isTerminalAReady || mode !== "logs") {
        return;
      }

      if (logWs.current && (logWs.current.readyState === WebSocket.CONNECTING || logWs.current.readyState === WebSocket.OPEN)) {
        return;
      }

      if (isConnectingRef.current.logs) {
        return;
      }

      connectionAttemptRef.current.logs++;
      if (connectionAttemptRef.current.logs > 1) {
        return;
      }

      isConnectingRef.current.logs = true;

      const wsUrl = source === "combined" ? `${getWsBaseUrl()}/ws/logs/combined` : `${getWsBaseUrl()}/ws/logs/service/${source}`;

      try {
        const ws = new WebSocket(wsUrl);
        logWs.current = ws;

        ws.onopen = () => {
          setConnectionState((prev) => ({ ...prev, logs: "connected" }));
          isConnectingRef.current.logs = false;
          connectionAttemptRef.current.logs = 0;
        };

        ws.onmessage = (event) => {
          try {
            let data: string;
            if (typeof event.data === "string") {
              data = event.data;
            } else if (event.data instanceof ArrayBuffer) {
              const decoder = new TextDecoder();
              data = decoder.decode(event.data);
            } else if (event.data instanceof Blob) {
              const reader = new FileReader();
              reader.onload = () => {
                if (typeof reader.result === "string") {
                  handleWebSocketMessage(reader.result);
                }
              };
              reader.readAsText(event.data);
              return;
            } else {
              return;
            }

            const logMessage: LogMessage = JSON.parse(data);

            if (logMessage.line === "Log stream connected") {
              logWs.current?.send(JSON.stringify({ type: "request_history", source: terminalSource }));
              return;
            }

            if (!logMessage.line || logMessage.line.trim() === "") {
              return;
            }

            if (terminalAInstanceRef.current && mode === "logs") {
              const processedLine = logMessage.line;
              const formattedMessage = logMessage.service === "combined" ? `${processedLine}\r\n` : `[${new Date(logMessage.ts).toLocaleTimeString()}] [${logMessage.service.toUpperCase()}] ${processedLine}\r\n`;

              terminalAInstanceRef.current.write(formattedMessage);
            }
          } catch (error) {}
        };

        const handleWebSocketMessage = (data: string) => {
          try {
            const logMessage: LogMessage = JSON.parse(data);

            if (logMessage.line === "Log stream connected") {
              logWs.current?.send(JSON.stringify({ type: "request_history", source: terminalSource }));
              return;
            }

            if (!logMessage.line || logMessage.line.trim() === "") {
              return;
            }

            if (terminalAInstanceRef.current && mode === "logs") {
              const processedLine = logMessage.line;
              const formattedMessage = logMessage.service === "combined" ? `${processedLine}\r\n` : `[${new Date(logMessage.ts).toLocaleTimeString()}] [${logMessage.service.toUpperCase()}] ${processedLine}\r\n`;

              terminalAInstanceRef.current.write(formattedMessage);
            }
          } catch (error) {}
        };

        ws.onclose = (event) => {
          setConnectionState((prev) => ({ ...prev, logs: "disconnected" }));
          isConnectingRef.current.logs = false;
          connectionAttemptRef.current.logs = 0;
          if (event.code !== 1000 && mode === "logs") {
            setTimeout(() => {
              if (mode === "logs") {
                connectLogWebSocket(terminalSource);
              }
            }, 3000);
          }
        };

        ws.onerror = () => {
          setConnectionState((prev) => ({ ...prev, logs: "error" }));
          isConnectingRef.current.logs = false;
          connectionAttemptRef.current.logs = 0;
        };
      } catch (error) {
        setConnectionState((prev) => ({ ...prev, logs: "failed" }));
        isConnectingRef.current.logs = false;
        connectionAttemptRef.current.logs = 0;

        setTimeout(() => {
          if (mode === "logs") {
            connectLogWebSocket(terminalSource);
          }
        }, 3000);
      }
    },
    [getWsBaseUrl, mode, isTerminalAReady, terminalSource],
  );

  const connectSessionWebSocket = useCallback(() => {
    if (sessionWs.current) {
      sessionWs.current.onclose = null;
      sessionWs.current.close();
      sessionWs.current = null;
    }

    if (!isTerminalBReady || mode !== "judo-terminal") {
      return;
    }

    if (sessionWs.current && (sessionWs.current.readyState === WebSocket.CONNECTING || sessionWs.current.readyState === WebSocket.OPEN)) {
      return;
    }

    if (isConnectingRef.current.session) {
      return;
    }

    connectionAttemptRef.current.session++;
    if (connectionAttemptRef.current.session > 1) {
      return;
    }

    isConnectingRef.current.session = true;

    try {
      const ws = new WebSocket(`${getWsBaseUrl()}/ws/session`);
      sessionWs.current = ws;

      ws.onopen = () => {
        setConnectionState((prev) => ({ ...prev, session: "connected" }));
        isConnectingRef.current.session = false;
        connectionAttemptRef.current.session = 0;
        if (terminalBInstanceRef.current) {
          ws.send(JSON.stringify({ type: "resize", cols: terminalBInstanceRef.current.cols, rows: terminalBInstanceRef.current.rows }));
        }
      };

      ws.onmessage = (event) => {
        try {
          let data: string;
          if (typeof event.data === "string") {
            data = event.data;
          } else if (event.data instanceof ArrayBuffer) {
            const decoder = new TextDecoder();
            data = decoder.decode(event.data);
          } else if (event.data instanceof Blob) {
            const reader = new FileReader();
            reader.onload = () => {
              if (typeof reader.result === "string") {
                handleSessionWebSocketMessage(reader.result);
              }
            };
            reader.readAsText(event.data);
            return;
          } else {
            return;
          }

          const message = JSON.parse(data);
          if (terminalBInstanceRef.current && mode === "judo-terminal") {
            switch (message.type) {
              case "handshake":
                if (message.welcome) {
                  terminalBInstanceRef.current.write(message.welcome);
                }
                break;
              case "output":
                terminalBInstanceRef.current.write(message.data || "");
                break;
              case "status":
                if (message.state === "exited") {
                  terminalBInstanceRef.current.write(`\r\n\x1b[31mSession exited with code ${message.exitCode}\x1b[0m\r\n`);
                }
                break;
              case "prompt":
                terminalBInstanceRef.current.write(message.data || "");
                break;
            }
          }
        } catch (error) {}
      };

      const handleSessionWebSocketMessage = (data: string) => {
        try {
          const message = JSON.parse(data);
          if (terminalBInstanceRef.current && mode === "judo-terminal") {
            switch (message.type) {
              case "handshake":
                if (message.welcome) {
                  terminalBInstanceRef.current.write(message.welcome);
                }
                break;
              case "output":
                terminalBInstanceRef.current.write(message.data || "");
                break;
              case "status":
                if (message.state === "exited") {
                  terminalBInstanceRef.current.write(`\r\n\x1b[31mSession exited with code ${message.exitCode}\x1b[0m\r\n`);
                }
                break;
              case "prompt":
                terminalBInstanceRef.current.write(message.data || "");
                break;
            }
          }
        } catch (error) {}
      };

      ws.onclose = (event) => {
        setConnectionState((prev) => ({ ...prev, session: "disconnected" }));
        isConnectingRef.current.session = false;
        connectionAttemptRef.current.session = 0;
        if (terminalBInstanceRef.current && mode === "judo-terminal") {
          terminalBInstanceRef.current.write("\r\n\x1b[31m✗ Session disconnected\x1b[0m\r\n");
        }

        if (event.code !== 1000 && mode === "judo-terminal") {
          setTimeout(() => {
            if (mode === "judo-terminal") {
              connectSessionWebSocket();
            }
          }, 3000);
        }
      };

      ws.onerror = () => {
        setConnectionState((prev) => ({ ...prev, session: "error" }));
        isConnectingRef.current.session = false;
        connectionAttemptRef.current.session = 0;
      };
    } catch (error) {
      setConnectionState((prev) => ({ ...prev, session: "failed" }));
      isConnectingRef.current.session = false;
      connectionAttemptRef.current.session = 0;

      setTimeout(() => {
        if (mode === "judo-terminal") {
          connectSessionWebSocket();
        }
      }, 3000);
    }
  }, [getWsBaseUrl, mode, isTerminalBReady]);

  const [command, setCommand] = useState("");

  const handleTerminalBInput = useCallback(
    (data: string) => {
      const char = data;
      if (char === "\r") {
        // Enter key
        if (sessionWs.current && sessionWs.current.readyState === WebSocket.OPEN) {
          const message = {
            type: "input",
            data: command,
          };
          sessionWs.current.send(JSON.stringify(message));
        }
        setCommand("");
        terminalBInstance?.writeln("");
      } else if (char === "\u007f") {
        // Backspace
        if (command.length > 0) {
          setCommand(command.slice(0, -1));
          terminalBInstance?.write("\b \b");
        }
      } else {
        setCommand(command + char);
        terminalBInstance?.write(char);
      }
    },
    [command, terminalBInstance],
  );

  useEffect(() => {
    if (terminalBInstance) {
      const onDataDisposable = terminalBInstance.onData(handleTerminalBInput);
      return () => {
        onDataDisposable.dispose();
      };
    }
  }, [terminalBInstance, handleTerminalBInput]);

  const checkProjectInitialized = useCallback(async () => {
    try {
      const response = await axios.get(`${getApiBaseUrl()}/api/project/init/status`);
      setIsProjectInitialized(response.data.initialized);
      if (!response.data.initialized) {
        setShowInitModal(true);
      }
    } catch (error) {
      setIsProjectInitialized(false);
      setShowInitModal(true);
    }
  }, [getApiBaseUrl]);

  const handleProjectInit = async (initialize: boolean) => {
    setShowInitModal(false);
    if (initialize) {
      try {
        await axios.post(`${getApiBaseUrl()}/api/commands/judo%20init`);
        await checkProjectInitialized();
      } catch (error) {
        setIsProjectInitialized(false);
      }
    } else {
      setIsProjectInitialized(false);
    }
  };

  const fetchServiceStatuses = useCallback(async () => {
    try {
      const response = await axios.get(`${getApiBaseUrl()}/api/services/status`);
      const statuses = response.data;

      const statusMap: { [key: string]: ServiceStatus } = {};
      statuses.forEach((status: ServiceStatus) => {
        statusMap[status.service] = status;
      });

      setServiceStatus(statusMap);
    } catch (error) {
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
      } catch (fallbackError) {}
    }
  }, [getApiBaseUrl]);

  const handleServiceStart = async (service: string) => {
    setLoadingServices((prev) => ({ ...prev, [service]: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/start`);
      startStatusPolling(service);
    } catch (error) {
      setLoadingServices((prev) => ({ ...prev, [service]: false }));
    }
  };

  const handleServiceStop = async (service: string) => {
    setLoadingServices((prev) => ({ ...prev, [service]: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/${service}/stop`);
      startStatusPolling(service);
    } catch (error) {
      setLoadingServices((prev) => ({ ...prev, [service]: false }));
    }
  };

  const handleAllServicesStart = async () => {
    setLoadingServices((prev) => ({ ...prev, all: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/start`);
      startStatusPolling("all");
    } catch (error) {
      setLoadingServices((prev) => ({ ...prev, all: false }));
    }
  };

  const handleAllServicesStop = async () => {
    setLoadingServices((prev) => ({ ...prev, all: true }));
    try {
      await axios.post(`${getApiBaseUrl()}/api/services/stop`);
      startStatusPolling("all");
    } catch (error) {
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
    setTimeout(() => {
      if (mode === "logs" && terminalAInstance) {
        terminalAInstance.focus();
      } else if (mode === "judo-terminal" && terminalBInstance) {
        terminalBInstance.focus();
      }
    }, 100);
  }, [mode, terminalAInstance, terminalBInstance]);

  useEffect(() => {
    if (mode === "logs") {
      if (sessionWs.current) {
        sessionWs.current.close();
      }
    } else if (mode === "judo-terminal") {
      if (logWs.current) {
        logWs.current.close();
      }
    }
  }, [mode, connectLogWebSocket, connectSessionWebSocket]);

  useEffect(() => {
    fetchServiceStatuses();
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
    if (mode === "logs" && isTerminalAReady) {
      connectLogWebSocket(terminalSource);
    }
  }, [terminalSource, mode, connectLogWebSocket, isTerminalAReady]);

  useEffect(() => {
    if (mode === "judo-terminal" && isTerminalBReady) {
      connectSessionWebSocket();
    }
  }, [mode, isTerminalBReady, connectSessionWebSocket]);

  useEffect(() => {
    const handleResize = () => {
      if (mode === "logs" && terminalAInstanceRef.current && fitAddonA.current) {
        fitTerminal(terminalAInstanceRef.current, fitAddonA.current);
      } else if (mode === "judo-terminal" && terminalBInstanceRef.current && fitAddonB.current) {
        fitTerminal(terminalBInstanceRef.current, fitAddonB.current);
        if (sessionWs.current && sessionWs.current.readyState === WebSocket.OPEN) {
          sessionWs.current.send(
            JSON.stringify({
              type: "resize",
              cols: terminalBInstanceRef.current.cols,
              rows: terminalBInstanceRef.current.rows,
            }),
          );
        }
      }
    };

    window.addEventListener("resize", handleResize);
    setTimeout(handleResize, 200);

    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, [mode]);

  return (
    <div className={`App ${isServicePanelOpen ? "service-panel-open" : ""}`}>
      <AppHeader
        isServicePanelOpen={isServicePanelOpen}
        onToggleServicePanel={() => setIsServicePanelOpen(!isServicePanelOpen)}
        mode={mode}
        onModeChange={(newMode) => {
          const oldMode = mode;
          setMode(newMode);
          handleModeChange(newMode, oldMode);
        }}
        isProjectInitialized={isProjectInitialized}
        terminalSource={terminalSource}
        onSourceChange={setTerminalSource}
        connectionState={connectionState}
      />

      <div className="main-content">
        <ServicePanel
          isOpen={isServicePanelOpen}
          serviceStatus={serviceStatus}
          loadingServices={loadingServices}
          onStartService={handleServiceStart}
          onStopService={handleServiceStop}
          onStartAllServices={handleAllServicesStart}
          onStopAllServices={handleAllServicesStop}
        />

        <TerminalContainer mode={mode} isProjectInitialized={isProjectInitialized} terminalARef={terminalARef} terminalBRef={terminalBRef} terminalOptions={terminalOptions} onTerminalBInput={handleTerminalBInput} />
      </div>

      {showInitModal && <ProjectInitModal onInitialize={handleProjectInit} />}
    </div>
  );
}

export default App;
