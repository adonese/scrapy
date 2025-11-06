---
description: Execute a complete wave of parallel agents
---

# Parallel Wave Execution

Execute multiple agents in parallel for a specific wave from plan.md

## Available Waves:
- **wave1**: Foundation (project setup, database, temporal, monitoring)
- **wave2**: Core Services (API, scrapers, UX, frontend)
- **wave3**: Data Collection (all scrapers)
- **wave4**: Features (calculator, trends, growth, UI)
- **wave5**: Integration (sequential)
- **wave6**: Deployment (sequential)

## Usage:
Specify which wave to execute and the system will launch all agents in that wave simultaneously.

## Example:
"Execute wave1" will launch:
- Agent 1: Project Scaffolding
- Agent 2: Database Setup
- Agent 3: Temporal Foundation
- Agent 4: Monitoring Stack

All agents will work in parallel on their respective tasks.