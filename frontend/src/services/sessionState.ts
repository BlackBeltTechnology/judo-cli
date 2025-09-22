export interface SessionState {
  commandHistory: string[];
  inputBuffer: string;
  scrollPosition: number;
  cursorPosition: { x: number; y: number };
  lastActiveMode: 'logs' | 'judo-terminal';
  timestamp: number;
}

const SESSION_STORAGE_KEY = 'judo-terminal-session-state';

export const saveSessionState = (state: Partial<SessionState>): void => {
  try {
    const currentState = getSessionState();
    const newState = {
      ...currentState,
      ...state,
      timestamp: Date.now()
    };
    localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(newState));
  } catch (error) {
    console.warn('Failed to save session state:', error);
  }
};

export const getSessionState = (): SessionState => {
  try {
    const stored = localStorage.getItem(SESSION_STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch (error) {
    console.warn('Failed to load session state:', error);
  }
  
  // Return default state
  return {
    commandHistory: [],
    inputBuffer: '',
    scrollPosition: 0,
    cursorPosition: { x: 0, y: 0 },
    lastActiveMode: 'logs',
    timestamp: Date.now()
  };
};

export const clearSessionState = (): void => {
  try {
    localStorage.removeItem(SESSION_STORAGE_KEY);
  } catch (error) {
    console.warn('Failed to clear session state:', error);
  }
};

export const preserveTerminalState = (
  terminal: any,
  mode: 'logs' | 'judo-terminal'
): void => {
  if (!terminal) return;

  const state: Partial<SessionState> = {
    lastActiveMode: mode,
    scrollPosition: terminal.buffer.active.viewportY,
    cursorPosition: {
      x: terminal.buffer.active.cursorX,
      y: terminal.buffer.active.cursorY
    }
  };

  saveSessionState(state);
};

export const restoreTerminalState = (terminal: any): void => {
  if (!terminal) return;

  const state = getSessionState();
  
  // Restore scroll position
  if (state.scrollPosition !== undefined) {
    terminal.scrollToLine(state.scrollPosition);
  }
  
  // Restore cursor position
  if (state.cursorPosition) {
    terminal.write('\x1b[0;0H'); // Move to home position first
    terminal.write(`\x1b[${state.cursorPosition.y + 1};${state.cursorPosition.x + 1}H`);
  }
};