import React, { createContext, useContext, useState, ReactNode } from 'react';

export type TerminalMode = 'logs' | 'judo-terminal';

interface TerminalModeContextType {
  mode: TerminalMode;
  setMode: (mode: TerminalMode) => void;
  isJudoTerminalActive: boolean;
}

const TerminalModeContext = createContext<TerminalModeContextType | undefined>(undefined);

interface TerminalModeProviderProps {
  children: ReactNode;
  initialMode?: TerminalMode;
}

export const TerminalModeProvider: React.FC<TerminalModeProviderProps> = ({
  children,
  initialMode = 'logs'
}) => {
  const [mode, setMode] = useState<TerminalMode>(initialMode);

  const value: TerminalModeContextType = {
    mode,
    setMode,
    isJudoTerminalActive: mode === 'judo-terminal'
  };

  return (
    <TerminalModeContext.Provider value={value}>
      {children}
    </TerminalModeContext.Provider>
  );
};

export const useTerminalMode = (): TerminalModeContextType => {
  const context = useContext(TerminalModeContext);
  if (context === undefined) {
    throw new Error('useTerminalMode must be used within a TerminalModeProvider');
  }
  return context;
};