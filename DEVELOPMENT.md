# GateWars v3 — Development Guide

## Prerequisites

- **Go 1.22+** — [install](https://go.dev/dl/)
- A terminal with 256-color support
- An SSH client (OpenSSH, etc.)

## Setup

```bash
git clone <repo-url>
cd ssh-gatewars
go mod download
```

## Build & Run

```bash
make build    # Compile to ./gatewars
make run      # Build + run
make dev      # Build + run with fixed seed and 20 planets
make vet      # Run go vet
make clean    # Remove binary + database
```

On first launch, the server generates an SSH host key pair and creates the SQLite database.

## Testing Locally

```bash
# Terminal 1 — server
make dev

# Terminal 2 — client
ssh -p 2222 -o StrictHostKeyChecking=no localhost
```

Multiple clients can connect simultaneously. Each SSH key gets its own progression.

## Package Layout

```
cmd/server/main.go              Entry point, flag parsing, goroutine wiring
internal/
  core/
    config.go                   GameConfig with all tunable parameters
    session.go                  Per-SSH-connection metadata
    types.go                    Shared types: Vec2, Rect, Pos
  game/
    galaxy.go                   Galaxy state, planet collection
    planet.go                   Planet struct (status, invasion, surge, bounty)
    chair.go                    Ancient Control Chair (faction-based scaling)
    drone.go                    Drone types, tiers, per-owner tracking
    replicator.go               Enemy types and stats
    faction.go                  Ancient vs Ori paths + scaling definitions
    tactic.go                   Drone targeting tactics (Spread/Focus/Perimeter)
  engine/
    engine.go                   Main tick loop, surge rotation, milestones, New Game+
    defense.go                  Per-planet defense, salvo firing, tactic targeting, surge modifier
  store/
    player.go                   PlayerStore (SQLite CRUD, ZPM, upgrades)
    migrations.go               v3 schema (players, galaxy_planets, chat, teams)
    chat.go                     Chat message persistence
    team.go                     Team management
    callsign.go                 Callsign uniqueness, mutes
  chat/
    hub.go                      Single-goroutine chat event router
    channel.go                  Channel + RingBuffer
    message.go                  Message/Event type definitions
    commands.go                 Slash command handlers
    walter.go                   Walter NPC + game event announcements
  server/
    ssh.go                      Wish SSH server + per-connection handler
    identity.go                 SSH key fingerprinting
    limiter.go                  Token bucket rate limiter
  tui/
    model.go                    Bubbletea model, state machine, input handling
    state.go                    State enum (Splash -> Callsign -> Atlantis -> Throne/Galaxy -> Defense)
    views/
      theme.go                  Centralized color palette, pre-built styles, layout helpers
      splash.go                 Multi-phase animated boot sequence
      callsign.go               Biometric identification terminal
      atlantis.go               Responsive hub (stats, chair art, chat side-by-side)
      galaxy.go                 Sensor display planet browser with detail panel
      throne.go                 Upgrade terminal (factions, drone tiers, reset)
      defense.go                Radial defense view with concentric rings, entity rendering
      chatpanel.go              Rounded-border chat panel with word-aware wrapping
```

## State Machine

```
Splash -> Callsign -> Atlantis (hub)
                        | t          | g
                     Throne       Galaxy (browser)
                     (upgrades)      | enter
                        | q       Defense (gameplay)
                     Atlantis        | q
                                  Atlantis
```

Chat is available in all states via `c` key.

## Concurrency Model

```
SSH Connection -> Bubbletea Program (per player)
                      | (actions)
Engine Goroutine -> tick loop -> DefenseInstances
                      | (game events)
Chat Hub Goroutine -> event routing -> session outboxes
                      ^ (chat events)
TUI Programs ----------
```

- **Engine**: single writer, sync.RWMutex for instance access, atomic snapshots for galaxy
- **Chat Hub**: single goroutine, channel-based, no locks
- **TUI**: reads snapshots (lock-free), sends events via channels

## Defense Simulation

Each planet with active defenders has a DefenseInstance:

1. **Wave spawning**: replicators spawn at SpawnRadius from center, approach inward
2. **Chair auto-fire**: each chair fires drones at nearest replicator on cooldown
3. **Drone tracking**: drones home toward their target, re-aim each tick
4. **Collision**: drone hits replicator -> damage, splash, or pierce depending on tier
5. **Breach check**: replicator reaches chair -> shield damage
6. **Hold timer**: accumulates ticks while chairs survive; reaches threshold -> liberation

## Replicator Types

| Type | HP | Speed | Damage | ZPM Drop |
|------|-----|-------|--------|----------|
| Bug (Basic) | 1 | 1.0x | 1 | 5 |
| Sentinel (Armored) | 3 | 0.7x | 2 | 15 |
| Queen | 10 | 0.4x | 5 | 50 |

## Factions (Ancient vs Ori)

Players choose a faction path in the Throne. Each has different scaling:

| Stat | Ancient | Ori |
|------|---------|-----|
| Max Drones | 5 + 4/level (up to 45) | 3 + 2/level (up to 23) |
| Drone Damage | 1.0x base | 2.0x base |
| Shield HP | 1.25x (125%) | 0.8x (80%) |
| Fire Rate | 10 → 4 ticks | 7 → 3 ticks |
| Salvo | 1 + level/3 (max 4) | 1 + level/4 (max 3) |

Switching faction resets all upgrades. A full reset option is also available in the Throne (zeroes ZPM, chair level, drone tier, and faction).

## Drone Tiers

Base damage before faction multiplier:

| Tier | Color | Damage | Speed | Special |
|------|-------|--------|-------|---------|
| Standard | Yellow | 1 | 1.0x | -- |
| Swift | Cyan | 1 | 1.5x | -- |
| Blast | Magenta | 2 | 1.0x | Splash (2.0 radius) |
| Piercing | White | 3 | 1.2x | Passes through targets |

## Drone Tactics

Switchable during defense with `[1]` `[2]` `[3]` keys:

| Tactic | Key | Targeting |
|--------|-----|-----------|
| Spread | 1 | Nearest enemies (default) |
| Focus | 2 | Strongest threat first (queens) |
| Perimeter | 3 | Enemies closest to center |

## Galaxy Events

- **Replicator Surges** — random invaded planet gets surge flag: 2x spawns, 2x ZPM rewards
- **Planet Bounties** — each planet has bounty = invasion_level x 10 x cycle ZPM, awarded on liberation
- **Liberation Milestones** — server-wide announcements at 25%, 50%, 75%, 100% galaxy freed
- **New Game+** — when 100% liberated, all planets reset with scaled difficulty (cycle counter increments)

## Hold Timer

- Base: 5 minutes (300 seconds = 3000 ticks at 10Hz)
- Scales: 5 min x number of players
- 1 player = 5 min, 2 players = 10 min, 3 = 15 min

## Database Schema

- `players` — SSH fingerprint, callsign, ZPM, chair level, drone tier, faction, stats
- `galaxy_planets` — persistent galaxy state
- `chat_messages` — persistent chat (#ops, #team)
- `teams` / `team_members` — SG team management
- `callsigns` — unique callsign mapping
- `mutes` — player mute list

## Chat Channels

- `#ops` — global, persistent, auto-joined
- `#planet:<name>` — per-planet, ephemeral, joined on deploy
- `#sg-<name>` — team, persistent
- `[DM]` — direct messages, ephemeral

## Debugging

```bash
# Deterministic galaxy
./gatewars --seed 42 --planets 10

# Check database
sqlite3 gatewars.db "SELECT * FROM players;"
sqlite3 gatewars.db "SELECT * FROM galaxy_planets;"
sqlite3 gatewars.db "SELECT * FROM chat_messages ORDER BY created_at DESC LIMIT 20;"
```

The server logs to stderr via charmbracelet/log.

## Roadmap

- [x] Phase 1 — Foundation (SSH, engine, defense sim, TUI, chat)
- [x] Phase 2 — Upgrade system (Throne view, ZPM spending, chair levels, drone tier unlocks)
- [x] Phase 3 — Visual polish (SGC terminal aesthetic, theme system, animations)
- [x] Phase 4 — Drone tactics (Spread/Focus/Perimeter targeting modes)
- [x] Phase 5 — Galaxy events (surges, bounties, milestones, New Game+)
- [ ] Phase 6 — Specializations, ascension, ASCII art mastery
