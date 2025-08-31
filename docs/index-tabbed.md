---
layout: default
title: "JUDO CLI Features"
description: "Tabbed overview of JUDO CLI features and capabilities"
nav_order: 3
permalink: /docs/features/

# Tab configuration
tabs:
  - id: features
    title: "Core Features"
    subtitle: "Everything you need for low-code development"
    
  - id: commands
    title: "Command Categories"
    subtitle: "Organized toolset for every development phase"
    
  - id: runtime-modes
    title: "Runtime Environments"
    subtitle: "Flexible deployment options"
    
  - id: interactive-session
    title: "Interactive Session"
    subtitle: "Enhanced development workflow"
    
  - id: configuration
    title: "Configuration"
    subtitle: "Profile-based configuration system"

# Content for each tab section
features_content:
  blocks:
    - heading: "Interactive Session Mode"
      text: "Command history, auto-completion, and persistent state for seamless development workflow."
      icon: "terminal"
    
    - heading: "Multi-Runtime Support"
      text: "Choose between Karaf and Docker Compose environments based on your needs."
      icon: "cogs"
    
    - heading: "Database Management"
      text: "Built-in PostgreSQL operations including dump, import, and schema migrations."
      icon: "database"
    
    - heading: "Cross-platform Support"
      text: "Works on macOS, Linux, and Windows with consistent functionality."
      icon: "desktop"

commands_content:
  blocks:
    - heading: "System Commands"
      text: "Health checks with 'judo doctor', project initialization with 'judo init', and interactive sessions."
      icon: "gear"
    
    - heading: "Build & Run"
      text: "Build projects, start applications, and use 'judo reckless' for fast development cycles."
      icon: "rocket"
    
    - heading: "Application Lifecycle"
      text: "Start, stop, check status, and view logs for your applications and services."
      icon: "cycle"
    
    - heading: "Database Operations"
      text: "Backup with 'judo dump', restore with 'judo import', and manage schema upgrades."
      icon: "server"

runtime_modes_content:
  blocks:
    - heading: "Karaf Runtime"
      text: "Local development with Apache Karaf application server plus Docker services for database and authentication."
      icon: "container"
    
    - heading: "Compose Runtime"
      text: "Full Docker Compose environment with all services containerized for consistent deployment."
      icon: "compose"

interactive_session_content:
  blocks:
    - heading: "Command History"
      text: "Persistent command history across sessions with search and navigation"
      icon: "history"
    
    - heading: "Tab Completion"
      text: "Auto-completion for commands and flags with intelligent suggestions"
      icon: "tab"
    
    - heading: "Real-time Status"
      text: "Live service status indicators showing system health"
      icon: "status"
    
    - heading: "Context Awareness"
      text: "Smart suggestions based on current project state and workflow"
      icon: "brain"

configuration_content:
  blocks:
    - heading: "Default Profile"
      text: "Use judo.properties for default application and database settings"
      icon: "config"
    
    - heading: "Environment Profiles"
      text: "Create environment-specific files like compose-dev.properties for different setups"
      icon: "environment"
    
    - heading: "Version Constraints"
      text: "Define minimum versions in judo-version.properties for compatibility"
      icon: "version"
---

<!-- Hero Section -->
<section class="hero is-primary is-medium">
  <div class="hero-body">
    <div class="container has-text-centered">
      <h1 class="title is-1">{{ page.title }}</h1>
      <p class="subtitle is-3">{{ page.subtitle }}</p>
      {% if page.hero_image %}
      <figure class="image is-128x128 is-inline-block">
        <img src="{{ page.hero_image }}" alt="{{ page.title }}">
      </figure>
      {% endif %}
    </div>
  </div>
</section>

<!-- Tab Navigation -->
<section class="section tab-navigation">
  <div class="container">
    <div class="tabs is-centered is-boxed">
      <ul>
        {% for tab in page.tabs %}
        <li class="{% if forloop.first %}is-active{% endif %}" data-tab="{{ tab.id }}">
          <a>
            <span>{{ tab.title }}</span>
          </a>
        </li>
        {% endfor %}
      </ul>
    </div>
  </div>
</section>

<!-- Tab Content -->
<div class="tab-content-container">
  {% for tab in page.tabs %}
  <section class="section tab-content {% if forloop.first %}is-active{% endif %}" id="{{ tab.id }}-content">
    <div class="container">
      <div class="content has-text-centered">
        <h2 class="title is-2">{{ tab.title }}</h2>
        {% if tab.subtitle %}
        <p class="subtitle is-4">{{ tab.subtitle }}</p>
        {% endif %}
      </div>
      
      <div class="columns is-multiline is-centered">
        {% assign content_var = tab.id | append: '_content' %}
        {% assign tab_content = page[content_var] %}
        
        {% for block in tab_content.blocks %}
        <div class="column is-one-quarter">
          <div class="box feature-box">
            <div class="content has-text-centered">
              {% if block.icon %}
              <div class="icon-container">
                <span class="icon is-large">
                  <i class="fas fa-{{ block.icon }} fa-2x" aria-hidden="true"></i>
                </span>
              </div>
              {% endif %}
              
              <h3 class="title is-4">{{ block.heading }}</h3>
              <p>{{ block.text }}</p>
            </div>
          </div>
        </div>
        {% endfor %}
      </div>
    </div>
  </section>
  {% endfor %}
</div>

<style>
.hero.is-primary {
  background: linear-gradient(135deg, #0366d6 0%, #0556b3 100%);
}

.tab-navigation {
  padding-top: 2rem;
  padding-bottom: 0;
}

.tab-content-container {
  position: relative;
}

.tab-content {
  display: none;
  animation: fadeIn 0.3s ease;
}

.tab-content.is-active {
  display: block;
}

.feature-box {
  border: 1px solid #e1e4e8;
  border-radius: 8px;
  padding: 2rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  height: 100%;
  text-align: center;
}

.feature-box:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
}

.icon-container {
  margin-bottom: 1.5rem;
}

.icon-container .icon {
  color: #0366d6;
}

.feature-box h3 {
  margin-bottom: 1rem;
  color: #24292e;
}

.feature-box p {
  color: #586069;
  line-height: 1.6;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media screen and (max-width: 768px) {
  .columns.is-multiline .column {
    flex: none;
    width: 100%;
  }
  
  .feature-box {
    padding: 1.5rem;
  }
  
  .tabs ul {
    flex-direction: column;
  }
  
  .tabs li {
    margin-bottom: 0.5rem;
  }
}
</style>

<script>
document.addEventListener('DOMContentLoaded', function() {
  const tabLinks = document.querySelectorAll('.tabs li');
  const tabContents = document.querySelectorAll('.tab-content');
  
  tabLinks.forEach(function(tab) {
    tab.addEventListener('click', function() {
      const tabId = this.getAttribute('data-tab');
      
      // Remove active class from all tabs and contents
      tabLinks.forEach(function(t) {
        t.classList.remove('is-active');
      });
      
      tabContents.forEach(function(content) {
        content.classList.remove('is-active');
      });
      
      // Add active class to clicked tab and corresponding content
      this.classList.add('is-active');
      document.getElementById(tabId + '-content').classList.add('is-active');
    });
  });
});
</script>