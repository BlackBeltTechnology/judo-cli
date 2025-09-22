import React from "react";
import { XTerm } from "react-xtermjs";
import { Terminal } from "xterm";

interface TerminalContainerProps {
  mode: string;
  isProjectInitialized: boolean | null;
  terminalARef: React.RefObject<HTMLDivElement>;
  terminalBRef: React.RefObject<HTMLDivElement>;
  terminalOptions: any;
}

const TerminalContainer: React.FC<TerminalContainerProps> = ({
  mode,
  isProjectInitialized,
  terminalARef,
  terminalBRef,
  terminalOptions,
}) => {
  return (
    <div className="terminal-container">
      <XTerm
        ref={terminalARef}
        className={`terminal terminal-a ${mode === 'logs' ? 'active' : 'hidden'}`}
        options={terminalOptions}
      />
      <XTerm
        ref={terminalBRef}
        className={`terminal terminal-b ${mode === 'judo-terminal' ? 'active' : 'hidden'}`}
        options={terminalOptions}
      />
    </div>
  );
};

export default TerminalContainer;
