import React from "react";

interface ServiceStatus {
  service: string;
  status: string;
  timestamp: string;
}

interface ServicePanelProps {
  isOpen: boolean;
  serviceStatus: { [key: string]: ServiceStatus };
  loadingServices: { [key: string]: boolean };
  onStartService: (service: string) => void;
  onStopService: (service: string) => void;
  onStartAllServices: () => void;
  onStopAllServices: () => void;
}

const ServicePanel: React.FC<ServicePanelProps> = ({
  isOpen,
  serviceStatus,
  loadingServices,
  onStartService,
  onStopService,
  onStartAllServices,
  onStopAllServices,
}) => {
  return (
    <div className={`service-panel ${isOpen ? "open" : ""}`}>
      <h2>Services</h2>
      <div className="service-controls">
        {/* Parallel controls */}
        <div className="service-control parallel-controls">
          <span className="service-name">All Services</span>
          <div className="service-buttons">
            <button
              onClick={onStartAllServices}
              className="btn btn-service-start"
              disabled={loadingServices.all}
            >
              {loadingServices.all ? "Starting All..." : "Start All"}
            </button>
            <button
              onClick={onStopAllServices}
              className="btn btn-service-stop"
              disabled={loadingServices.all}
            >
              {loadingServices.all ? "Stopping All..." : "Stop All"}
            </button>
          </div>
        </div>

        {/* Individual service controls */}
        {Object.entries(serviceStatus).map(([service, status]) => (
          <div key={service} className="service-control">
            <span className="service-name">{service}</span>
            <span className={`service-status ${status.status}`}>
              {status.status}
            </span>
            <div className="service-buttons">
              <button
                onClick={() => onStartService(service)}
                className="btn btn-service-start"
                disabled={
                  status.status === "starting" ||
                  status.status === "running" ||
                  loadingServices[service] ||
                  loadingServices.all
                }
              >
                {loadingServices[service] ? "Starting..." : "Start"}
              </button>
              <button
                onClick={() => onStopService(service)}
                className="btn btn-service-stop"
                disabled={
                  status.status === "stopping" ||
                  status.status === "stopped" ||
                  loadingServices[service] ||
                  loadingServices.all
                }
              >
                {loadingServices[service] ? "Stopping..." : "Stop"}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ServicePanel;
