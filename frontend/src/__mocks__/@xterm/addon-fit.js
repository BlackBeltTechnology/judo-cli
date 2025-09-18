const { vi } = require('vitest');

const mockFit = vi.fn();

const FitAddon = vi.fn().mockImplementation(() => ({
  fit: mockFit,
}));

module.exports = { FitAddon };