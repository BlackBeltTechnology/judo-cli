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
As a JUDO CLI user, I want to run the CLI as a server and interact with it through a web browser. This will provide a user-friendly, graphical interface for executing commands, viewing logs, and managing the status of embedded Karaf, PostgreSQL, and Keycloak services, similar to a chat client, without needing to stay in a terminal.

### Acceptance Scenarios
1. **Given** the JUDO CLI is not running, **When** I run `judo server`, **Then** the CLI should start a web server and open a new browser tab to the web UI.
2. **Given** the web UI is open, **When** I look at the interface, **Then** I should see a command input area, an output display for command results, a real-time log viewer, and status buttons for `start`, `stop`, and `status`.
3. **Given** the embedded services (Karaf, PostgreSQL, Keycloak) are stopped, **When** I click the "Start" button, **Then** the CLI should start all embedded services and the log viewer should show the services' startup logs in real-time.
4. **Given** the application is running, **When** I type `judo build -q` into the command input and submit, **Then** the command should be executed, and the output should be displayed in the output area.
5. **Given** I want to view the service logs, **When** I look at the log viewer, **Then** I should see a continuously updated tail of the Karaf, PostgreSQL, and Keycloak service logs, with the ability to filter by individual service or view combined logs.
6. **Given** I want to focus on a specific service, **When** I select to view only Karaf logs, **Then** the log viewer should display only Karaf service logs in real-time.
7. **Given** I want to focus on a specific service, **When** I select to view only PostgreSQL logs, **Then** the log viewer should display only PostgreSQL service logs in real-time.
8. **Given** I want to focus on a specific service, **When** I select to view only Keycloak logs, **Then** the log viewer should display only Keycloak service logs in real-time.
6. **Given** the compiled frontend is ready, **When** a new release of the CLI is built, **Then** the frontend assets MUST be embedded into the final binary.

### Edge Cases
- What happens if the default server port is already in use? The server should attempt to use the next available port.
- How does the UI handle long-running commands? The UI should remain responsive, and output should be streamed.
- What happens if the browser tab is closed? The underlying CLI server should continue to run until explicitly stopped.

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: The CLI MUST include a new `server` command that starts a local web server.
- **FR-002**: The `server` command MUST serve a single-page web application to the browser.
- **FR-003**: The web UI MUST provide a terminal-like interface for command input and output.
- **FR-004**: The web UI MUST display status buttons to `start`, `stop`, and check the `status` of the embedded Karaf, PostgreSQL, and Keycloak services.
- **FR-005**: The web UI MUST include a log viewer that tails the service outputs in real-time, with individual filtering capability for Karaf, PostgreSQL, and Keycloak services, as well as a combined view option.
- **FR-006**: The CLI server MUST handle command execution and stream output back to the web UI.
- **FR-007**: The compiled frontend assets MUST be embedded into the Go binary for distribution.
- **FR-008**: The server MUST provide individual service control (start/stop/status) for Karaf, PostgreSQL, and Keycloak.
- **FR-009**: The server MUST display individual service statuses and health indicators.
- **FR-010**: The log viewer MUST support individual service log filtering (Karaf-only, PostgreSQL-only, Keycloak-only) and combined log display, with clear visual indicators for each service type.
- **FR-011**: All frontend functionality, including command execution, log viewing, and status updates, MUST be covered by UI tests.

### Key Entities
- **CLI Server**: The Go application that runs the web server, manages embedded service state (Karaf, PostgreSQL, Keycloak), and executes commands.
- **Web UI**: The single-page application served to the browser, providing the user interface for service management.
- **Command Executor**: The component responsible for running CLI commands and capturing their output.
- **Log Streamer**: The component that tails service logs (Karaf, PostgreSQL, Keycloak) and sends them to the UI in real-time.
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

*Based on Constitution v2.2.0 - See `/memory/constitution.md`*
