module.exports = {
  Terminal: jest.fn().mockImplementation(() => {
    return {
      loadAddon: jest.fn(),
      open: jest.fn(),
      write: jest.fn(),
      onData: jest.fn().mockReturnValue({ dispose: jest.fn() }),
      dispose: jest.fn(),
      resize: jest.fn(),
      cols: 80,
      rows: 24,
    };
  }),
};
