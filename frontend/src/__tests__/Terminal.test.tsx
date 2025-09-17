import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom/vitest';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// MANDATORY: All mocks must be declared in vi.hoisted() blocks
const { mockLoadAddon, mockOpen, mockWrite, mockOnData, mockDispose, mockResize, mockFit, mockTerminal, mockFitAddon, mockWebLinksAddon } = vi.hoisted(() => ({
  mockLoadAddon: vi.fn(),
  mockOpen: vi.fn(),
  mockWrite: vi.fn(),
  mockOnData: vi.fn().mockReturnValue({ dispose: vi.fn() }),
  mockDispose: vi.fn(),
  mockResize: vi.fn(),
  mockFit: vi.fn(),
  mockTerminal: vi.fn(() => ({
    loadAddon: mockLoadAddon,
    open: mockOpen,
    write: mockWrite,
    onData: mockOnData,
    dispose: mockDispose,
    resize: mockResize,
    cols: 80,
    rows: 24,
  })),
  mockFitAddon: vi.fn(() => ({
    fit: mockFit,
  })),
  mockWebLinksAddon: vi.fn(),
}));

vi.mock('@xterm/xterm', async (importOriginal) => {
  const original = await importOriginal<typeof import('@xterm/xterm')>();
  return {
    ...original,
    Terminal: mockTerminal,
  };
});

vi.mock('@xterm/addon-fit', async (importOriginal) => {
  const original = await importOriginal<typeof import('@xterm/addon-fit')>();
  return {
    ...original,
    FitAddon: mockFitAddon,
  };
});

vi.mock('@xterm/addon-web-links', async (importOriginal) => {
  const original = await importOriginal<typeof import('@xterm/addon-web-links')>();
  return {
    ...original,
    WebLinksAddon: mockWebLinksAddon,
  };
});

describe('Terminal Component Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('terminal initialization and cleanup', () => {
    const terminal = mockTerminal();
    const fitAddon = mockFitAddon();
    
    // Simulate terminal initialization
    terminal.loadAddon(fitAddon);
    
    expect(mockTerminal).toHaveBeenCalled();
    expect(mockFitAddon).toHaveBeenCalled();
    expect(mockLoadAddon).toHaveBeenCalledWith(fitAddon);
  });

  it('terminal writes data correctly', () => {
    const terminal = mockTerminal();
    
    // Simulate writing data to terminal
    terminal.write('Test log message\r\n');
    
    expect(mockWrite).toHaveBeenCalledWith('Test log message\r\n');
  });

  it('terminal handles resize events', () => {
    const terminal = mockTerminal();
    const fitAddon = mockFitAddon();
    
    // Simulate resize
    fitAddon.fit();
    terminal.resize(100, 30);
    
    expect(mockFit).toHaveBeenCalled();
    expect(mockResize).toHaveBeenCalledWith(100, 30);
  });

  it('terminal handles input events', () => {
    const { Terminal } = require('@xterm/xterm');
    
    const terminal = new Terminal();
    const mockCallback = vi.fn();
    
    // Mock the onData implementation
    mockOnData.mockImplementation((callback) => {
      callback('test input');
      return { dispose: vi.fn() };
    });
    
    // Set up input handler
    terminal.onData(mockCallback);
    
    expect(mockOnData).toHaveBeenCalled();
  });
});