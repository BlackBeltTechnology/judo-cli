import { useState, useCallback, useEffect } from 'react';

const COMMAND_HISTORY_KEY = 'judo-terminal-command-history';
const MAX_HISTORY_LENGTH = 100;

export const useCommandHistory = () => {
  const [commandHistory, setCommandHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);

  // Load command history from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(COMMAND_HISTORY_KEY);
      if (stored) {
        const history = JSON.parse(stored);
        setCommandHistory(history);
      }
    } catch (error) {
      console.warn('Failed to load command history:', error);
    }
  }, []);

  const addToHistory = useCallback((command: string) => {
    if (!command.trim()) return;

    setCommandHistory(prev => {
      // Remove duplicates and add new command to beginning
      const newHistory = [
        command.trim(),
        ...prev.filter(cmd => cmd !== command.trim())
      ].slice(0, MAX_HISTORY_LENGTH);

      // Save to localStorage
      try {
        localStorage.setItem(COMMAND_HISTORY_KEY, JSON.stringify(newHistory));
      } catch (error) {
        console.warn('Failed to save command history:', error);
      }

      return newHistory;
    });
    
    setHistoryIndex(-1); // Reset history index after adding new command
  }, []);

  const getPreviousCommand = useCallback(() => {
    if (commandHistory.length === 0) return '';
    
    const newIndex = Math.min(historyIndex + 1, commandHistory.length - 1);
    setHistoryIndex(newIndex);
    return commandHistory[newIndex];
  }, [commandHistory, historyIndex]);

  const getNextCommand = useCallback(() => {
    if (commandHistory.length === 0 || historyIndex <= 0) {
      setHistoryIndex(-1);
      return '';
    }
    
    const newIndex = historyIndex - 1;
    setHistoryIndex(newIndex);
    return commandHistory[newIndex];
  }, [commandHistory, historyIndex]);

  const clearHistory = useCallback(() => {
    setCommandHistory([]);
    setHistoryIndex(-1);
    try {
      localStorage.removeItem(COMMAND_HISTORY_KEY);
    } catch (error) {
      console.warn('Failed to clear command history:', error);
    }
  }, []);

  const resetHistoryIndex = useCallback(() => {
    setHistoryIndex(-1);
  }, []);

  return {
    commandHistory,
    addToHistory,
    getPreviousCommand,
    getNextCommand,
    clearHistory,
    resetHistoryIndex
  };
};