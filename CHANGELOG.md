# Changelog

All notable changes to SSH GateWars will be documented in this file.

---

## [2.0.0] — 2026-03-08

### Changed — Complete Rewrite: 4X Strategy Game (Master of Orion Clone)

**The game has been completely redesigned from a cooperative PvE Replicator defense game into a full 4X strategy game inspired by Master of Orion (1993), themed as Stargate SG-1.** Players now explore, expand, exploit, and exterminate across a procedurally generated galaxy.

### Added

**Procedural Galaxy Generation**
- Configurable galaxy size (20-100 systems, default 50) via `--systems` flag
- Reproducible galaxies via `--seed` flag
- 5 homeworlds placed in pentagonal arrangement: Earth, Chulak, Othala, Vorash, Cal Mah
- Dakara: center system, Ancient superweapon, Huge Terran Ultra-Rich planet
- ~30 named SG-1 canon planets: Abydos, Langara, Tollana, Heliopolis, Edora, etc.
- Procedural P#X-### names for remaining systems
- 5 star types with weighted distribution: Yellow (35%), Red (25%), Blue (15%), White (15%), Binary (10%)
- Stargate network: k=3 nearest neighbors, bidirectional, connectivity ensured via BFS
- 90% of systems have planets with random type, size, and mineral richness
- ~5% of systems have Ancient artifact specials

**Planet System (13 types)**
- 7 habitable types: Terran, Ocean, Jungle, Arid, Steppe, Tundra, Desert
- 6 hostile types requiring tech: Barren, Volcanic, Toxic, Inferno, Dead, Radiated
- 5 sizes affecting max population: Tiny (2), Small (4), Medium (7), Large (10), Huge (14)
- 5 mineral richness levels affecting production: Ultra-Poor (0.33x) → Ultra-Rich (1.67x)

**Colony Management**
- 5 production sliders per colony (Ship, Defense, Industry, Ecology, Research) summing to 100%
- Production formula: `min(pop, factories) * mineralMult * factionMod * onlineBonus`
- Online bonus: +5% per player, capped at +50%
- Industry slider builds factories (cost 10, cap = population * 10 + faction bonus)
- Defense slider builds missile bases (cost 50)
- Ecology slider cleans waste (population generates waste at 0.1%/s)
- Research slider feeds faction research pool across 6 tech trees
- Ship slider accumulates toward build queue (ready for Layer 2)
- Responsive local slider editing with auto-rebalancing (Research adjusts as catch-all)
- Population growth: logistic model (0.5%/s base rate, slowing near max capacity)

**Faction Traits (MOO-style)**
- Tau'ri: Diplomacy +30%, Research +10%, special: cheaper treaties
- Goa'uld: Production +20%, Attack +20%, Research -15%, special: +3 factory cap per pop
- Asgard: Research +50%, Ground -30%, pop growth -15%, special: cheaper miniaturization
- Tok'ra: Espionage +50%, Production -20%, special: sabotage & intel
- Jaffa Free Nation: Ground +50%, Defense +10%, Research -20%, special: cheaper invasions

**6 Tech Trees**
- SGC Systems, Goa'uld Engineering, Asgard Shields, Ancient Knowledge, Hyperdrive Tech, Weapons
- Research allocation adjustable via tech tree screen (6-way slider summing to 100%)
- Tier advancement with exponential RP costs
- Miniaturization formula: -5% cost/size per tier above component requirement (min 50%)

**New Rendering**
- Galaxy map: character-grid rendering with colored star dots, dim gate connection lines, faction-colored colonies
- Directional navigation between systems (nearest system in arrow direction)
- System view: star type, planet details (type/size/minerals), colony summary
- Colony management screen: 5 interactive slider bars with real-time production rates
- Tech tree browser: 6 colored allocation bars with tier and RP display
- Diplomacy view: 5x5 faction relation matrix with colored status indicators
- New Stargate ASCII art splash screen with faction selection and trait display
- Context-sensitive HUD: campaign status, faction name, Naquadah, system count, online players
- Scoreboard: faction comparison table (systems, population, naquadah, online, avg tech tier)

**Session State Machine (10 states)**
- factionSelect → galaxyMap (hub) → systemView → colonyManage
- galaxyMap → techTree, diplomacyView, shipDesigner (stub), scoreboard, help
- All screens: Esc → parent, q → quit
- Galaxy map: Enter/t/d/s/Tab/? navigate to sub-screens

**Game Data Package**
- `internal/gamedata/planettypes.go`: 13 planet types, 5 sizes, 5 mineral richness, max pop calculation
- `internal/gamedata/techtree.go`: 6 trees, tier cost scaling, miniaturization formula
- `internal/gamedata/hulls.go`: 4 hull sizes (Al'kesh → City-ship) with space/HP/cost
- `internal/gamedata/components.go`: starting weapons, shields, armor, engines, computers

**Stubs for Future Layers**
- `fleet.go`: Fleet struct with movement state (gate/hyperspace), ready for Layer 2
- `ship.go`: ShipDesign struct with hull/components, ready for Layer 2
- `combat.go`: Combat/CombatGroup structs, ready for Layer 3
- `diplomacy.go`: DiplomacyState with relation matrix and proposals, ready for Layer 5
- `shipdesign.go`, `combatview.go`: placeholder screens

**Infrastructure**
- New `--seed` flag for reproducible galaxy generation
- New `--systems` flag for configurable galaxy size (20-100)
- Updated HTTP `/stats` endpoint with 4X data model (factions, systems, colonies, campaign)
- Simplified SQLite schema (players table, campaign tables)

### Removed
- Replicator swarm system (swarm types, infection mechanics, wave progression)
- PvE cooperative tower-defense gameplay model
- Fixed 55-planet galaxy with concentric rings
- Defense actions (turret/shield/wall/cleanse/repair with resource costs)
- Resource system (shared Naquadah + per-faction Trinium)
- Patron system (faction ownership of planets)
- Wave timing, difficulty scaling, campaign victory/defeat (Earth-based)
- `replicator.go`, `effects.go`, `gateart.go`, `planetpanel.go`
- Old migration files (`001_initial.sql`, `002_galaxy.sql`, `003_revamp.sql`)
- Lucian Alliance faction (replaced by Tok'ra)
- Jaffa Rebellion name (renamed to Jaffa Free Nation)

### Changed
- `factions.go` — 5 factions with full MOO-style trait modifiers (diplomacy, research, production, attack, defense, ground, pop growth, espionage, factory cap bonus)
- `engine.go` — complete rewrite: economy tick loop (population growth, production splitting, factory building, waste, research accumulation), player action processing (slider/tech allocation validation)
- `galaxy.go` — complete rewrite: procedural generation with Poisson disc placement, k-nearest gate network, named SG-1 planets
- `session.go` — complete rewrite: 10-state machine, colony slider editing with local state, tech allocation, flash messages
- `hud.go` — complete rewrite: campaign status, faction resources, system count, controls hint
- `identity.go` — simplified schema (players table only)
- `stats.go` — new response shape: factions with systems/population/naquadah, campaign state
- `main.go` — new Engine constructor with seed/systemCount, new flags
- `ssh.go` — removed Starfield from Server struct (kept for future combat view)

---

## [1.0.0] — 2026-03-07

### Changed — Full Revamp: Cooperative Replicator Defense

**The game was redesigned from PvP faction warfare to PvE cooperative tower-defense against the Replicators.**

### Added

**Galaxy Foundation (Phase 1)**
- 55 canon Stargate planets organized in concentric rings
- 90 Stargate edges connecting planets
- ASCII galaxy map renderer with navigation
- Planet detail panel with defense stats

**Replicator Infection System (Phase 2)**
- 6 swarm types unlocking by wave: Basic, Shielded, Fast, Regen, Queen, Horde
- Infection mechanics (0.0 → 1.0), wave progression, campaign lifecycle

**Player Actions (Phase 3)**
- 8 defense actions: Deploy Turret, Upgrade Turrets, Build Shield, Reinforce Wall, Station Fleet, Launch Fleet, Cleanse, Repair
- Fleet movement along Stargate edges
- Dual resource system (Naquadah shared, Trinium per-faction)

**Persistence**
- SQLite schema: campaigns, planet state, player contributions

### Removed
- PvP faction-vs-faction ship combat
- Half-block pixel battlefield renderer (kept infrastructure)
- Ship spawning, territory calculation, faction powers

---

## [0.6.0] — 2026-03-07

### Added — Phase 3: Strategic Galaxy Layer + Rendering Overhaul
- Half-block pixel rendering with doubled vertical resolution
- 8-planet galaxy with Stargate network, dual resource economy
- Galaxy map view + planet detail panels

---

## [0.5.0] — 2026-03-06

### Added — Phase 2: Powers & Interaction
- 5 faction powers, beam rendering, ship trails
- Tab view cycling, network map, help overlay
- SQLite persistence, HTTP stats API, security hardening

---

## [0.1.0] — 2026-03-06

### Added — Phase 0+1: Initial Scaffold
- Go project scaffold with Wish SSH server + Bubbletea TUI
- 10 tick/sec simulation engine with ship combat
- 5 faction definitions, procedural starfield, viewport system
