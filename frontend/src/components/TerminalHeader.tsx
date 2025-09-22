import React from 'react';
import { TerminalMode } from '../contexts/TerminalModeContext';

interface TerminalHeaderProps {
  currentMode: TerminalMode;
  terminalASource: string;
  onSourceChange: (source: string) => void;
  disabled?: boolean;
}

const TerminalHeader: React.FC<TerminalHeaderProps> = ({
  currentMode,
  terminalASource,
  onSourceChange,
  disabled = false
}) => {
  const getModeIndicator = () => {
    switch (currentMode) {
      case 'logs':
        return (
          <div className="mode-indicator logs-mode">
            <span className="indicator-dot"></span>
            <span className="mode-text">Logs Mode</span>
            <span className="source-info">- Watching: {terminalASource}</span>
          </div>
        );
      case 'judo-terminal':
        return (
          <div className="mode-indicator judo-terminal-mode">
            <span className="indicator-dot interactive"></span>
            <span className="mode-text">JUDO Terminal Mode</span>
            <span className="source-info">- Interactive Session</span>
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="terminal-header">
      {getModeIndicator()}
      
      {currentMode === 'logs' && (
        <div className="terminal-a-controls">
          <span>Source: </span>
          <select 
            value={terminalASource} 
            onChange={(e) => onSourceChange(e.target.value)}
            className="source-selector"
            disabled={disabled}
          >
            <option value="combined">Combined</option>
            <option value="karaf">Karaf</option>
            <option value="postgresql">PostgreSQL</option>
            <option value="keycloak">Keycloak</option>
          </select>
        </div>
      )}
    </div>
  );
};

export default TerminalHeader;