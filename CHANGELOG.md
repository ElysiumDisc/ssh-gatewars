# Changelog

All notable changes to SSH GateWars are documented here.

## v2.3.0 — Star Map (2026-03-11)

### Added
- **Star map** — `m` to open astroterm-inspired gate network browser
- **Starfield rendering** — procedural background stars with depth layers
- **Constellation lines** — named planets connected by dotted lines forming a Stargate constellation
- **Star glyphs by threat** — ∗ (low), ✦ (medium), ★ (high), ✹ (extreme), ◉ (named)
- **Biome-colored stars** — each star colored by its planet's biome type
- **Pan & zoom** — WASD/arrows to pan, +/- to zoom, Tab/Shift+Tab to cycle stars
- **Star info panel** — planet name, biome, gate address (symbols + code), threat bar
- **Quick dial from map** — Enter on selected star to dial directly
- **Current location indicator** — your planet highlighted in green
- **Named planet positions** — Earth, Abydos, Chulak, Tollana, Cimmeria, Dakara, Langara, Atlantis at fixed iconic positions

---

## v2.2.0 — Phase 3: Combat Overhaul (2026-03-11)

### Added
- **Ranged combat** — `f` to enter aim mode with targeting reticle, LOS line, Enter to fire
- **Projectile system** — projectiles travel across the map with per-weapon glyphs and colors
- **Line of sight** — Bresenham ray casting for ranged attacks and enemy detection
- **Cover system** — walls (75%), crates/consoles (40%), trees (30%), rubble (25%) reduce accuracy
- **Aim mode TUI** — targeting reticle, green/red LOS indicator, range display
- **25+ weapons** — full catalog across 5 tech origins (Earth, Goa'uld, Ancient, Asgard, Ori)
  - Earth: Combat Knife, M9, P-90, USAS-12, M249 SAW, C-4
  - Goa'uld: Staff Weapon, Zat, Kara Kesh, Pain Stick, Shock Grenade
  - Ancient: Drone, Hand Weapon, Anti-Replicator Gun
  - Asgard: Plasma Beam
  - Ori: Staff, Stun Weapon
- **Ammo types** — 11 ammo types (9mm, 5.7mm, naquadah charges, zat charges, etc.)
- **20+ enemy types** — full bestiary with faction affiliations
  - Goa'uld: Jaffa Warrior, Serpent Guard, Horus Guard, Kull Warrior, Ashrak, Commander, System Lord
  - Replicator: Bug, Soldier, Human-Form
  - Ori: Soldier, Prior, Commander
  - Wildlife: Unas, Giant Scarab, Crystal Entity, Sodan Warrior
- **Enemy ranged AI** — enemies fire projectiles, seek cover, flee at low HP
- **AI state machine** — Idle, Patrol, Alert, Chase, Attack, Flee, Regroup, Stunned states
- **New tile types** — half walls, pillars, altars, inscriptions, vents
- **Crafting materials** — naquadah samples, ancient data pads, kull fragments, replicator shards
- **Expanded loot tables** — per-faction loot (jaffa, serpent guard, kull, replicator, ori, wildlife, crates)

## v2.1.0 — Phase 2: Chat & Multiplayer (2026-03-11)

### Added
- **Chat system** — devzat-inspired with Hub goroutine, channel-based routing
- **Chat channels** — #ops (global), #local (planet-scoped), #sg-team, @DM
- **Walter NPC** — announces gate activations, arrivals, departures, level ups, deaths
- **Slash commands** — /help, /tune, /roster, /who, /callsign, /me, /dm, /mute, /unmute, /motd, /clear, /team, /iris, /indeed, /kree, /shol'va
- **SG teams** — create, invite, leave, kick, disband with team chat channels
- **Chat panel overlay** — 3 states (Hidden/Compact/Expanded), composited over game view
- **Focus management** — FocusGame/FocusChat routing for keyboard input
- **Toast notifications** — chat messages appear as toast when chat panel is hidden
- **Player list modal** — Tab to view online players with callsign, level, location
- **Chat persistence** — messages saved to SQLite, backlog delivery on channel join
- **Mute system** — per-player mute lists persisted to DB
- **Rate limiting** — 5 messages per 3 seconds
- **DHD redesign** — circular ASCII art with 39 glyphs in concentric rings, lit-up dialed symbols

## v2.0.0 — The Reboot (2026-03-11)

Complete reimagining from 4X strategy to multiplayer roguelike/MUD.

### Added
- Stargate network with 7-symbol gate addresses and 39-glyph alphabet
- Procedural planet generation (BSP rooms, biomes, threat levels)
- SGC home base with fixed layout (gate room, armory, briefing room, infirmary, mess hall)
- Bump-to-attack combat system (walk into enemies to fight)
- Real-time 10Hz simulation (enemies patrol and hunt independently)
- Per-planet instance management (loaded on-demand, unloaded when empty)
- Character system (HP, level, XP, equipment, inventory)
- Fog of war (per-player exploration tracking)
- Tile-based world rendering with Lipgloss colors
- DHD gate dialing interface with address input
- Address book for discovered gate addresses
- Character persistence (stats, inventory, addresses saved to SQLite)

### Removed
- 4X strategy systems (colonies, economy, tech trees, faction selection)
- Galaxy map view, colony management, scoreboard
- Ship design, fleet command, diplomacy stubs

---

## v1.0.0 — Layer 1: 4X Strategy (2026-03-10)

Full 4X rewrite with colonies, economy, and tech trees. (Superseded by v2.0.0)

---

## v0.5.0 — Phase 2: Powers, Persistence, Multiplex Views (2025)

5 factions, SQLite persistence, multiple views, rate limiting.

---

## v0.1.0 — Initial Scaffold (2025)

SSH server with Wish + Bubbletea, basic galaxy generation.
