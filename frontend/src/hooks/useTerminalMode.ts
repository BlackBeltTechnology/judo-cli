import { useCallback } from 'react';
import { TerminalMode } from '../contexts/TerminalModeContext';

interface UseTerminalModeProps {
  onModeChange?: (newMode: TerminalMode, oldMode: TerminalMode) => void;
  onTerminalClear?: (mode: TerminalMode) => void;
}

export const useTerminalMode = ({
  onModeChange,
  onTerminalClear
}: UseTerminalModeProps = {}) => {
  const handleModeChange = useCallback((newMode: TerminalMode, oldMode: TerminalMode) => {
    // Clear terminal when switching modes
    if (onTerminalClear) {
      onTerminalClear(oldMode);
    }
    
    // Notify parent component of mode change
    if (onModeChange) {
      onModeChange(newMode, oldMode);
    }
  }, [onModeChange, onTerminalClear]);

  return {
    handleModeChange
  };
};