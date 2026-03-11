# Development Guide

## Prerequisites

- **Go 1.22+** — [install](https://go.dev/dl/)
- A terminal with 256-color support (most modern terminals)
- An SSH client (OpenSSH, PuTTY, etc.)

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
make vet      # Run go vet
make clean    # Remove binary
```

On first launch, the server generates an SSH host key pair (`gatewars_host_key`, `gatewars_host_key.pub`) and creates the SQLite database.

```bash
./gatewars --port 2222 --seed 42 --db dev.db
```

## Testing Locally

```bash
# Terminal 1 — server
make run

# Terminal 2 — client
ssh -p 2222 -o StrictHostKeyChecking=no localhost
```

Multiple clients can connect simultaneously. Each SSH key gets its own character.

## Project Structure

```
cmd/server/main.go          Entry point — flag parsing, wiring, server start
internal/
  core/
    config.go               GameConfig — network, world gen, combat tuning
    types.go                Vec2, Rect, Pos, Direction
    session.go              SessionInfo (per-connection metadata)
  simulation/
    engine.go               10 Hz tick loop, per-planet ticking, enemy AI, projectiles
    action.go               PlayerAction types (move, interact, dial, pickup, fire, reload)
    snapshot.go             Per-planet immutable snapshots (tiles, enemies, projectiles)
    planetinstance.go       Active planet lifecycle (load, tick, unload, projectile spawning)
    galaxy.go               Galaxy-level state, planet registry
    campaign.go             Campaign / game session management
    colony.go               Colony state (legacy, retained for future)
    tech.go                 Tech state (legacy, retained for future)
  gamedata/
    tiles.go                Tile types (floor, wall, door, gate, cover values, opacity)
    biomes.go               Biome templates with tile distributions and enemy spawns
    items.go                25+ weapons, armor, consumables, ammo types, crafting materials
    enemies.go              20+ enemy types with AI behaviors, ranged stats, faction tags
    gateaddresses.go        39-glyph alphabet, named addresses, hash function
    loot.go                 Faction-specific loot tables (jaffa, kull, ori, replicator, etc.)
    factions.go             NPC reputation factions (Tok'ra, Free Jaffa, Asgard)
    hulls.go                Ship hull definitions (legacy)
    components.go           Ship component definitions (legacy)
  world/
    tilemap.go              TileMap — 2D grid, collision, tile lookup
    generator.go            BSP room generator, corridor carving, entity placement
    sgc.go                  Fixed SGC hub layout (gate room, armory, briefing, infirmary, mess)
    gate.go                 Stargate dialing, address validation, address→seed mapping
    los.go                  Bresenham line-of-sight, LOS checks, ray casting
  entity/
    character.go            Player character — HP, level, XP, position, equipment, ammo
    enemy.go                NPC enemy — 8-state AI (idle, patrol, alert, chase, attack, flee, regroup, stunned)
    inventory.go            Item container, equip slots, weight calculation
  combat/
    damage.go               Bump-attack damage calc, death, XP rewards
    ranged.go               Ranged attack resolution — range, LOS, accuracy, cover penalty
    cover.go                Cover calculation — per-tile cover values, attacker-defender geometry
    projectile.go           Projectile entity — precomputed Bresenham path, per-tick advance
  chat/
    hub.go                  Chat message hub — goroutine with channel-based routing
    channel.go              Channel management (#ops, #local, #sg-team, @DM)
    message.go              ChatMessage type, formatting, timestamps
    commands.go             Slash commands (/help, /tune, /roster, /who, /callsign, /team, etc.)
    walter.go               Walter NPC — announces gate events, arrivals, departures, level ups
  server/
    ssh.go                  Wish SSH server, Bubbletea middleware
    identity.go             SSH key → fingerprint, display name extraction
    limiter.go              Token-bucket rate limiter, per-key session caps
  store/
    migrations.go           SQLite schema (characters, inventory, addresses, factions, chat, teams)
    player.go               Character CRUD, inventory, discovered addresses, chat history
  tui/
    model.go                Bubbletea Model — state machine, Update, View, chat integration
    state.go                ViewState enum, FocusTarget (game/chat)
    keybinds.go             Key → action mapping (movement, fire, reload, chat, etc.)
    termcaps.go             Terminal capability detection
    views/
      splash.go             Title screen with stargate ASCII art
      dhd.go                Circular DHD — 3 concentric glyph rings, lit-up dialed symbols
      planet.go             Planet exploration tile map with fog of war
      aimmode.go            Aim mode overlay — targeting reticle, LOS line, range display
      hud.go                HUD bar (HP, weapon, ammo, location, threat level)
      inventory.go          Inventory/equipment modal
      chatpanel.go          Chat panel overlay (hidden/compact/expanded), toast notifications
      starmap.go            Astroterm-inspired gate network browser with constellation lines
```

## Architecture Overview

### Concurrency Model

```
┌─────────────┐     actions      ┌──────────────┐    atomic     ┌──────────────┐
│  TUI Session │ ──────────────▶ │    Engine     │ ──────────▶  │ PlanetSnap   │
│  (reader)    │  chan(1000)      │ (single writer)│  Pointer    │  (immutable)  │
└─────────────┘                  └──────────────┘              └──────────────┘
       ▲                                                              │
       └──────────────────── reads (lock-free) ◀──────────────────────┘

┌──────────────┐    chan         ┌──────────────┐
│  Chat Hub    │ ◀────────────  │  TUI Session  │
│  (goroutine) │  outbox(200)   │  chatConnect  │
└──────────────┘                └──────────────┘
       │ broadcasts                    ▲
       └───────────────────────────────┘
```

- **Single writer:** The engine goroutine ticks all active planet instances at 10Hz. Only it mutates game state.
- **Multi reader:** TUI sessions read the snapshot for their current planet. No locks needed.
- **Planet lifecycle:** Instances loaded on first dial, unloaded when all players leave (unless persistent).
- **Chat hub:** Separate goroutine with channel-based message routing. Each TUI session has a buffered outbox channel.

### Engine Tick (10Hz)

1. **Drain actions** — Process all queued player actions (move, interact, dial, pickup, fire, reload)
2. **Tick planets** — For each active planet instance:
   - Enemy AI state machine (patrol → alert → chase → attack → flee)
   - Enemy ranged attacks (LOS check, projectile spawning)
   - Stun/cooldown countdowns
3. **Tick projectiles** — Advance all projectiles along precomputed paths, check collisions (walls, enemies, players), apply damage
4. **Publish snapshots** — Atomic pointer swap per-planet for lock-free reader access
5. **Cleanup** — Unload empty non-persistent planets periodically

### Combat System

**Melee:** Walk into an enemy tile to bump-attack. Damage = weapon power - armor defense. Enemies bump-attack back during their AI tick.

**Ranged:** Press `f` to enter aim mode. Move targeting reticle with WASD/arrows. Green LOS line = clear shot, red = blocked. Enter to fire. Projectiles travel across the map at weapon-specific speed.

**Cover:** Tiles between attacker and defender provide cover bonuses that reduce accuracy:
- Wall: 75%, Half wall: 50%, Crate/Console: 40%, Tree: 30%, Rubble: 25%, Pillar: 75%, Altar: 60%

**Enemy AI states:** Idle → Patrol → Alert (heard something) → Chase → Attack (melee or ranged) → Flee (low HP) → Regroup → Stunned

### Chat System

Devzat-inspired chat with Hub goroutine for message routing:
- **Channels:** #ops (global), #local (planet-scoped), #sg-team (team channel), @DM (direct messages)
- **Walter NPC:** Announces gate activations, player arrivals/departures, level ups, deaths
- **SG teams:** Create, invite, leave, kick, disband — each team gets a chat channel
- **Slash commands:** /help, /tune, /roster, /who, /callsign, /me, /dm, /mute, /unmute, /motd, /clear, /team, /iris, /indeed, /kree, /shol'va
- **TUI integration:** Chat panel overlay with 3 states (hidden/compact/expanded), focus management (game/chat), toast notifications

### Star Map

Astroterm-inspired gate network browser (`m` key):
- **Positioning:** Each gate address seed deterministically generates (x, y) coordinates in world space. Named planets (Earth, Abydos, etc.) have fixed iconic positions forming a constellation pattern.
- **Background:** Procedural star field (dim `.` and `·` dots) seeded from a fixed seed for consistency.
- **Constellation lines:** Named planets connected by dotted lines (Earth↔Abydos, Earth↔Chulak, etc.).
- **Star rendering:** Glyph varies by threat level (∗ → ✦ → ★ → ✹ → ✵), colored by biome theme. Named planets use ◉.
- **Navigation:** Arrow keys pan, +/- zoom, Tab/Shift+Tab cycle through stars, Enter dials selected address.
- **Info panel:** Shows selected star's name, biome, address (glyph + numeric), and threat bar.

### Adding a New Feature

**New enemy type:** Add to `internal/gamedata/enemies.go`, update biome spawn tables in `biomes.go`, add loot table in `loot.go`.

**New weapon:** Add to `internal/gamedata/items.go` with range, accuracy, ammo type, projectile glyph/color. Add ammo type if needed.

**New item:** Add to `internal/gamedata/items.go`, add to loot tables in `loot.go`.

**New player action:** Add type to `simulation/action.go`, handle in `engine.go`, trigger from `tui/model.go`.

**New TUI view:** Add `ViewState` in `tui/state.go`, add render function in `views/`, wire in `model.go`.

**New chat command:** Add handler in `chat/commands.go`, register in the command dispatch table.

**New persistence:** Add migration to `store/migrations.go`, add CRUD methods to `store/player.go`.

## Debugging

```bash
# Deterministic world generation
./gatewars --seed 12345

# Check SQLite database
sqlite3 gatewars.db "SELECT * FROM characters;"
sqlite3 gatewars.db "SELECT * FROM inventory;"
sqlite3 gatewars.db "SELECT * FROM discovered_addresses;"
sqlite3 gatewars.db "SELECT * FROM chat_messages;"
sqlite3 gatewars.db "SELECT * FROM teams;"
```

The server logs to stderr via `charmbracelet/log`.

## Roadmap

- [x] Phase 1 — SGC hub, gate dialing, procedural planets, bump combat, loot, persistence
- [x] Phase 2 — Chat system, SG teams, Walter NPC, multiplayer visibility
- [x] Phase 3 — Ranged combat, aim mode, cover system, projectiles, 25+ weapons, 20+ enemies
- [ ] Phase 4 — Tech tree, faction reputation, crafting, SGC lab
- [ ] Phase 5 — Named planets (Abydos, Chulak, Dakara...), missions, server-wide events
- [ ] Phase 6 — Specializations, perks, ascension, ASCII art polish
