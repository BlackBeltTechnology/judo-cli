---
title: "Command Reference"
weight: 2
has_children: true
description: "Complete reference for all JUDO CLI commands"
---

# Command Reference

Complete documentation for all JUDO CLI commands organized by category.

## Quick Reference

| Command | Description |
|---------|-------------|
| [`judo doctor`](doctor/) | Check system health and dependencies |
| [`judo init`](init/) | Initialize new JUDO project |
| [`judo build`](build/) | Build project components |
| [`judo start`](start/) | Start application and services |
| [`judo stop`](stop/) | Stop all running services |
| [`judo status`](status/) | Check service status |
| [`judo clean`](clean/) | Clean environment |
| [`judo session`](session/) | Interactive session mode |

## Command Categories

### [System Commands](system/)
Essential commands for project setup and health checking.

### [Build Commands](build/)
Commands for building and packaging your JUDO application.

### [Application Lifecycle](lifecycle/)
Commands for starting, stopping, and monitoring your application.

### [Database Operations](database/)
Commands for database backup, restore, and schema management.

### [Maintenance Commands](maintenance/)
Commands for cleaning, updating, and maintaining your development environment.

## Global Flags

All commands support these global flags:

| Flag | Description |
|------|-------------|
| `-e, --env <environment>` | Use alternate environment profile (default: judo) |
| `-h, --help` | Show help for command |

## Getting Help

For any command, use the `--help` flag to get detailed usage information:

```bash
judo <command> --help
```

Or use the interactive session for contextual help:

```bash
judo session
help <command>
```