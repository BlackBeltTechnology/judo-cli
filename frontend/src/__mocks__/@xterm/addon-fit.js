const mockFit = jest.fn();

const FitAddon = jest.fn().mockImplementation(() => ({
  fit: mockFit,
}));

module.exports = { FitAddon };