const mockLoadAddon = jest.fn();
const mockOpen = jest.fn();
const mockWrite = jest.fn();
const mockOnData = jest.fn().mockReturnValue({ dispose: jest.fn() });
const mockDispose = jest.fn();
const mockResize = jest.fn();

const Terminal = jest.fn().mockImplementation(() => ({
  loadAddon: mockLoadAddon,
  open: mockOpen,
  write: mockWrite,
  onData: mockOnData,
  dispose: mockDispose,
  resize: mockResize,
  cols: 80,
  rows: 24,
}));

module.exports = { Terminal };