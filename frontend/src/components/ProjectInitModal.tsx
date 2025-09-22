import React from "react";

interface ProjectInitModalProps {
  onInitialize: (initialize: boolean) => void;
}

const ProjectInitModal: React.FC<ProjectInitModalProps> = ({ onInitialize }) => {
  return (
    <div className="modal-overlay">
      <div className="modal">
        <h2>Project Not Initialized</h2>
        <p>
          This directory does not appear to be a JUDO project. Would you
          like to initialize it?
        </p>
        <div className="modal-buttons">
          <button
            className="btn btn-service-start"
            onClick={() => onInitialize(true)}
          >
            Yes, Initialize
          </button>
          <button
            className="btn btn-service-stop"
            onClick={() => onInitialize(false)}
          >
            No, Continue Anyway
          </button>
        </div>
      </div>
    </div>
  );
};

export default ProjectInitModal;
