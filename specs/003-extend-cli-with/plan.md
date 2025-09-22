
# Implementation Plan: Browser-Based Interactive CLI Server

**Branch**: `003-extend-cli-with` | **Date**: 2025-09-20 | **Spec**: `/specs/003-extend-cli-with/spec.md`
**Input**: Feature specification from `/specs/003-extend-cli-with/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Extend JUDO CLI with browser-based server functionality providing real-time service management, log streaming, and interactive terminal sessions. The feature adds a `judo server` command that starts a web server serving a React frontend with Xterm.js terminal for logs and JUDO Terminal for interactive sessions. Frontend assets are embedded in the Go binary, and the system provides comprehensive service control for Karaf, PostgreSQL, and Keycloak with real-time WebSocket connections for log streaming and interactive sessions.

## Technical Context
**Language/Version**: Go 1.25+, React/TypeScript, Node.js 18+  
**Primary Dependencies**: Cobra CLI framework, React/Xterm.js, WebSocket, Docker API  
**Storage**: File-based logs, Docker container state, in-memory session state  
**Testing**: Go testing framework, Vitest/React Testing Library, Playwright for E2E  
**Target Platform**: Cross-platform CLI (darwin, linux, windows) with web browser UI
**Project Type**: Web application (Go backend + React frontend)  
**Performance Goals**: Real-time log streaming with <100ms latency, WebSocket reconnection <5s, responsive UI with 60fps  
**Constraints**: Embedded frontend assets in Go binary, cross-platform compatibility, offline-capable service management  
**Scale/Scope**: Single user per instance, 50+ CLI commands, 4 concurrent service management, real-time log streaming for 3 services

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Simplicity & Single Responsibility**: Feature extends existing CLI with cohesive server functionality
- [x] **Consistent CLI UX**: New `server` command follows Cobra patterns with proper flags and documentation
- [x] **Test-First Discipline**: Comprehensive UI and E2E testing requirements specified
- [x] **Version Compliance**: Uses existing versioning system and build processes
- [x] **Observability & Error Handling**: Real-time logging with proper error handling and reconnection
- [x] **Documentation Integrity**: Documentation updates required for new server functionality
- [x] **Security & Secrets**: No secret handling required for this feature
- [x] **Frontend Testing**: Comprehensive Vitest and React Testing Library coverage specified
- [x] **CI/CD Integrity**: Build process includes frontend asset embedding, CI workflows maintained

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure]
```

**Structure Decision**: Option 2 (Web application) - Go backend with React frontend, embedded assets

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh opencode` for your AI assistant
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each WebSocket endpoint → contract test task [P]
- Each API endpoint → contract test task [P]
- Each data model entity → model creation task [P]
- Each user story → integration test task
- Implementation tasks for server command, WebSocket handlers, and React components
- JUDO Terminal specific tasks for interactive session functionality

**Ordering Strategy**:
- TDD order: Tests before implementation 
- Dependency order: Backend models → WebSocket handlers → API endpoints → Frontend components
- Infrastructure first: Server setup → Log streaming → Service management → JUDO Terminal
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 30-35 numbered, ordered tasks in tasks.md covering:
- Backend server implementation with WebSocket support
- Frontend React components with Xterm.js integration
- JUDO Terminal interactive session functionality
- Real-time log streaming for all services
- Service management controls
- Comprehensive testing suite

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan


## Phase 3: Bug Fixes and Refactoring

The initial implementation was followed by a critical phase of bug fixing and refactoring to address issues discovered during testing. The following work has been completed:

1.  **Frontend Bug Fixes:**
    -   **WebSocket Race Conditions:** Resolved race conditions that caused unreliable connections for both the log and session terminals on initial page load. The connection logic is now synchronized with the terminal component's lifecycle.
    -   **Terminal Interactivity:** Fixed issues that prevented user input, including a CSS bug that disabled the terminal and a focus management problem that prevented keystroke capture.

2.  **Code Refactoring:**
    -   **Component Extraction:** The main `App.tsx` file was broken down into smaller, single-responsibility components (`AppHeader`, `ServicePanel`, `TerminalContainer`, `ProjectInitModal`) to improve maintainability.
    -   **Code Consolidation:** Repetitive code, such as the terminal resizing logic, was extracted into a reusable `fitTerminal` function.

3.  **UX and Session Consistency:**
    -   **Interactive Terminal:** The JUDO Terminal was enhanced to provide a true `telnet`-like experience with real-time character echoing, local input buffering, and proper handling of the Enter and Backspace keys.
    -   **Output Formatting:** The backend was updated to send `\r\n` for all newlines, ensuring correct cursor behavior in the terminal.
    -   **Session Consistency:** The web terminal's appearance and behavior now align with the native `judo session`, including the display of the JUDO banner, a detailed status message on connection, and consistent `help` text.

## Phase 4+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command) - research.md exists
- [x] Phase 1: Design complete (/plan command) - data-model.md, contracts/, quickstart.md exist
- [ ] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
