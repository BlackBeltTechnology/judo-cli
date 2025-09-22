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
- [ ] T019 Implement a WebSocket client to receive and display logs in react-xtermjs terminals with message batching.
- [ ] T020 Implement Terminal A/B switch and Terminal A’s source selector (Combined, Karaf, PostgreSQL, Keycloak).
- [ ] T021 Implement individual service status display and control UI.
- [ ] T022 Implement UI state persistence (active terminal, Terminal A source, Service panel open/closed).
- [ ] T023 Add per-terminal controls: Clear, Pause/Resume, Copy selection; show connection status (connecting, disconnected).
- [ ] T031 Implement Terminal B interactive react-xtermjs client for `/ws/session` with input handling, paste, and auto-resize.
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
- Frontend implementation (T018-T023) before frontend testing (T058-T074).
- JUDO Terminal implementation (T101-T114) before E2E testing (T115-T144).
- Integration testing (T075-T100) should run after all implementation is complete.

## Phase 3.7: Amendments (UI/Init/TTY)
- [ ] T033 [P] Rename UI labels to 'Logs' and 'JUDO Terminal' across frontend and docs.
- [ ] T034 [P] Move Services toggle to a left-edge control; remove header button.
- [ ] T035 [P] Add GET `/api/project/init/status` endpoint returning initialized state.
- [ ] T036 [P] Add POST `/api/project/init` endpoint to trigger initialization.
- [ ] T037 [P] Frontend init gate: on load check init state; modal prompt (Yes/No); disable terminals until initialized.
- [ ] T038 UI: Show non-blocking notice when user declines init; include 'Initialize now' action.
- [ ] T039 [P] Session handshake: client `init` (term, cols, rows) and server PTY config; remove any client-side prompt injection.
- [ ] T040 [P] PTY parity: ensure native 'judo session' behavior (Ctrl+C, resize, history, prompt) with no UI-added formatting.
- [ ] T041 Tests: E2E for init flow, labels, Services toggle; integration tests for new endpoints.
- [ ] T042 Docs: Update spec, plan, quickstart to reflect labels and init flow.

### Additional Dependencies
- Init endpoints (T035–T036) before frontend init gate (T037–T038).
- Session handshake and PTY parity (T039–T040) before parity tests (T041).

## Phase 3.8: Test-Model Integration
- [ ] T043 [P] Document `test-model/` usage in README and developer docs (scope: generate, build, start, stop, dump, import, export).
- [ ] T044 Ensure `judo server` works when launched from `test-model/` and reflects service state correctly in the UI.
- [ ] T045 Add E2E test scripts to run against `test-model/` (build → start → server smoke-test → stop/clean).
- [ ] T046 Add DB workflow tests in `test-model/` (dump → import → verify tables present).
- [ ] T047 Update quickstart/examples with `test-model/` flows for reproducible demos.
- [ ] T048 Add cleanup guidance for `test-model/` (stop, clean, prune) to ensure repeatability.
- [ ] T049 Validate log sources map to Karaf/PostgreSQL/Keycloak in `test-model/` including Combined stream labels.
- [ ] T050 Establish CI job matrix step that exercises `test-model/` flows on supported platforms.

## Phase 3.9: Documentation & CI Alignment
- [ ] T051 [P] Update internal CLI help and docs pages for `server` (internal/help, docs/commands/, docs/getting-started.md)
- [ ] T052 [P] Update README sections referencing server and interactive session
- [ ] T053 [P] Review `.github/workflows/build.yml` for frontend build/embed steps; adjust if impacted
- [ ] T054 [P] Verify `.github/workflows/hugo.yml` remains correct; update if documentation structure changed
- [ ] T055 [P] Ensure release process includes embedded frontend assets (GoReleaser/build pipeline)
- [ ] T056 [P] Validate docs site builds and deploys after changes (local + CI)
- [ ] T057 [P] Add notes to `RELEASE.md` about server feature and embedding steps

## Phase 3.10: Comprehensive Vite/Vitest Frontend Testing Implementation

### Phase 3.10.1: Test Infrastructure & Setup
- [ ] T058 [P] Update vitest.config.ts for comprehensive testing (coverage, UI mode, watch)
- [ ] T059 [P] Enhance setupTests.tsx with global test utilities and custom render
- [ ] T060 [P] Create test data factories for WebSocket, terminal, and session data

### Phase 3.10.2: Update Existing Tests to Constitution Standards
- [ ] T061 [P] Update Terminal.test.tsx for behavior-driven testing with full state coverage
- [ ] T062 [P] Update WebSocket.test.tsx for realistic mocking and type-safe API responses
- [ ] T063 [P] Update App.test.tsx for comprehensive application state testing

### Phase 3.10.3: Add Missing Test Coverage
- [ ] T064 [P] Create component tests for all React components (ConnectionStatus, TerminalContainer, ErrorBoundary, LoadingSpinner)
- [ ] T065 [P] Create service tests for WebSocket, session, terminal, and error handling services
- [ ] T066 [P] Create hook tests for useWebSocket, useTerminal, and useSession custom hooks

### Phase 3.10.4: Behavior-Driven Testing Scenarios
- [ ] T067 [P] Implement user journey integration tests (connection flow, terminal interaction, error recovery)
- [ ] T068 [P] Create comprehensive edge case tests (network failures, invalid input, memory leaks)
- [ ] T069 [P] Implement visual regression testing with vitest-image-snapshot

### Phase 3.10.5: Accessibility & Performance Testing
- [ ] T070 [P] Implement comprehensive accessibility testing (screen readers, keyboard nav, ARIA, contrast)
- [ ] T071 [P] Add performance benchmarking (render speed, WebSocket processing, memory usage)

### Phase 3.10.6: CI Integration & Quality Gates
- [ ] T072 [P] Update CI workflows for frontend testing with coverage thresholds
- [ ] T073 [P] Implement quality gates (90% coverage, headless execution, accessibility compliance)
- [ ] T074 [P] Configure visual regression and performance budget enforcement in CI

### Testing Standards Compliance (Constitution Articles VIII-IX)
- ✅ Behavior-first approach focusing on user behavior, not implementation
- ✅ Realistic mock data matching actual API response structures
- ✅ Comprehensive state coverage (loading, success, error, edge cases)
- ✅ All mocks in vi.hoisted() blocks with type safety
- ✅ Component lifecycle, user interaction, and permission-based rendering tests
- ✅ Cross-browser compatibility and accessibility compliance
- ✅ Visual regression detection and performance monitoring

## Phase 3.11: JUDO Terminal Interactive Session Implementation

### Phase 3.11.1: Frontend Mode Switching
- [ ] T075 [P] Implement JUDO Terminal button component in frontend/src/components/TerminalSwitch.tsx
- [ ] T076 [P] Add terminal mode state management using React context in frontend/src/contexts/TerminalModeContext.tsx
- [ ] T077 [P] Implement mode switching logic with terminal clearing in frontend/src/hooks/useTerminalMode.ts
- [ ] T078 [P] Add visual indicators for current mode (Logs vs JUDO Terminal) in frontend/src/components/TerminalHeader.tsx

### Phase 3.11.2: Session State Preservation
- [ ] T079 [P] Implement session state storage during mode switches in frontend/src/services/sessionState.ts
- [ ] T080 [P] Add command history preservation using localStorage in frontend/src/hooks/useCommandHistory.ts
- [ ] T081 [P] Create buffer preservation for partial commands in frontend/src/utils/terminalState.ts
- [ ] T082 [P] Implement scroll position preservation across mode switches

### Phase 3.11.3: Terminal Control Features
- [ ] T083 [P] Add resize event handling for interactive sessions in frontend/src/hooks/useTerminalResize.ts
- [ ] T084 [P] Implement control character handling (Ctrl+C, Ctrl+D) in frontend/src/services/terminalInput.ts
- [ ] T085 [P] Add real-time command output streaming through WebSocket in frontend/src/services/websocketSession.ts
- [ ] T086 [P] Implement session interruption and restart functionality in frontend/src/components/JUDOTerminal.tsx

### Phase 3.11.4: Backend Session Management
- [ ] T087 [P] Enhance WebSocket session handler for state preservation in internal/server/websocket_session.go
- [ ] T088 [P] Add session state storage across connections in internal/server/session_manager.go
- [ ] T089 [P] Implement command execution with proper PTY handling in internal/server/command_executor.go
- [ ] T090 [P] Add control character translation to signals in internal/server/signal_handler.go

### Phase 3.11.5: Integration Testing
- [ ] T091 [P] Create integration tests for mode switching in frontend/src/__tests__/terminalMode.test.tsx
- [ ] T092 [P] Implement tests for session state preservation in frontend/src/__tests__/sessionState.test.tsx
- [ ] T093 [P] Add tests for command history functionality in frontend/src/__tests__/commandHistory.test.tsx
- [ ] T094 [P] Create tests for control character handling in frontend/src/__tests__/terminalInput.test.tsx
- [ ] T095 [P] Implement performance tests for real-time command execution in tests/performance/terminal_perf_test.go
- [ ] T096 [P] Add E2E tests for JUDO Terminal functionality in tests/e2e/judo_terminal.test.js

### Phase 3.11.6: Parity Validation
- [ ] T097 [P] Create validation tests ensuring JUDO Terminal matches native session behavior in tests/integration/terminal_parity_test.go
- [ ] T098 [P] Implement command output comparison tests in tests/validation/command_output_test.go
- [ ] T099 [P] Add control character behavior validation tests in tests/validation/control_chars_test.go
- [ ] T100 [P] Create session state consistency tests across mode switches in tests/integration/session_state_test.go

## Phase 3.12: JUDO Terminal Interactive Session Implementation (Continued)
- [ ] T101 [P] Implement JUDO Terminal button in frontend UI for mode switching
- [ ] T102 [P] Add terminal mode state management (logs vs interactive session)
- [ ] T103 [P] Implement terminal clearing when switching modes
- [ ] T104 [P] Add session state preservation during mode switches
- [ ] T105 [P] Implement command history preservation in JUDO Terminal
- [ ] T106 [P] Add terminal resize event handling for interactive sessions
- [ ] T107 [P] Implement control character handling (Ctrl+C, Ctrl+D) in JUDO Terminal
- [ ] T108 [P] Add real-time command output streaming through WebSocket
- [ ] T109 [P] Implement session interruption and restart functionality
- [ ] T110 [P] Create integration tests for JUDO Terminal mode switching
- [ ] T111 [P] Implement tests for session state preservation
- [ ] T112 [P] Add tests for command history functionality
- [ ] T113 [P] Create tests for control character handling
- [ ] T114 [P] Implement performance tests for real-time command execution

## Phase 3.12: Comprehensive E2E Testing Implementation
- [ ] T115 [P] Set up E2E testing framework with Playwright or similar tool for browser automation
- [ ] T116 [P] Create test environment setup and teardown utilities for `test-model/`
- [ ] T117 [P] Implement CLI server startup and browser launch sequence tests
- [ ] T118 [P] Create tests for embedded frontend asset serving and proper rendering
- [ ] T119 [P] Implement WebSocket connection establishment and validation tests
- [ ] T120 [P] Create comprehensive service lifecycle management tests (start, stop, status)
- [ ] T121 [P] Implement real log streaming validation tests for all services
- [ ] T122 [P] Create JUDO Terminal command execution parity tests vs native CLI
- [ ] T123 [P] Implement project initialization flow tests from uninitialized to operational
- [ ] T124 [P] Create database operation tests (dump, import, export) through JUDO Terminal
- [ ] T125 [P] Implement service status indicator synchronization tests with actual service state
- [ ] T126 [P] Create error scenario tests including service failures and network issues
- [ ] T127 [P] Implement system recovery and consistency maintenance tests
- [ ] T128 [P] Create port configuration and conflict handling tests
- [ ] T129 [P] Implement authentication and security validation tests if applicable
- [ ] T130 [P] Create comprehensive user interaction to system outcome validation tests
- [ ] T131 [P] Implement performance benchmarking tests for critical user journeys
- [ ] T132 [P] Create resource cleanup and proper shutdown procedure tests
- [ ] T133 [P] Implement cross-platform compatibility tests on supported OS
- [ ] T134 [P] Configure CI/CD pipeline integration for E2E test execution
- [ ] T135 [P] Create detailed logging and debugging infrastructure for E2E test failures
- [ ] T136 [P] Implement stability measures and waiting strategies to prevent test flakiness
- [ ] T137 [P] Ensure E2E tests cover all acceptance scenarios from the specification
- [ ] T138 [P] Create test data management and cleanup procedures for reproducible results
- [ ] T139 [P] Implement test parallelization and optimization for faster execution
- [ ] T140 [P] Create test reporting and visualization for easy results interpretation
- [ ] T141 [P] Implement JUDO Terminal specific E2E tests for mode switching and session preservation
- [ ] T142 [P] Create tests for interactive terminal functionality identical to native session
- [ ] T143 [P] Implement tests for real-time command execution through WebSocket
- [ ] T144 [P] Add tests for terminal resize events and control character handling

## Phase 3.13: Browser-Based Interactive CLI Server (Updated)
- [x] T145 Implement the `judo server` command and basic HTTP server.
- [x] T136 Set up the initial React frontend with a dual-terminal UI.
- [x] T137 Implement WebSocket endpoints for log and session streaming.
- [x] T138 Implement API endpoints for service status and control.
- [x] T139 **Fix:** Resolve WebSocket race conditions for log and session terminals on initial load.
- [x] T140 **Fix:** Implement a handshake protocol for log history requests.
- [x] T141 **Fix:** Correct backend to send `\r\n` for all terminal output.
- [x] T142 **Fix:** Update all command outputs in `api.go` to use `\r\n`.

## Phase 3.14: Refactoring & UX Improvements (Completed)
- [x] T143 Refactor `App.tsx` into smaller components (`AppHeader`, `ServicePanel`, `TerminalContainer`, `ProjectInitModal`).
- [x] T144 Extract repetitive terminal resizing logic into a reusable `fitTerminal` function.
- [x] T145 Implement robust, imperative input handling for the JUDO Terminal.
- [x] T146 Remove CSS that incorrectly disabled the terminal.
- [x] T147 Align the web session's look and feel with the native `judo session` (banner, help, status).

## Phase 3.15: Comprehensive Testing & Documentation (Next Steps)
- [ ] T148 Write comprehensive unit and integration tests for the backend server, including WebSocket handlers and API endpoints.
- [ ] T149 Write comprehensive unit and integration tests for the frontend, including component behavior, state management, and WebSocket interactions.
- [ ] T150 Write end-to-end tests using Playwright to cover the full user flow, including service management, log streaming, and interactive command execution.
- [ ] T151 Update the project's `README.md` and `TESTING.md` to document the new server functionality and testing strategy.
- [ ] T152 Create user-facing documentation for the `judo server` command and its features.



---

*Constitution v2.4.1 Compliance: Frontend testing tasks (T058-T074) fully implement Articles VIII-IX requirements for behavior-driven testing, realistic mocking, comprehensive coverage, accessibility, and performance testing. JUDO Terminal tasks (T101-T114) implement the interactive session functionality specified in the spec extension.*
