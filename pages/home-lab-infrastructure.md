---
title: Home Lab Infrastructure
tags: [home-lab, infrastructure, network, services]
created: 2026-05-10T14:18:00+10:00
updated: 2026-05-10T15:40:00+10:00
---

# Home Lab Infrastructure

## Network

| Item | Value |
|------|-------|
| LAN IP | 192.168.1.125 |
| Subnet | 192.168.1.0/24 |
| VPN | Tailscale |

## Server

- **Host**: Zotac CI527 (Intel NUC)
- **RAM**: 16GB
- **GPU**: NVIDIA GTX 1070
- **OS**: Linux Mint 22

## Core Services

| Service | Port | Description |
|---------|------|-------------|
| Hermes Agent | 3008 | Agent orchestration |
| MissionControl Main | 3000 | Production |
| MissionControl Dev | 3002 | Development |
| Checkin API | 3005 | Family check-in PWA backend |
| Checkin PWA | 5173 | Family check-in PWA frontend |
| WebGPU | 3010 | Shader experiments |

## Data

### Obsidian Vaults

- **AgentVault** (`~/Documents/`) - Projects, notes, research
- **FamilyVault** (`~/Documents/`) - Personal life
- **Work-XRG** (`~/Documents/`) - Work notes

### GitHub Projects

- Checkin - Family check-in PWA
- ServiceDashboard - Services monitor
- DroneSim - Drone simulator
- SplashPainter27 - Paint app

## Access

- Tailscale VPN enabled for remote access
- Use `--insecure` flag for LAN access (Hermes requires it for 0.0.0.0 binding)
- All services exposed on all interfaces

## Diagrams

### Infrastructure Diagram
![Infra Diagram](assets/infra-diagram.png)

### KiwiFS Spaces
![KiwiFS Spaces](assets/kiwifs-diagram.png)

## KiwiFS Knowledge

KiwiFS runs on `localhost:3333` and provides knowledge management.

### Spaces

| Space | Purpose |
|-------|---------|
| default | General docs, home lab |
| family | People, health, finance, events |
| projects | checkin, mcdev, mcmain, webgpu |
| research | Graphics, physics, game |
| agent | AI, models, skills |

### Structure

```
knowledge/
├── pages/       # Permanent docs
├── episodes/    # Temp notes, TODOs
└── assets/      # Images, binaries
```

### API

```bash
curl -X PUT "http://localhost:3333/api/kiwi/{space}/file?path=pages/name.md"
curl -X GET "http://localhost:3333/api/kiwi/{space}/file?path=pages/name.md"
curl -X POST "http://localhost:3333/api/kiwi/{space}/bulk"
```
