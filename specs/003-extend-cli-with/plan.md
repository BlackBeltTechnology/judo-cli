# Implementation Plan: Browser-Based Interactive CLI Server

**Branch**: `003-extend-cli-with-a-server-function` | **Date**: 2025-09-15 | **Spec**: `/specs/003-extend-cli-with/spec.md`
**Input**: Feature specification from user description

## Summary
This plan outlines the implementation of a `server` command for the JUDO CLI that provides a browser-based UI focused on real-time service logs and an interactive session. The UI presents two terminals (A: logs with source selector; B: interactive `judo session`) with an A/B switch and a collapsible left-side Service panel. The frontend will be built with React, communicating with the Go backend via REST (status/actions) and WebSockets (logs and session). The compiled frontend will be embedded in the final Go binary.

## Technical Context
**Language/Version**: Go 1.25+, Node.js 18+ (for frontend)
**Primary Dependencies**: Go (Cobra, Gorilla WebSocket), React (Create React App, Xterm.js)
**Storage**: N/A (state is managed in memory by the CLI server)
**Testing**: Go testing, Jest/React Testing Library
**Target Platform**: Local machine (browser-based UI)
**Project Type**: Web (frontend + backend)
**Performance Goals**: Real-time log streaming for multiple services, responsive UI for command execution and service management
**Constraints**: Frontend assets must be embeddable in the Go binary.

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Simplicity**:
- Projects: 2 (Go CLI backend, React frontend)
- Using framework directly: Yes (Cobra, React)
- Single data model: Yes (for commands and logs)
- Avoiding patterns: Yes, keeping the interaction model simple.

**Architecture**:
- EVERY feature as library: The server functionality will be a new package within the CLI.
- Libraries listed: `server` (Go package), `frontend` (React app)
- CLI per library: The `server` command will be the entry point.
- Library docs: N/A

**Testing (NON-NEGOTIABLE)**:
- RED-GREEN-Refactor cycle enforced: Yes
- Git commits show tests before implementation: Yes
- Order: Contract→Integration→E2E→Unit strictly followed: Yes
- Real dependencies used: Yes
- Integration tests for: API endpoints, WebSocket communication, service management functionality.

**Observability**:
- Structured logging included: Yes, for the server component.
- Frontend logs → backend: Yes, for critical errors.
- Error context sufficient: Yes.

**Versioning**:
- Version number assigned: N/A for this feature, follows main CLI version.
- BUILD increments on every change: N/A
- Breaking changes handled: N/A

## Project Structure

### Documentation (this feature)
```
specs/003-extend-cli-with/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (OpenAPI spec)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)
```
# Backend (Go)
internal/server/
├── server.go
├── handlers.go
├── websocket.go
└── embed.go

# Frontend (React)
frontend/
├── public/
├── src/
│   ├── components/
│   ├── App.js
│   └── index.js
├── package.json
└── build/ (gitignored)
```

**Structure Decision**: Web application (Option 2)

## Phase 0: Outline & Research
1. **Go & React Integration**: Research best practices for embedding a React frontend into a Go binary.
2. **WebSocket Libraries**: Evaluate Gorilla WebSocket vs. other Go WebSocket libraries.
3. **Command Execution**: Determine the best way to execute and stream output from Cobra commands.
4. **Service Management**: Research how to control and monitor embedded Karaf, PostgreSQL, and Keycloak services individually.
5. **Frontend State Management**: Decide on a state management library for React (e.g., Redux, Zustand).

**Output**: `research.md` with decisions and best practices.

## Phase 1: Design & Contracts
1. **Data Model**: Define the JSON structures for commands, responses, and log messages in `data-model.md`.
2. **API Contracts**: Create an OpenAPI 3.0 specification for REST endpoints in `/contracts/openapi.yml`.
    - `GET /api/status`: Get application status.
    - `POST /api/actions/start`: Start the application.
    - `POST /api/actions/stop`: Stop the application.
 3. **WebSocket Contracts**: Define endpoints and message structure for log streaming and interactive session.
    - Log endpoints: `/ws/logs/combined`, `/ws/logs/service/{name}` where `{name} ∈ {karaf, postgresql, keycloak}`.
    - Log message: `{ "ts": "ISO-8601", "service": "karaf|postgresql|keycloak", "line": "..." }`
    - Session endpoint: `/ws/session` (one session per connection).
    - Session messages: client→server `{type: input|resize|control, ...}`, server→client `{type: output|status, ...}`
4. **Quickstart Guide**: Write a `quickstart.md` with setup, build, and run instructions.
5. **Agent File Update**: N/A for this feature.

**Output**: `data-model.md`, `/contracts/openapi.yml`, `quickstart.md`.

## Phase 2: Task Planning Approach
**Task Generation Strategy**:
- **Backend**:
  - Setup server framework and routing.
  - Implement REST endpoints for status and actions.
  - Implement WebSocket for log streaming and interactive session.
  - Implement combined and per-service log streams.
  - Implement interactive session PTY bridge (`/ws/session`) with input/output, resize, and interrupt handling.
  - Add file embedding for the frontend.
- **Frontend**:
  - Set up React project with Create React App.
  - Create components for dual terminals (A: logs with source selector; B: interactive session), A/B switch, and Service panel with status buttons.
  - Implement API calls to the backend.
  - Implement WebSocket clients for logs and session; use Xterm.js with fit addon and message batching.
  - Set up build process for production assets.

**Ordering Strategy**:
1. Backend server setup.
2. Frontend project setup.
3. Implement one feature end-to-end (e.g., status button).
4. Implement remaining features.
5. Implement frontend embedding.

**Estimated Output**: 20-25 tasks in `tasks.md`.

## Complexity Tracking
N/A

## Progress Tracking
- [ ] Phase 0: Research complete
- [ ] Phase 1: Design complete
- [ ] Phase 2: Task planning complete
- [ ] Phase 3: Tasks generated
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

---
## Amendment (2025-09-15): UI Relabeling, Services Toggle, Init Gate, TTY Parity

Scope
- Relabel terminals to 'Logs' and 'JUDO Terminal'.
- Move Services toggle to a left-edge control (remove header button).
- Gate Logs/JUDO Terminal behind project initialization with a modal prompt and clear notice on decline.
- Ensure JUDO Terminal parity with native 'judo session' (pure TTY bridge).

Design & Contracts
- REST: GET `/api/project/init/status` → `{ initialized: boolean, message?: string }`.
- REST: POST `/api/project/init` → starts initialization and returns `{ state: 'started' }`; progress surfaced via existing logs and status polling.
- WS Session handshake: client sends `init` with `{ term: 'xterm-256color', cols, rows }` immediately after connect; server configures PTY accordingly; subsequent `resize` updates dimensions. No client-side prompt injection.

Frontend UX Flow
- On load, fetch init status. If not initialized: show modal 'Initialize project now?'.
  - Yes: call init endpoint; show progress (logs/status); enable terminals when complete.
  - No: show non-blocking banner/toast explaining initialization is required to connect; keep terminals disabled until initialized.
- Replace A/B labels with 'Logs' and 'JUDO Terminal'.
- Replace header Services button with left-edge toggle.

Testing
- E2E: init gate prompt, decline notice, terminals disabled/enabled, labels, Services toggle placement.
- Parity: compare `judo session` outputs and control behavior (Ctrl+C, history, prompt) between browser terminal and OS terminal.

Tasks Reference
- See tasks T033–T044.

## Amendment (2025-09-15): Test Model Integration

Test Environment
- Use `test-model/` as the canonical environment for development, demos, and automated tests.
- Run all end-to-end flows (generate, build, start, stop, dump, import, export) inside `test-model/`.

Developer Flow
- cd into `test-model/` and execute CLI commands; server UI features (logs, services, session) must reflect this project’s runtime state.

Testing
- Prefer `test-model/` for integration and E2E tests to avoid external project drift. Seed/cleanup via existing CLI commands (dump/import, stop/clean).

Tasks Reference
- See tasks T043–T050.

*Based on Constitution v2.2.0 - See `/memory/constitution.md`*
