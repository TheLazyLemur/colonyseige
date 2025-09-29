# Colony Siege

A minimal defensive real-time strategy game built with Go and a custom Entity Component System (ECS) architecture.

## Overview

Colony Siege is a compact RTS where players build defenses, train armies, and fight toward a fortified enemy Hive. The game features constant wave-based attacks that force players to balance survival with gradual offense until the Hive Core is destroyed.

**Key Features:**
- Short, intense matches (10-20 minutes)
- Wave-based enemy attacks with escalating difficulty
- Defensive building system (walls, turrets, barracks)
- Resource management with population limits
- Custom ECS architecture for game logic

## Architecture

The project is organized into two main packages:

- **`internal/ecs/`** - Core Entity Component System implementation
  - `world.go` - World management and entity orchestration
  - `entity.go` - Entity creation and management
  - `component.go` - Component storage and retrieval
  - `system.go` - System interface and execution

- **`game/`** - Game-specific logic and components
  - `main.go` - Game entry point and main loop
  - `components.go` - Game-specific components (Position, Velocity, Health, etc.)
  - `systems.go` - Game systems (Movement, Health, etc.)

## Getting Started

### Prerequisites

- Go 1.25.1 or later

### Running the Game

```bash
go run game/main.go
```

The game will start a demo loop showing the ECS system in action with entities moving and updating over 20 frames.

## Game Design

**Core Gameplay Loop:**
1. Enemy waves attack on timed intervals
2. Players defend using units and defensive structures
3. Surviving waves grants resources for reinforcement
4. Players gradually push toward the enemy Hive
5. Victory achieved by destroying the Hive Core

**Design Pillars:**
- **Tension Through Pressure** - Constant waves keep players under threat
- **Defensive Creativity** - Experiment with defensive layouts and strategies
- **Forward Momentum** - Victory requires offensive action, not just survival
- **Simplicity & Clarity** - Easy to learn mechanics with tactical depth

## Documentation

See the `docs/` directory for detailed game design documentation:
- `docs/overview.md` - Complete game concept and design details
- `docs/index.md` - Documentation index

## Development Status

This project is in early development. The current implementation demonstrates the core ECS architecture with basic movement and health systems.
