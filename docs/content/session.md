---
layout: page
title: Interactive Session
nav_order: 4
description: "Using JUDO CLI's interactive session mode"
permalink: /session/
---

# Interactive Session Mode

JUDO CLI includes a powerful interactive session mode that provides enhanced productivity features for development workflows.

## Starting a Session

```bash
judo session
```

## Using the Included Test Model

For a ready-to-use environment, run the session inside `test-model/` (included in this repo). It contains a full sample project wired to real infrastructure (Karaf, PostgreSQL, Keycloak).

```bash
cd test-model
judo session
```

This lets you try history, completion, and the service-aware prompt against a working setup.

## Features

### Command History
- **Persistent**: Command history is saved across sessions
- **Searchable**: Use Ctrl+R to search through previous commands
- **Smart**: Maintains last 100 commands with timestamps

### Auto-completion
- **Commands**: Tab completion for all available commands
- **Flags**: Complete command flags and options
- **History**: Suggestions based on previous commands

### Dynamic Service Status
The prompt shows real-time status of your JUDO services:

```
judo [⚙️karaf:✓ 🔐keycloak:✗ 🐘postgres:✓]> 
```

**Status Indicators:**
- ⚙️ **karaf**: ✓ (running) / ✗ (stopped)
- 🔐 **keycloak**: ✓ (running) / ✗ (stopped)  
- 🐘 **postgres**: ✓ (running) / ✗ (stopped)

### Enhanced Help System
- **Contextual**: Get help specific to your current project state
- **Command-specific**: Type `<command>?` for detailed help
- **Interactive**: Guided command building

## Session Commands

In addition to all regular JUDO commands, session mode provides special commands:

| Command | Description |
|---------|-------------|
| `help` | Show session help with all available commands |
| `exit` or `quit` | Exit the interactive session |
| `clear` | Clear the terminal screen |
| `history` | Show command history for current session |
| `status` | Show detailed session and project information |
| `doctor` | Run system health check with verbose output |

## Usage Examples

### Basic Commands
```bash
# Within session
help                    # Show all commands
build -f               # Build frontend
start --skip-keycloak  # Start without Keycloak
status                 # Check service status
```

### Tab Completion Examples
- `bu<Tab>` → `build`
- `build --ski<Tab>` → `build --skip-`
- `start --opt<Tab>` → `start --options`

### History Features
```bash
history                # Show recent commands
<Ctrl+R>build          # Search for commands containing "build"
!!                     # Repeat last command
!build                 # Repeat last command starting with "build"
```

## Development Workflows

### Quick Development Cycle
```bash
judo session
> reckless             # Fast build and start
> log -f               # Follow logs
> <Ctrl+C>             # Stop log following
> build -a             # Rebuild app module
> status               # Check services
```

### Frontend Development
```bash
judo session
> build -f -q          # Quick frontend build
> start --skip-keycloak # Start without auth
> build -f -q          # Rebuild as needed
```

### Database Operations
```bash
judo session
> dump                 # Create backup
> clean               # Clean environment
> import               # Restore from backup
> status               # Verify services
```

## Advanced Features

### Execution Feedback
Commands show duration and success/failure status:
```
> build -f
✓ Command completed in 2.34s

> start
✗ Command failed in 0.12s (exit code 1)
```

### Context Awareness
- Shows project initialization status
- Adapts suggestions based on current state
- Remembers environment preferences

### Session Statistics
```bash
> status
Session Duration: 15m 23s
Commands Executed: 12
Project: MyProject (karaf runtime)
Services: karaf(✓) postgres(✓) keycloak(✗)
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Tab` | Auto-complete command/flag |
| `Ctrl+R` | Search command history |
| `Ctrl+C` | Cancel current input |
| `Ctrl+D` | Exit session |
| `Ctrl+L` | Clear screen |
| `↑/↓` | Navigate command history |

## Tips and Best Practices

1. **Use Tab Completion**: Always use tab completion to avoid typos and discover available options

2. **Leverage History**: Use `!!` and `!<prefix>` to quickly repeat commands

3. **Monitor Status**: Keep an eye on the service status indicators in the prompt

4. **Use Session Commands**: Take advantage of `clear`, `history`, and `status` for better workflow

5. **Quick Help**: Use `<command>?` for quick help without leaving the session

6. **Development Patterns**: Develop muscle memory for common command sequences like `build -f -q` for frontend development

## Troubleshooting

### Session Won't Start
- Check if another session is running
- Verify terminal compatibility
- Try running `judo session --debug` (if available)

### Auto-completion Not Working
- Ensure you're using a compatible terminal
- Check that the session started properly
- Try typing a few characters before using Tab

### History Not Persisting
- Check write permissions in your home directory
- Verify the session is closing properly with `exit`

## Exiting the Session

Always exit properly to save history and session state:

```bash
> exit
# or
> quit
# or press Ctrl+D
```