import React from "react";
import TerminalSwitch from "./TerminalSwitch";
import TerminalHeader from "./TerminalHeader";
import { TerminalMode } from "../contexts/TerminalModeContext";

interface AppHeaderProps {
  isServicePanelOpen: boolean;
  onToggleServicePanel: () => void;
  mode: TerminalMode;
  onModeChange: (newMode: TerminalMode) => void;
  isProjectInitialized: boolean | null;
  terminalSource: string;
  onSourceChange: (source: string) => void;
  connectionState: { logs: string; session: string };
}

const AppHeader: React.FC<AppHeaderProps> = ({
  isServicePanelOpen,
  onToggleServicePanel,
  mode,
  onModeChange,
  isProjectInitialized,
  terminalSource,
  onSourceChange,
  connectionState,
}) => {
  return (
    <header className="App-header">
      <button
        className="btn btn-service-panel"
        onClick={onToggleServicePanel}
      >
        {isServicePanelOpen ? "◀" : "▶"} Services
      </button>

      <h1>JUDO CLI Server</h1>

      <TerminalSwitch
        currentMode={mode}
        onModeChange={onModeChange}
        disabled={isProjectInitialized === false}
      />

      <TerminalHeader
        currentMode={mode}
        terminalASource={terminalSource}
        onSourceChange={onSourceChange}
        disabled={isProjectInitialized === false}
      />
      
      <div style={{fontSize: '12px', color: '#666', padding: '5px 10px', background: '#f0f0f0', borderRadius: '4px'}}>
        Connection Status: Logs: {connectionState.logs}, Session: {connectionState.session}
      </div>
    </header>
  );
};

export default AppHeader;
