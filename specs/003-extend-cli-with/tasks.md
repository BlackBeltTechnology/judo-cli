# Tasks: Browser-Based Interactive CLI Server

**Input**: Design documents from `/specs/003-extend-cli-with/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/

## Phase 3.1: Backend Setup
- [ ] T001 [P] Create `internal/server` package and initial `server.go` file.
- [ ] T002 [P] Add a new `server` command to `cmd/judo/main.go`.
- [ ] T003 [P] Set up a basic HTTP server using `net/http`.
- [ ] T004 [P] Add routing for API endpoints and WebSocket.

## Phase 3.2: Frontend Setup
- [ ] T005 [P] Set up a new React application in the `frontend` directory using Create React App.
- [ ] T006 [P] Create basic UI components for the command input, output display, log viewer, and status buttons.

## Phase 3.3: Core Implementation (Backend)
- [ ] T007 Implement the `GET /api/status` endpoint.
- [ ] T008 Implement the `POST /api/actions/start` endpoint.
- [ ] T009 Implement the `POST /api/actions/stop` endpoint.
- [ ] T010 Implement the WebSocket endpoint for log streaming.
- [ ] T011 Implement the `POST /api/commands/{command}` endpoint to execute commands using `os/exec`.
- [ ] T012 [P] Implement handlers for all session commands (`help`, `exit`, `quit`, `clear`, `history`, `status`, `doctor`).
- [ ] T013 [P] Implement handlers for all project commands (`init`, `build`, `start`, `stop`, `status`, `clean`, `generate`, `dump`, `import`, `update`, `prune`, `reckless`, `self-update`).
- [ ] T015 Implement the `GET /api/logs/{service}` endpoint to tail individual service logs from `postgresql`, `karaf`, and `keycloak`.
- [ ] T016 Implement service-specific status endpoints (`GET /api/services/karaf/status`, `GET /api/services/postgresql/status`, `GET /api/services/keycloak/status`).
- [ ] T017 Implement service-specific control endpoints (`POST /api/services/karaf/start`, `POST /api/services/karaf/stop`, etc.).

## Phase 3.4: Core Implementation (Frontend)
- [ ] T018 Implement API calls to the backend for status and actions.
- [ ] T019 Implement a WebSocket client to receive and display logs with service filtering.
- [ ] T020 Implement service-specific log filtering UI (Karaf-only, PostgreSQL-only, Keycloak-only, combined).
- [ ] T021 Implement individual service status display and control UI.
- [ ] T022 Implement the command input to send commands to the backend.
- [ ] T023 Display command output received from the backend.

## Phase 3.5: Integration
- [ ] T024 [P] Add file embedding for the React frontend into the Go binary.
- [ ] T025 [P] Configure the Go server to serve the embedded frontend.
- [ ] T026 [P] Add a build script to `package.json` to build the frontend for production.

## Phase 3.6: Polish
- [ ] T027 [P] Add error handling to the frontend and backend.
- [ ] T028 [P] Style the frontend to be user-friendly with service-specific visual indicators.
- [ ] T029 [P] Add unit tests for the backend service management functionality.
- [ ] T030 [P] Add UI tests for service filtering and individual service controls.

## Dependencies
- Backend setup (T001-T004) before backend implementation (T007-T017).
- Frontend setup (T005-T006) before frontend implementation (T018-T023).
- Service management implementation (T015-T017) before frontend service controls (T020-T021).
- Core implementation (T007-T023) before integration (T024-T026).
- Integration (T024-T026) before polish (T027-T030).
