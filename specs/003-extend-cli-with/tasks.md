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
- [ ] T015 Implement the `GET /api/logs/{service}` endpoint to tail logs from `postgresql`, `karaf`, and `keycloak`.

## Phase 3.4: Core Implementation (Frontend)
- [ ] T013 Implement API calls to the backend for status and actions.
- [ ] T014 Implement a WebSocket client to receive and display logs.
- [ ] T015 Implement the command input to send commands to the backend.
- [ ] T016 Display command output received from the backend.

## Phase 3.5: Integration
- [ ] T017 [P] Add file embedding for the React frontend into the Go binary.
- [ ] T018 [P] Configure the Go server to serve the embedded frontend.
- [ ] T019 [P] Add a build script to `package.json` to build the frontend for production.

## Phase 3.6: Polish
- [ ] T020 [P] Add error handling to the frontend and backend.
- [ ] T021 [P] Style the frontend to be user-friendly.
- [ ] T022 [P] Add unit tests for the backend.
- [ ] T024 [P] Add UI tests for all frontend functionality.

## Dependencies
- Backend setup (T001-T004) before backend implementation (T007-T012).
- Frontend setup (T005-T006) before frontend implementation (T013-T016).
- Core implementation (T007-T016) before integration (T017-T019).
- Integration (T017-T019) before polish (T020-T023).
