# Changelog

## v3.4.0 — Astrologic View, Stargate Network & Faction Deepening (2026-03-13)

Galaxy exploration overhaul with two new views, interactive Stargate network mechanics, and deeply differentiated faction paths.

### New — Astrologic Star Map (`a` key)
- **2D galaxy projection** with all planets positioned in a golden spiral
- **Constellation lines** drawn along Stargate network routes between planets
- **Region labels** — Alpha Quadrant, Pegasus Sector, Ori Territory, Asgard Reach
- **Pan/zoom** — arrow keys to pan, +/- to zoom, camera tracks across the galaxy
- **Twinkling** — liberated planets shimmer, surging planets pulse with warning
- **Deploy direct** — select a planet and Enter to deploy from the star map
- Grid-based rendering (same engine as defense field)

### New — Stargate Network Tube Map (`n` key)
- **Subway-style topology** — planets as stations, connected by color-coded route lines
- **6 named routes** — Milky Way Core, Pegasus Rim, Ori Frontier, Asgard Reach, Nox Passage, Tok'ra Circuit
- **Route legend** on wide terminals, auto-scroll to selected station
- **Station info** — connection count, link levels, cumulative network bonuses
- **Deploy direct** — select a station and Enter to deploy

### New — Gate Link Upgrades (`u` in Network view)
- **3 upgrade levels** per gate link: 50 / 150 / 400 ZPM
- **Defense bonuses** from upgraded links: shield regen, damage boost, spawn reduction
- **Faction-modified** — Ancient get 1.5x shield regen from links, Ori get 2x damage boost
- **Ori discount** — gate upgrades cost 20% less for Ori players
- Persisted in new `gate_links` SQLite table

### New — Resource Transfers (`s` in Network view)
- **Shield Boost** (30 ZPM) — +20 HP to all chairs on a planet (Ancient: +30 HP)
- **Drone Boost** (50 ZPM) — +2 temporary bonus drones for 60s (Ori: 90s)
- **ZPM Gift** (25 ZPM) — adds 25 to planet's bounty pool
- Enables cooperative play from Atlantis — support active defenders remotely
- Logged in new `resource_transfers` SQLite table

### New — Deepened Faction Differentiation

**Ancient Path — Network Masters:**
- Gate shield regen 1.5x, gate spawn reduction 1.25x
- Shield transfers 50% more effective
- +10% ZPM earnings from kills
- **Ascension Pulse** passive (chair level 5+): heals all friendly chairs +3 HP every 5 seconds
- **Adaptive Targeting**: drones retarget mid-flight to track moving enemies

**Ori Path — Divine Firepower:**
- Gate damage bonus 2x (divine wrath through the network)
- Gate upgrades 20% cheaper (zealotry efficiency)
- Drone boost transfers last 50% longer
- **Prior's Wrath** passive (chair level 5+): AOE 2 damage to all replicators within radius 8 every 8 seconds
- **Locked Trajectory**: drones commit to initial target vector (no retargeting)

### New — Cross-Navigation
- Galaxy (`g`), Astro (`a`), and Network (`n`) views are cross-navigable from each other
- All three accessible from Atlantis hub
- Atlantis bottom bar updated with new key hints

### Changed
- Planet positions now stored in data model (previously calculated then discarded)
- Network topology auto-generated from planet positions (MST + short-edge redundancy)
- Throne view shows passive ability info, gate affinity stats, and drone targeting mode
- Network bonuses applied at defense tick time (shield regen, damage boost, spawn reduction)
- Bonus drones from transfers expire after TTL countdown

### Database
- Migration 9: `gate_links` table (from_id, to_id, level, upgraded_by, upgraded_at)
- Migration 10: `resource_transfers` table (sender_fp, target_planet_id, bonus_type, amount)

### New Files
- `internal/game/network.go` — GateLink, GateRoute, GalaxyNetwork, transfer/upgrade constants
- `internal/tui/views/astro.go` — AstroModel + RenderAstro
- `internal/tui/views/network.go` — NetworkModel + RenderNetwork
- `internal/tui/views/draw.go` — Bresenham line drawing, grid string helpers
- `internal/store/network.go` — Gate link and transfer persistence

---

## v3.3.0 — Factions, Tactics & Galaxy Events (2026-03-13)

Full drone upgrade overhaul, Ancient vs Ori factions, drone tactics, galaxy events, New Game+.

### Changed — ZPM Economy
- **Bug ZPM**: 1 → 5
- **Sentinel ZPM**: 3 → 15
- **Queen ZPM**: 10 → 50
- Upgrade costs unchanged (Swift 100, Blast 250, Piercing 500, Chair 50-550)

### New — Drone Overhaul
- **Aggressive drone scaling** — Ancients: 5 + 4/level (up to 45 drones), Ori: 3 + 2/level (up to 23)
- **Salvo firing** — chairs fire 1-4 drones per shot based on level (salvos scale by faction)
- **Fire rate scaling** — higher chair level = faster firing (down to 0.3-0.4s intervals)
- **Per-owner drone counting** — each player's drones tracked independently (fixes shared-pool bug)
- **Faction-adjusted damage** — Ori drones deal 2x damage; Ancients have +25% shields

### New — Factions (Ancient vs Ori)
- **Ancient Path** — drone swarm masters: more drones, stronger shields, balanced damage
- **Ori Path** — devastating firepower: fewer drones, 2x damage, faster fire rate, weaker shields
- **Faction switching** via Throne — resets all upgrades (fresh build)
- **Visual identity** — Ancients pulse cyan, Ori pulse orange
- **Persistent** — faction stored in SQLite, survives reconnect

### New — Drone Tactics (Phase 4)
- **Spread** `[1]` — target nearest enemies (default)
- **Focus** `[2]` — all drones focus strongest threat (queens first)
- **Perimeter** `[3]` — prioritize enemies closest to center
- Switchable during defense via number keys, shown in HUD

### New — Galaxy Events (Phase 5)
- **Replicator Surges** — random invaded planets get surge flag: 2x spawns, 2x ZPM
- **Planet bounties** — each planet has a bounty (invasion_level × 10 ZPM) awarded on liberation
- **Liberation milestones** — server-wide announcements at 25%, 50%, 75%, 100% galaxy freed
- **New Game+** — when 100% galaxy is liberated, all planets reset with scaled-up difficulty
- **Cycle counter** — galaxy map shows current threat cycle

### New — Player Reset
- **Full reset** via Throne — zeroes ZPM, chair level, drone tier, faction
- **Faction switch** — switching path also resets upgrades (respec)

### Changed
- Defense HUD shows current tactic, surge indicator, and bounty
- Galaxy view shows bounty ZPM, surge warnings, liberation percentage, cycle number
- Throne view shows power preview (next level stats), faction badge, drone damage per tier
- Engine event bridge uses proper type mapping (fixes misrouted chat announcements)
- DroneFireRate default: 15→10 ticks (faster base firing)
- DB migration adds `faction` column to players table

### Fixed
- Event bridge between engine and chat used direct integer cast causing wrong announcements
- Drone count was shared across all players with same tier (now per-owner)

---

## v3.2.0 — Upgrade Throne (2026-03-13)

Ancient Control Chair upgrade terminal — spend ZPM to power up.

### New
- **Throne view** (`views/throne.go`) — full-screen upgrade terminal accessible from Atlantis via `[t]`
- **Chair upgrades** — spend ZPM to level up shield generator (Lv0→10), each level grants +1 drone slot and +5 shield HP
- **Drone tier unlocks** — Swift (100 ZPM), Blast (250 ZPM), Piercing (500 ZPM)
- **Animated throne** — pulsing cyan glow on control chair art, timed status messages on upgrade success/failure
- **Cost scaling** — chair cost = (level+1) x 50 ZPM; higher levels cost more
- **State machine** — new `StateThrone` between Atlantis and Galaxy in flow

### Changed
- Atlantis bottom bar now shows `[t] Throne` key hint

---

## v3.1.0 — SGC Terminal Aesthetic (2026-03-13)

Visual overhaul: centralized theme system, animated screens, atmospheric rendering.

### New
- **Centralized theme** (`views/theme.go`) — True Color palette (Ancient cyan `#00D9FF`, ZPM gold `#FFD700`), pre-built styles, entity style lookup tables for performance
- **Layout helpers** — `Center()`, `SideBySide()`, `PanelBox()`, `RoundedBox()`, `FormatKeyHint()`, `ProgressBar()`, `ShieldBar()` — eliminate ad-hoc padding throughout
- **Animated splash** — multi-phase boot sequence: typewriter title reveal → subtitle → gold tagline → progress bar → pulsing prompt, full-screen centered
- **Callsign terminal** — biometric identification feel, double-line outer panel, rounded input sub-box, pulsing cursor
- **Atlantis hub** — responsive split layout via `lipgloss.JoinHorizontal`, double-line outer frame, commander status in rounded box, upgraded chair ASCII art, online player count in top bar
- **Galaxy sensor display** — Unicode status symbols (`●` invaded / `◆` contested / `✧` free), gold-bg highlighted selection, right detail panel on wide terminals (>90 cols), colored aggregate stats
- **Defense field overhaul** — 3 concentric defense rings (dim→bright), crosshair at center, background star dots, entity glyphs per type (`●`/`■`/`◉` replicators, `✸`/`✦`/`►` drones), queen pulsing red animation, shield bars under chairs (green→yellow→red gradient)
- **Chat panel** — rounded border, word-wrap on plain text before styling (fixes ANSI-escape break bug), message-type coloring (system=gold, announce=green, whisper=purple)
- **Frame counter** — `frameCount` field in TUI model for animation (queen pulse, cursor blink)

### Fixed
- Chat word-wrap no longer breaks mid-ANSI-escape sequence
- Defense field no longer allocates `lipgloss.NewStyle()` per cell per frame (uses pre-built style lookup tables)

---

## v3.0.0 — Ancient Defense Network (2026-03-13)

Complete reboot from roguelike to cooperative tower defense.

### New
- **Radial defense view** — chair at center, replicators approach from all directions, drones intercept
- **Ancient Control Chair** — each player deploys a chair as their defense unit
- **Drone system** — 4 tiers (Standard/Swift/Blast/Piercing) with auto-targeting and homing
- **Replicator enemies** — 3 types (Bug/Sentinel/Queen) with scaling difficulty waves
- **Galaxy map** — shared universe with 50 planets under replicator invasion
- **Atlantis hub** — personal base showing stats, upgrades, ASCII chair art
- **Hold timer** — 5 min per player to liberate a planet (co-op scales time)
- **ZPM economy** — earn from kills, persist across sessions (roguelite loop)
- **Planet liberation events** — server-wide announcements when planets are freed
- **Chat system** — carried forward with channel routing, Walter NPC, slash commands, SG teams, DMs
- **Animated splash** — line-by-line reveal with blinking prompt
- **Callsign system** — unique persistent callsigns with rename support

### Architecture
- 10Hz defense engine with per-planet instances and wave spawning
- Single-writer chat hub with channel-based routing (unchanged)
- SQLite persistence for player progression and galaxy state
- Lock-free galaxy snapshots via atomic.Value
- Radial coordinate system for defense field rendering

### Removed
- Tile-based exploration, BSP generation, fog of war
- Roguelike combat (melee, ranged, aim mode, cover)
- 25+ weapons catalog, 20+ enemy bestiary (replaced with drone/replicator system)
- Inventory system, equipment slots
- DHD circular dial interface, gate address system
- SGC hub layout

---

## v2.2.0 — Combat Overhaul (archived)

Previous roguelike version. See git tag `v2.2.0-archive` for full history.
Includes: ranged combat, aim mode, cover system, projectiles, 25+ weapons, 20+ enemies,
star map, DHD, chat system, SG teams, gate address network.
