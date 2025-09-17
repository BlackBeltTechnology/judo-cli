// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';

// Setup canvas mocking for xterm.js
import 'jest-canvas-mock';

// Mock matchMedia for xterm.js
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock ResizeObserver for xterm.js
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Mock additional browser APIs for xterm.js
Object.defineProperty(global, 'requestAnimationFrame', {
  writable: true,
  value: jest.fn().mockImplementation((callback) => {
    setTimeout(callback, 0);
    return 1;
  }),
});

Object.defineProperty(global, 'cancelAnimationFrame', {
  writable: true,
  value: jest.fn(),
});

// Mock devicePixelRatio for xterm.js
Object.defineProperty(global, 'devicePixelRatio', {
  writable: true,
  value: 1,
});

// Mock window.getComputedStyle for xterm.js
Object.defineProperty(window, 'getComputedStyle', {
  writable: true,
  value: jest.fn().mockImplementation(() => ({
    getPropertyValue: jest.fn().mockReturnValue(''),
  })),
});

// Mock xterm.js Terminal to avoid browser API issues
jest.mock('@xterm/xterm', () => {
  const mockLoadAddon = jest.fn();
  const mockOpen = jest.fn();
  const mockWrite = jest.fn();
  const mockOnData = jest.fn().mockReturnValue({ dispose: jest.fn() });
  const mockDispose = jest.fn();
  const mockResize = jest.fn();
  
  return {
    Terminal: jest.fn().mockImplementation(() => ({
      loadAddon: mockLoadAddon,
      open: mockOpen,
      write: mockWrite,
      onData: mockOnData,
      dispose: mockDispose,
      resize: mockResize,
      cols: 80,
      rows: 24,
    })),
  };
});

// Mock xterm addons
jest.mock('@xterm/addon-fit', () => {
  const mockFit = jest.fn();
  return {
    FitAddon: jest.fn().mockImplementation(() => ({
      fit: mockFit,
    })),
  };
});

jest.mock('@xterm/addon-web-links', () => {
  return {
    WebLinksAddon: jest.fn(),
  };
});

// Mock axios to avoid ESM issues
jest.mock('axios', () => ({
  __esModule: true,
  default: jest.fn(() => Promise.resolve({ data: {} })),
  get: jest.fn(() => Promise.resolve({ data: {} })),
  post: jest.fn(() => Promise.resolve({ data: {} })),
  create: jest.fn(() => ({
    get: jest.fn(() => Promise.resolve({ data: {} })),
    post: jest.fn(() => Promise.resolve({ data: {} })),
  })),
  defaults: {
    baseURL: '',
  },
}));
