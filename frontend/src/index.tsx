import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import { TerminalModeProvider } from './contexts/TerminalModeContext';

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <TerminalModeProvider>
      <App />
    </TerminalModeProvider>
  </React.StrictMode>
);
