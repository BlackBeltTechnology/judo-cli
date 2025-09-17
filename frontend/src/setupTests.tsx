import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

// Mock CSS imports
vi.mock('xterm/css/xterm.css', () => ({}));

// Mock react-xtermjs globally
vi.mock('react-xtermjs', () => ({
  XTerm: vi.fn(() => <div data-testid="mock-xterm" />),
}));

vi.mock('@xterm/addon-fit', () => ({
  FitAddon: vi.fn(() => ({
    fit: vi.fn(),
  })),
}));

// Mock WebSocket
export const mockWebSocket = vi.fn(() => ({
  send: vi.fn(),
  close: vi.fn(),
  readyState: WebSocket.OPEN,
  onopen: null,
  onmessage: null,
  onclose: null,
  onerror: null,
}));

Object.defineProperty(window, 'WebSocket', {
  writable: true,
  value: mockWebSocket,
});

global.self = global.window;

// Mock CloseEvent
class MockCloseEvent extends Event {
  readonly code: number;
  readonly reason: string;
  readonly wasClean: boolean;

  constructor(type: string, eventInitDict?: CloseEventInit) {
    super(type, eventInitDict);
    this.code = eventInitDict?.code || 0;
    this.reason = eventInitDict?.reason || '';
    this.wasClean = eventInitDict?.wasClean || false;
  }
}

Object.defineProperty(window, 'CloseEvent', {
  writable: true,
  value: MockCloseEvent,
});

// Mock matchMedia for xterm.js
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock ResizeObserver for xterm.js
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock additional browser APIs for xterm.js
Object.defineProperty(global, 'requestAnimationFrame', {
  writable: true,
  value: vi.fn().mockImplementation((callback) => {
    setTimeout(callback, 0);
    return 1;
  }),
});

Object.defineProperty(global, 'cancelAnimationFrame', {
  writable: true,
  value: vi.fn(),
});

// Mock devicePixelRatio for xterm.js
Object.defineProperty(global, 'devicePixelRatio', {
  writable: true,
  value: 1,
});

// Mock window.getComputedStyle for xterm.js
Object.defineProperty(window, 'getComputedStyle', {
  writable: true,
  value: vi.fn().mockImplementation(() => ({
    getPropertyValue: vi.fn().mockReturnValue(''),
  })),
});

// Mock HTMLCanvasElement for xterm.js
Object.defineProperty(global, 'HTMLCanvasElement', {
  writable: true,
  value: vi.fn().mockImplementation(() => ({
    getContext: vi.fn().mockReturnValue({
      fillRect: vi.fn(),
      clearRect: vi.fn(),
      getImageData: vi.fn().mockReturnValue({ data: new Uint8ClampedArray() }),
      putImageData: vi.fn(),
      createImageData: vi.fn().mockReturnValue({ data: new Uint8ClampedArray() }),
      setTransform: vi.fn(),
      drawImage: vi.fn(),
      save: vi.fn(),
      fillText: vi.fn(),
      restore: vi.fn(),
      beginPath: vi.fn(),
      moveTo: vi.fn(),
      lineTo: vi.fn(),
      closePath: vi.fn(),
      stroke: vi.fn(),
      translate: vi.fn(),
      scale: vi.fn(),
      rotate: vi.fn(),
      arc: vi.fn(),
      fill: vi.fn(),
      measureText: vi.fn().mockReturnValue({ width: 0 }),
      transform: vi.fn(),
      rect: vi.fn(),
      clip: vi.fn(),
    }),
    width: 0,
    height: 0,
    style: {},
  })),
});
