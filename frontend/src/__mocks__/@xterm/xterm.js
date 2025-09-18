const { vi } = require('vitest');

module.exports = {
  Terminal: vi.fn().mockImplementation(() => {
    return {
      loadAddon: vi.fn(),
      open: vi.fn(),
      write: vi.fn(),
      onData: vi.fn().mockReturnValue({ dispose: vi.fn() }),
      dispose: vi.fn(),
      resize: vi.fn(),
      cols: 80,
      rows: 24,
    };
  }),
};