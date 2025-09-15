# Research: Browser-Based Interactive CLI Server

## Go & React Integration
**Decision**: Use Go's `embed` package to embed the compiled React frontend into the Go binary.
**Rationale**: This is the standard, idiomatic way to bundle static assets in Go. It creates a single, self-contained binary, which is a core requirement. The `net/http` package can easily serve these embedded files.
**Alternatives considered**:
- **Serving from disk**: This would require distributing the frontend assets alongside the binary, which complicates deployment and violates the single-binary requirement.
- **Git submodules**: This would manage the frontend as a separate repository, but still requires a separate build and embedding step.

## WebSocket Libraries
**Decision**: Use `gorilla/websocket`.
**Rationale**: It is the most popular and well-documented WebSocket library for Go. It is robust, feature-rich, and has a stable API. It is the de-facto standard for WebSocket communication in Go.
**Alternatives considered**:
- **`nhooyr.io/websocket`**: A newer, well-regarded library, but less battle-tested than Gorilla.
- **`golang.org/x/net/websocket`**: The official Go sub-repository package, but it is less feature-rich and has some known limitations compared to Gorilla.

## Command Execution
**Decision**: Use the `os/exec` package to run `judo` commands as subprocesses. The output (stdout and stderr) will be captured and streamed to the frontend.
**Rationale**: This approach isolates the command execution from the server process, preventing crashes or hangs from affecting the server. It also accurately captures the command's output as it would appear in a real terminal.
**Alternatives considered**:
- **Calling Cobra commands directly in-process**: This is more complex to manage, as it requires careful handling of I/O streams and can lead to state management issues. It also risks crashing the server if a command panics.

## Frontend State Management
**Decision**: Use React's built-in state management (useState, useContext) for now.
**Rationale**: The application's state is relatively simple. A dedicated state management library like Redux or Zustand would be overkill for the initial implementation. We can introduce one later if the complexity grows.
**Alternatives considered**:
- **Redux**: Too much boilerplate for this project's needs.
- **Zustand**: A good lightweight option, but still not necessary for the initial version.

