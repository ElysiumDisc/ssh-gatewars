# Changelog

All notable changes to SSH GateWars will be documented in this file.

---

## [0.5.0] — 2026-03-06

### Added

**Phase 2 — Powers & Interaction**
- 5 faction powers with shared cooldowns and state machine (Ready → Active → Cooldown)
  - Tau'ri: Coordinated Strike — all ships lock same target for 5s
  - Goa'uld: Bombardment — ships freeze, deal 2x damage for 5s
  - Jaffa: Kree! — max speed, ignore incoming damage for 8s
  - Lucian: Kassa Rush — +50% attack speed, +25% damage taken for 10s
  - Asgard: Ion Cannon — piercing beam across battlefield, instant
- Space bar activates faction power with cross-session notifications
- Power cooldown bar in HUD (visual progress bar replaces static text)
- Sector focus voting (keys 1-5 bias spawn heading toward target sector)
- Tau'ri coordinated strike target override in combat system
- Beam rendering (Bresenham line drawing with directional characters ═/║)
- Ship trails (3-position fading trail behind moving ships)
- Power visual effects:
  - Jaffa Kree!: bright yellow ship highlighting when boosted
  - Lucian Kassa Rush: alternating purple/red background tint
  - Goa'uld Bombardment: ships freeze in place
  - Asgard Ion Cannon: cyan beam line across battlefield with area damage
- Tab view cycling: battlefield → scoreboard → network map → stats
- Network map view (territory overview with colored sector blocks + gate positions)
- Help overlay with controls and faction power description (? key)

**Faction-as-Username Login**
- `ssh tauri@sgc.games` skips faction selection, jumps to battlefield
- Supports all 5 faction names as SSH usernames

**Multiplex Views**
- `ssh sgc.games scoreboard` — full-screen live faction scoreboard
- `ssh sgc.games network` — stargate network territory map
- `ssh sgc.games stats` — personal stats and session info
- Same SSH key = 1 player for spawn rate (sessions deduplicated)
- Primary session has full controls, additional sessions are view-only

**SQLite Persistence**
- Player identity saved across sessions (SSH fingerprint → faction)
- Returning players auto-join their previous faction
- Faction stats table for lifetime kills/deaths tracking
- Pure Go SQLite (modernc.org/sqlite, no CGO required)
- WAL mode for concurrent read/write access

**HTTP Stats API**
- JSON endpoint at `/stats` on configurable address (default 127.0.0.1:8080)
- Returns live faction data: players, ships, territory %, kills, deaths
- CORS headers for web integration
- Read-only, no mutations exposed

**Security Hardening**
- Connection rate limiter (token bucket, configurable connections/sec)
- Maximum concurrent session cap (default 500)
- Per-SSH-key session cap (default 10, prevents abuse)
- Idle timeout (30 min default, view-only sessions exempt)
- Whitelist-only input handling (q, space, 1-5, tab, ?)
- SSH commands matched against fixed registry map
- Parameterized SQL queries only
- DB file created with owner-only permissions

**Build & Operations**
- Makefile with build/run/clean targets
- Configurable server flags: --db, --http, --max-sessions, --max-per-key, --connect-rate, --idle-timeout
- Graceful shutdown handles server close without fatal errors
- spawner.go with spawn configuration constants

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
