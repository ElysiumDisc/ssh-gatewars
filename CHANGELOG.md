# Changelog

All notable changes to SSH GateWars will be documented in this file.

---

## [0.1.0] — 2026-03-06

### Added

**Phase 0 — Proof of Concept**
- Go project scaffold with Wish SSH server + Bubbletea TUI
- SSH server accepts connections on configurable port (default 2222)
- Auto-generates ed25519 host key on first run
- Accepts all SSH public keys for identity (fingerprint-based player tracking)
- Procedural starfield background (400x200 world space, `.` `*` `+` stars)
- Viewport system mapping world coordinates to any terminal size
- Graceful terminal resize handling
- `q` / `ctrl+c` to disconnect cleanly
- Signal handling for graceful server shutdown

**Phase 0+1 — Simulation Engine**
- 10 tick/sec server-side simulation running in dedicated goroutine
- Ship entities with position, velocity, HP, damage, faction, state
- 5 faction definitions with distinct colors, directional symbols, base stats
- 5 Stargates arranged in pentagon formation as spawn points
- Ship spawning with configurable rates per faction
- Ship steering AI — seek nearest enemy with faction-specific behaviors:
  - Tau'ri: formation clustering (steer toward nearby allies)
  - Goa'uld: slow turning rate (advancing wall)
  - Jaffa: speed increases near enemies (battle fury)
  - Lucian: random velocity noise (erratic weaving)
  - Asgard: wide spacing force (steer away from allies)
- Grid-based spatial hash (20x20 cells) for O(n) neighbor queries
- Combat resolution with attack range and faction modifiers:
  - Tau'ri formation bonus (+6% per nearby ally, max +30%)
  - Goa'uld shield matrix (rear ships take 50% less damage)
- 4-frame explosion animations on ship death
- Soft-bounce world boundary clamping

**Phase 1 — The Living War**
- Faction selection screen with live player counts and territory bars
- SSH key fingerprint-based player identity (SHA256 hash)
- Anonymous connection support (session-based ID)
- Player registration/unregistration affects faction spawn rates
- Underdog bonus (+15% damage for factions with <50% players of leader)
- Territory zone system (40x20 grid, recalculated every 1 second)
- Territory background tinting with dim faction colors
- HUD with faction stats, ship count, territory %, kills/deaths, controls
- Max 2000 ships cap to prevent memory exhaustion

**Rendering**
- Frame builder compositing starfield + territory + ships + explosions
- Faction-colored directional ship symbols (multi-character rendering)
- Run-length style grouping for efficient Lipgloss styling
- Per-session Lipgloss renderer (correct SSH color profiles)
- 15 FPS render polling via Bubbletea tick commands
- Minimum terminal size check (60x15)

**Architecture**
- Single-writer/multi-reader concurrency (RWMutex on engine state)
- Snapshot-based rendering (sessions get immutable state copies)
- Shared starfield (generated once, read by all sessions)
- Clean session lifecycle with deferred cleanup on disconnect
