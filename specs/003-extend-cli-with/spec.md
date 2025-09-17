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
6. **Given** Terminal B is active, **When** I type a command at the prompt (e.g., `help`, `status`, `build`), **Then** the command executes using direct function calls (not by executing judo commands) and the output appears in the terminal in real time.
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
- **FR-003**: The web UI MUST provide two `react-xtermjs` terminals with an A/B switch.
- **FR-004**: Terminal A MUST support a source selector with options: Combined, Karaf, PostgreSQL, Keycloak; Combined MUST display clearly labeled, color-distinguished live logs from all three.
- **FR-005**: Terminal B MUST provide an interactive session terminal with prompt, command history, copy/paste, Ctrl+C (interrupt), and auto-resize, using direct function calls rather than executing judo commands.
- **FR-006**: The web UI MUST include a left-side collapsible Service panel with controls to `start`, `stop`, and view `status` for each embedded service.
- **FR-007**: The UI MUST preserve each terminal’s scrollback and scroll position across switches, and fit the active terminal to the available space.
- **FR-008**: The compiled frontend assets MUST be embedded into the Go binary for distribution.
- **FR-009**: The server MUST provide individual service control (start/stop/status) and display per-service status/health indicators.
- **FR-010**: Log streaming MUST be real-time, resilient to temporary disconnects, and support Combined and per-service views with clear visual indicators, using direct log access rather than executing judo commands.
- **FR-011**: Terminal B MUST reconnect gracefully after transient network failures, reattaching to the existing session when possible or starting a new session if the previous one has ended, using direct session management rather than command execution.
- **FR-012**: The server command MUST support a `-p` or `--port` flag to specify the server port, with a default of 6969.
- **FR-013**: All frontend functionality, including log viewing, terminal switching, interactive session behavior, and status updates, MUST be covered by UI tests.
- **FR-025**: UI tests MUST verify terminal switching behavior preserves scrollback and state across switches.
- **FR-026**: UI tests MUST validate log streaming functionality with different sources (Combined, Karaf, PostgreSQL, Keycloak).
- **FR-027**: UI tests MUST cover service panel interactions including start/stop operations and status updates.
- **FR-028**: UI tests MUST verify WebSocket connection and reconnection behavior for both log and session streams.
- **FR-029**: UI tests MUST validate project initialization flow including modal handling and terminal enablement.
- **FR-030**: UI tests MUST ensure JUDO Terminal provides parity with native 'judo session' behavior.
- **FR-031**: UI tests MUST cover error handling and user feedback for service operations and connection issues.
- **FR-032**: UI tests MUST validate visual indicators and color coding for different log sources.
- **FR-033**: UI tests MUST verify service status indicators update correctly based on actual service state.
- **FR-034**: UI tests MUST cover resize behavior and terminal fitting to available space.
- **FR-035**: UI tests MUST validate that all user interactions produce appropriate visual feedback.
- **FR-036**: UI tests MUST be integrated into CI/CD pipeline and run against the `test-model/` environment.

### Key Entities
- **CLI Server**: The Go application that runs the web server, manages embedded service state (Karaf, PostgreSQL, Keycloak), and executes commands.
- **Web UI**: The single-page application served to the browser, providing the user interface for service management.
- **Log Streamer**: The component that tails service logs (Karaf, PostgreSQL, Keycloak) and sends them to the UI in real-time.
- **Session Bridge**: The component that directly interfaces with the interactive session functionality, forwarding input, resize events, and output between the browser and the CLI process without executing external judo commands.
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
- FR-017: The JUDO Terminal MUST provide parity with a native 'judo session', acting as a pure TTY bridge in the browser using direct function calls rather than command execution.

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

## Spec Extension (Architectural Refinement)

Summary
- Server functionality uses direct function calls instead of executing judo commands for logs and session streaming
- Eliminates dependency on relative binary paths and improves reliability

Additional Acceptance Scenarios
- Given the server is running, when it streams logs or handles terminal sessions, then it uses direct internal function calls rather than executing external judo commands
- Given the server needs to access service logs, then it reads log files directly or uses internal log streaming APIs instead of executing judo log commands

Additional Requirements
- FR-020: Log streaming MUST use direct log file access or internal streaming APIs rather than executing judo commands
- FR-021: Terminal session management MUST use direct internal session functions rather than executing judo session commands
- FR-022: Service status checking MUST use direct internal status functions rather than executing judo status commands
- FR-023: Documentation MUST be updated (README, internal/help, docs/commands) to reflect server behavior and interactions
- FR-024: CI/CD workflows (build.yml, hugo.yml, release pipeline) MUST be reviewed and updated as needed to ensure docs and builds remain intact

## Spec Extension (Comprehensive UI Testing)

Summary
- Comprehensive UI test suite covering all user interactions, visual feedback, and edge cases
- Tests integrated into CI/CD pipeline and run against the `test-model/` environment
- Focus on user experience validation and regression prevention

Additional Acceptance Scenarios
- Given the UI test suite is executed, when all tests pass, then all user interactions and visual feedback mechanisms are verified to work correctly
- Given a UI regression is introduced, when the test suite runs, then it should detect the regression and fail appropriately
- Given the test environment is set up, when UI tests run against `test-model/`, then they should validate real service interactions and log streaming

Additional Requirements
- FR-037: UI tests MUST cover terminal initialization and connection state transitions (connecting, connected, disconnected, reconnecting)
- FR-038: UI tests MUST validate terminal resize behavior and proper fitting to available space during panel toggling
- FR-039: UI tests MUST verify scrollback preservation and scroll position maintenance during terminal switches
- FR-040: UI tests MUST cover all service panel interactions including start, stop, and status refresh operations
- FR-041: UI tests MUST validate service status indicator updates based on actual service state changes
- FR-042: UI tests MUST verify log source switching behavior and proper filtering of log streams
- FR-043: UI tests MUST cover Combined log stream functionality with proper labeling and color coding
- FR-044: UI tests MUST validate JUDO Terminal command execution, output display, and control key handling
- FR-045: UI tests MUST verify command history functionality and navigation in JUDO Terminal
- FR-046: UI tests MUST cover session interruption (Ctrl+C) and session restart functionality
- FR-047: UI tests MUST validate WebSocket reconnection behavior and state recovery after network interruptions
- FR-048: UI tests MUST verify project initialization flow including modal presentation and terminal enablement
- FR-049: UI tests MUST cover error handling and user feedback for service operation failures
- FR-050: UI tests MUST validate visual indicators for different connection states and service statuses
- FR-051: UI tests MUST verify that all interactive elements provide appropriate visual feedback on hover, focus, and activation
- FR-052: UI tests MUST cover accessibility features including keyboard navigation and screen reader compatibility
- FR-053: UI tests MUST validate responsive design behavior across different screen sizes and orientations
- FR-054: UI tests MUST be executable in headless mode for CI/CD pipeline integration
- FR-055: UI tests MUST include visual regression testing to detect unintended UI changes
- FR-056: UI tests MUST cover performance aspects including render speed and responsiveness under load
- FR-057: UI tests MUST validate cross-browser compatibility for supported browser versions
- FR-058: UI tests MUST include mobile browser testing for responsive design validation
- FR-059: UI tests MUST be integrated with the main test suite and run as part of the standard test workflow

## Spec Extension (Comprehensive E2E Testing)

Summary
- Comprehensive end-to-end test suite covering the complete user journey from CLI startup to browser interaction
- Tests validate the integrated system behavior across CLI backend, web server, and React frontend
- Focus on real-world usage scenarios and system integration validation

Additional Acceptance Scenarios
- Given the E2E test suite is executed, when all tests pass, then the complete integrated system from CLI command to browser UI functions correctly
- Given a system regression is introduced, when the E2E test suite runs, then it should detect integration failures and provide clear error reporting
- Given the `test-model/` environment is properly set up, when E2E tests run, then they should validate real service lifecycle management and log streaming

Additional Requirements
- FR-060: E2E tests MUST cover complete CLI server startup and browser launch sequence
- FR-061: E2E tests MUST validate that the embedded frontend assets are properly served by the Go backend
- FR-062: E2E tests MUST verify WebSocket connections are established correctly for both log streaming and session management
- FR-063: E2E tests MUST cover service lifecycle management including start, stop, and status monitoring operations
- FR-064: E2E tests MUST validate real log streaming from all services (Karaf, PostgreSQL, Keycloak) through the complete pipeline
- FR-065: E2E tests MUST verify JUDO Terminal command execution produces identical results to native CLI session
- FR-066: E2E tests MUST cover project initialization flow from uninitialized state to fully operational
- FR-067: E2E tests MUST validate database operations (dump, import, export) through the JUDO Terminal interface
- FR-068: E2E tests MUST verify service status indicators reflect actual service state in real-time
- FR-069: E2E tests MUST cover error scenarios including service startup failures and network interruptions
- FR-070: E2E tests MUST validate that the system recovers gracefully from failures and maintains consistency
- FR-071: E2E tests MUST verify port configuration and conflict handling behavior
- FR-072: E2E tests MUST cover authentication and security aspects if implemented
- FR-073: E2E tests MUST validate that all user interactions produce the expected system-level outcomes
- FR-074: E2E tests MUST be executable against the `test-model/` environment for realistic validation
- FR-075: E2E tests MUST include performance benchmarking for critical user journeys
- FR-076: E2E tests MUST validate resource cleanup and proper shutdown procedures
- FR-077: E2E tests MUST cover cross-platform compatibility on supported operating systems
- FR-078: E2E tests MUST be integrated into CI/CD pipeline with appropriate environment setup
- FR-079: E2E tests MUST provide detailed logging and debugging information for failure analysis
- FR-080: E2E tests MUST be maintainable and resistant to flakiness through proper waiting strategies and stability measures

*Based on Constitution v2.3.0 - See `/memory/constitution.md`*
