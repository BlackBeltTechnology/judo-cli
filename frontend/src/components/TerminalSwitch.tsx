import React from 'react';

interface TerminalSwitchProps {
  currentMode: 'logs' | 'judo-terminal';
  onModeChange: (mode: 'logs' | 'judo-terminal') => void;
  disabled?: boolean;
}

const TerminalSwitch: React.FC<TerminalSwitchProps> = ({
  currentMode,
  onModeChange,
  disabled = false
}) => {
  const handleModeChange = (mode: 'logs' | 'judo-terminal') => {
    if (!disabled) {
      onModeChange(mode);
    }
  };

  return (
    <div className="terminal-switcher">
      <button
        className={`btn btn-terminal ${currentMode === 'logs' ? 'active' : ''} ${disabled ? 'disabled' : ''}`}
        onClick={() => handleModeChange('logs')}
        disabled={disabled}
        title={disabled ? 'Project not initialized' : 'Switch to Logs mode'}
      >
        Logs
      </button>
      <button
        className={`btn btn-terminal ${currentMode === 'judo-terminal' ? 'active' : ''} ${disabled ? 'disabled' : ''}`}
        onClick={() => handleModeChange('judo-terminal')}
        disabled={disabled}
        title={disabled ? 'Project not initialized' : 'Switch to JUDO Terminal mode'}
      >
        JUDO Terminal
      </button>
    </div>
  );
};

export default TerminalSwitch;