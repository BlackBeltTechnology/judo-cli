# Feature Specification: Browser-Based Interactive CLI Server

**Feature Branch**: `003-extend-cli-with-a-server-function`  
**Created**: 2025-09-15  
**Status**: Draft  
**Input**: User description: "I would like to extend the CLI tool with a 'server' function, which has same functionality as session, but it can work in a browser with embedded server. The frontend have a command window, like in terminal. Very similar to LLM chat clients. It have to include statuses as buttons, where the server's start / stop can be performed. The server function controls the embedded karaf, postgresql and keycloak services. The server's log have to be accessed as tailing log. The server function have to be embedded in current cli program. The compiled frontend have to be included in binary release."

---

## ⚡ Quick Guidelines
- ✅ Focus on WHAT users need and WHY
- ❌ Avoid HOW to implement (no tech stack, APIs, code structure)
- 👥 Written for business stakeholders, not developers

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story
As a JUDO CLI user, I want to run the CLI as a server and interact with it through a web browser. This provides a user-friendly interface for viewing real-time service logs and using an interactive "judo session" terminal to run commands, while managing the status of embedded Karaf, PostgreSQL, and Keycloak services.

### Acceptance Scenarios
1. **Given** the JUDO CLI is not running, **When** I run `judo server`, **Then** the CLI should start a web server on port 6969 and open a new browser tab to the web UI.
1.1 **Given** I want to use a specific port, **When** I run `judo server -p 8080`, **Then** the CLI should start the web server on port 8080.
2. **Given** the web UI is open, **When** I look at the interface, **Then** I should see two Xterm terminals with an A/B toggle and a left-side Service panel toggle.
3. **Given** the Service panel is collapsed by default, **When** I click the left-edge button, **Then** the panel slides in and the active terminal resizes; collapsing the panel returns the terminal to full-screen.
4. **Given** I switch between Terminal A and Terminal B, **Then** each terminal preserves its scrollback and scroll position across switches.
5. **Given** Terminal A is active, **When** I choose a source (Combined, Karaf, PostgreSQL, Keycloak), **Then** the terminal displays only that selection; Combined shows labeled, color-distinguished live logs for all three.
6. **Given** Terminal B is active, **When** I type a command at the prompt (e.g., `help`, `status`, `build`), **Then** the command executes exactly as in `judo session` and the output appears in the terminal in real time.
7. **Given** I start or stop services using the Service panel, **Then** the corresponding startup/shutdown logs stream in real-time in Terminal A when its source includes that service.
8. **Given** a temporary network interruption occurs, **Then** the UI indicates reconnecting and resumes streaming without requiring a page refresh; Terminal B reattaches to the running session if available, otherwise starts a new session.
9. **Given** the compiled frontend is ready, **When** a new release of the CLI is built, **Then** the frontend assets MUST be embedded into the final binary.

### Edge Cases
- What happens if the default server port is already in use? The server should attempt to use the next available port.
- How does the UI handle high-volume log streaming? The UI should remain responsive, and rendering should be batched/throttled as needed.
- What happens if the Terminal B session process exits (e.g., user types `exit`)? The UI should show that the session ended and offer a button to start a new session.
- What happens if the browser tab is closed? The underlying CLI server should continue to run until explicitly stopped; Terminal B session is terminated when its WS disconnects unless configured to persist.

### Assumptions & Defaults
- Default terminal on load: Terminal A (Combined source).
- Terminal A sources: Combined, Karaf, PostgreSQL, Keycloak; Combined is default.
- Terminal B: Interactive `judo session` terminal; starts on first activation and stops when the session ends or the browser disconnects.
- Service panel state on load: collapsed by default; toggled via a left-edge button.
- Log line format: one timestamp (UTC) + service label + message; avoid duplicate timestamps if present in source logs.
- Visual distinction: accessible color assignments per service (Karaf, PostgreSQL, Keycloak).
- State persistence: active terminal, Terminal A source, and Service panel state persist across page reloads. Terminal B session is not persisted across reloads.
- Karaf log source: Karaf’s standard log output is available for real-time streaming.
- Container services: PostgreSQL and Keycloak log outputs are available for real-time streaming from their runtime environment.
- Security: Terminal B limits execution to `judo session` commands; no arbitrary OS shell is exposed.

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: The CLI MUST include a new `server` command that starts a local web server.
- **FR-002**: The `server` command MUST serve a single-page web application to the browser.
- **FR-003**: The web UI MUST provide two Xterm.js terminals with an A/B switch.
- **FR-004**: Terminal A MUST support a source selector with options: Combined, Karaf, PostgreSQL, Keycloak; Combined MUST display clearly labeled, color-distinguished live logs from all three.
- **FR-005**: Terminal B MUST provide an interactive `judo session` terminal with prompt, command history, copy/paste, Ctrl+C (interrupt), and auto-resize.
- **FR-006**: The web UI MUST include a left-side collapsible Service panel with controls to `start`, `stop`, and view `status` for each embedded service.
- **FR-007**: The UI MUST preserve each terminal’s scrollback and scroll position across switches, and fit the active terminal to the available space.
- **FR-008**: The compiled frontend assets MUST be embedded into the Go binary for distribution.
- **FR-009**: The server MUST provide individual service control (start/stop/status) and display per-service status/health indicators.
- **FR-010**: Log streaming MUST be real-time, resilient to temporary disconnects, and support Combined and per-service views with clear visual indicators.
- **FR-011**: Terminal B MUST reconnect gracefully after transient network failures, reattaching to the existing session when possible or starting a new session if the previous one has ended.
- **FR-012**: The server command MUST support a `-p` or `--port` flag to specify the server port, with a default of 6969.
- **FR-013**: All frontend functionality, including log viewing, terminal switching, interactive session behavior, and status updates, MUST be covered by UI tests.

### Key Entities
- **CLI Server**: The Go application that runs the web server, manages embedded service state (Karaf, PostgreSQL, Keycloak), and executes commands.
- **Web UI**: The single-page application served to the browser, providing the user interface for service management.
- **Log Streamer**: The component that tails service logs (Karaf, PostgreSQL, Keycloak) and sends them to the UI in real-time.
- **Session Bridge**: The component that runs and attaches to an interactive `judo session`, forwarding input, resize events, and output between the browser and the CLI process.
- **Service Manager**: The component that controls the lifecycle of embedded Karaf, PostgreSQL, and Keycloak services.
- **Service Monitor**: The component that tracks and reports the status and health of each embedded service.

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous  
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [ ] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [ ] Review checklist passed

---

## Spec Extension (2025-09-15)

Summary
- Services toggle: move to left edge; header button removed
- Terminal names: Terminal A → Logs; Terminal B → JUDO Terminal
- Project initialization: gate UI until initialized; modal prompt; 'No' shows clear notice that initialization is required to connect
- JUDO Terminal behavior: identical to OS 'judo session'; browser provides TTY only

Additional Acceptance Scenarios
- On first load when the project is not initialized, the UI prompts: 'Initialize project now?'. Choosing 'Yes' starts initialization and, after success, enables Logs and JUDO Terminal. Choosing 'No' shows a non-blocking notice that initialization is required to connect, and both terminals remain disabled.
- The Services toggle appears as a left-edge anchor that opens/closes the Services panel; the header contains no duplicate Services button.
- The tabs read 'Logs' and 'JUDO Terminal', and switching preserves scrollback and state as before.
- Commands typed in the JUDO Terminal behave exactly as in a native OS terminal running 'judo session', including control keys, history, prompts, and output formatting.

Additional Requirements
- FR-014: The Services toggle MUST be a left-edge control; the header MUST NOT include a redundant Services button.
- FR-015: The terminal labels MUST be 'Logs' and 'JUDO Terminal'.
- FR-016: If the project is not initialized, the application MUST prompt to initialize; selecting 'No' MUST present a clear notification that initialization is required to connect; until initialization completes, Logs and JUDO Terminal MUST be inactive.
- FR-017: The JUDO Terminal MUST provide parity with a native 'judo session', acting as a pure TTY bridge in the browser.

## Spec Extension (Test Model Project)

Summary
- A complete sample project is available under `test-model/` for end-to-end validation of the CLI and the browser-based server UI with real infrastructure (Karaf, PostgreSQL, Keycloak).
- All core flows (generate, build, start, stop, dump, import, export) are expected to work inside `test-model/` without external dependencies beyond documented prerequisites.

Additional Acceptance Scenarios
- Given I am inside `test-model/`, when I run `judo build start`, then the application builds and services start successfully, and the web UI connects and streams logs from those services.
- Given the environment was previously running, when I run `judo dump` followed by `judo import`, then the database is exported and restored correctly using the test-model configuration.
- Given I run `judo server` from `test-model/`, then service controls and logs in the UI reflect the state of the test-model services.

Additional Requirements
- FR-018: Documentation MUST describe how to use `test-model/` for local testing with infrastructure.
- FR-019: E2E/UI tests MUST target `test-model/` to ensure reproducible behavior.

*Based on Constitution v2.2.0 - See `/memory/constitution.md`*
