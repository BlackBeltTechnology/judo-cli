---
layout: home
title: JUDO CLI
subtitle: Command Line Interface for Low-Code Development
show_install_tabs: true
sections:
  - id: features
    title: "Core Features"
    subtitle: "Everything you need for low-code development"
    blocks:
      - heading: "Interactive Session Mode"
        text: "Command history, auto-completion, and persistent state for seamless development workflow."
      - heading: "Multi-Runtime Support"
        text: "Choose between Karaf and Docker Compose environments based on your needs."
      - heading: "Database Management"
        text: "Built-in PostgreSQL operations including dump, import, and schema migrations."
      - heading: "Cross-platform Support"
        text: "Works on macOS, Linux, and Windows with consistent functionality."
  - id: commands
    title: "Command Categories"
    subtitle: "Organized toolset for every development phase"
    blocks:
      - heading: "System Commands"
        text: "Health checks with 'judo doctor', project initialization with 'judo init', and interactive sessions."
      - heading: "Build & Run"
        text: "Build projects, start applications, and use 'judo reckless' for fast development cycles."
      - heading: "Application Lifecycle"
        text: "Start, stop, check status, and view logs for your applications and services."
      - heading: "Database Operations"
        text: "Backup with 'judo dump', restore with 'judo import', and manage schema upgrades."
  - id: runtime-modes
    title: "Runtime Environments"
    subtitle: "Flexible deployment options"
    blocks:
      - heading: "Karaf Runtime"
        text: "Local development with Apache Karaf application server plus Docker services for database and authentication."
      - heading: "Compose Runtime"
        text: "Full Docker Compose environment with all services containerized for consistent deployment."
  - id: interactive-session
    title: "Interactive Session"
    subtitle: "Enhanced development workflow"
    blocks:
      - heading: "Command History"
        text: "Persistent command history across sessions with search and navigation"
      - heading: "Tab Completion"
        text: "Auto-completion for commands and flags with intelligent suggestions"
      - heading: "Real-time Status"
        text: "Live service status indicators showing system health"
      - heading: "Context Awareness"
        text: "Smart suggestions based on current project state and workflow"
  - id: configuration
    title: "Configuration"
    subtitle: "Profile-based configuration system"
    blocks:
      - heading: "Default Profile"
        text: "Use judo.properties for default application and database settings"
      - heading: "Environment Profiles"
        text: "Create environment-specific files like compose-dev.properties for different setups"
      - heading: "Version Constraints"
        text: "Define minimum versions in judo-version.properties for compatibility"
---
