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
- [ ] T006 [P] Create basic UI components for dual log terminals (A/B), an A/B switch, and a collapsible left-side Service panel with status buttons.

## Phase 3.3: Core Implementation (Backend)
- [ ] T007 Implement the `GET /api/status` endpoint.
- [ ] T008 Implement the `POST /api/actions/start` endpoint.
- [ ] T009 Implement the `POST /api/actions/stop` endpoint.
- [ ] T010 Implement the WebSocket endpoint for log streaming.
- [ ] T011 Implement combined log stream multiplexer (Karaf, PostgreSQL, Keycloak) with clear source labeling.
- [ ] T012 [P] Implement `/ws/logs/combined` WebSocket endpoint for combined streaming.
- [ ] T013 [P] Implement `/ws/logs/service/{name}` WebSocket endpoint for single-service streaming.
- [ ] T014 [P] Implement `/ws/session` WebSocket endpoint for interactive `judo session` with PTY, input/output streaming, resize, and interrupt handling.
- [ ] T015 Implement log tailing adapters for `postgresql`, `karaf`, and `keycloak` with container/process restart handling.
- [ ] T016 Implement service-specific status endpoints (`GET /api/services/karaf/status`, `GET /api/services/postgresql/status`, `GET /api/services/keycloak/status`).
- [ ] T017 Implement service-specific control endpoints (`POST /api/services/karaf/start`, `POST /api/services/karaf/stop`, etc.).

## Phase 3.4: Core Implementation (Frontend)
- [ ] T018 Implement API calls to the backend for status and actions.
- [ ] T019 Implement a WebSocket client to receive and display logs in Xterm.js terminals with message batching.
- [ ] T020 Implement Terminal A/B switch and Terminal A’s source selector (Combined, Karaf, PostgreSQL, Keycloak).
- [ ] T021 Implement individual service status display and control UI.
- [ ] T022 Implement UI state persistence (active terminal, Terminal A source, Service panel open/closed).
- [ ] T023 Add per-terminal controls: Clear, Pause/Resume, Copy selection; show connection status (connecting, disconnected).
- [ ] T031 Implement Terminal B interactive Xterm client for `/ws/session` with input handling, paste, and auto-resize.
- [ ] T032 Implement Terminal B features: command history, Ctrl+C interrupt, session ended state with “Start New Session” action.

## Phase 3.5: Integration
- [ ] T024 [P] Add file embedding for the React frontend into the Go binary.
- [ ] T025 [P] Configure the Go server to serve the embedded frontend.
- [ ] T026 [P] Add a build script to `package.json` to build the frontend for production.

## Phase 3.6: Polish
- [ ] T027 [P] Add error handling to the frontend and backend.
- [ ] T028 [P] Style the frontend to be user-friendly with service-specific visual indicators.
- [ ] T029 [P] Add unit tests for the backend service management functionality.
- [ ] T030 [P] Add UI tests for service filtering, interactive session behavior, reconnect and resize.

## Dependencies
- Backend setup (T001-T004) before backend implementation (T007-T017).
- Frontend setup (T005-T006) before frontend implementation (T018-T023).
- Service management implementation (T015-T017) before frontend service controls (T020-T021).
- Core implementation (T007-T023) before integration (T024-T026).
- Integration (T024-T026) before polish (T027-T030).
