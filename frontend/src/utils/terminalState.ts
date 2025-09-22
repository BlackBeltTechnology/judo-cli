export interface TerminalBufferState {
  inputBuffer: string;
  cursorPosition: number;
  mode: 'logs' | 'judo-terminal';
  timestamp: number;
}

const BUFFER_STORAGE_KEY = 'judo-terminal-buffer-state';

export const saveBufferState = (state: Partial<TerminalBufferState>): void => {
  try {
    const currentState = getBufferState();
    const newState = {
      ...currentState,
      ...state,
      timestamp: Date.now()
    };
    localStorage.setItem(BUFFER_STORAGE_KEY, JSON.stringify(newState));
  } catch (error) {
    console.warn('Failed to save buffer state:', error);
  }
};

export const getBufferState = (): TerminalBufferState => {
  try {
    const stored = localStorage.getItem(BUFFER_STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch (error) {
    console.warn('Failed to load buffer state:', error);
  }
  
  // Return default state
  return {
    inputBuffer: '',
    cursorPosition: 0,
    mode: 'logs',
    timestamp: Date.now()
  };
};

export const clearBufferState = (): void => {
  try {
    localStorage.removeItem(BUFFER_STORAGE_KEY);
  } catch (error) {
    console.warn('Failed to clear buffer state:', error);
  }
};

export const preserveInputBuffer = (
  buffer: string,
  cursorPos: number,
  mode: 'logs' | 'judo-terminal'
): void => {
  saveBufferState({
    inputBuffer: buffer,
    cursorPosition: cursorPos,
    mode
  });
};

export const restoreInputBuffer = (): { buffer: string; cursorPos: number } => {
  const state = getBufferState();
  return {
    buffer: state.inputBuffer || '',
    cursorPos: state.cursorPosition || 0
  };
};

export const clearInputBuffer = (): void => {
  saveBufferState({
    inputBuffer: '',
    cursorPosition: 0
  });
};