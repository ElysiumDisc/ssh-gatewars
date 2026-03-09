# SSH GateWars — Development Guide

> v1.0.0 — Master of Orion Clone (Stargate SG-1 Themed)

---

## Table of Contents

1. [Dev Environment Setup](#dev-environment-setup)
2. [Core Concept](#core-concept)
3. [Architecture](#architecture)
4. [The Galaxy](#the-galaxy)
5. [Colony Economy](#colony-economy)
6. [Technology](#technology)
7. [Ship Design & Fleets](#ship-design--fleets)
8. [Combat](#combat)
9. [Diplomacy & Victory](#diplomacy--victory)
10. [Visual Design](#visual-design)
11. [Security](#security)
12. [Implementation Layers](#implementation-layers)

---

## Dev Environment Setup

### Prerequisites

- Ubuntu (or any Debian-based Linux)
- Git installed (`sudo apt install git`)
- SSH key configured for GitHub (`ssh -T git@github.com` should work)

### Step 1: Install Go

```bash
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz -O /tmp/go.tar.gz
mkdir -p ~/go-sdk
tar -C ~/go-sdk -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz
mkdir -p ~/go
```

Add to `~/.bashrc`:

```bash
export PATH=$HOME/go-sdk/go/bin:$HOME/go/bin:$PATH
export GOPATH=$HOME/go
```

Then: `source ~/.bashrc && go version`

### Step 2: Clone and Build

```bash
cd ~
git clone git@github.com:ElysiumDisc/ssh-gatewars.git
cd ssh-gatewars
go mod download
go build -o gatewars ./cmd/server/
```

### Step 3: Run and Connect

```bash
# Terminal 1: Start server
./gatewars --port 2222

# Terminal 2: Connect
ssh -p 2222 localhost
```

Press 1-5 to pick a faction. You'll see the galaxy map with ~50 procedural star systems.

### Quick Reference

| What | Command |
|------|---------|
| Build | `go build -o gatewars ./cmd/server/` |
| Run | `./gatewars --port 2222` |
| Run (custom galaxy) | `./gatewars --port 2222 --seed 42 --systems 60` |
| Connect | `ssh -p 2222 localhost` |
| Run vet | `go vet ./...` |
| Format code | `gofmt -w .` |
| Tidy deps | `go mod tidy` |

---

## Core Concept

SSH GateWars is a **multiplayer 4X strategy game** rendered over SSH, inspired by Master of Orion (1993) and themed as Stargate SG-1. Players SSH in, join one of 5 factions, and cooperatively manage their faction's galactic empire.

### Design Pillars

1. **4X depth** — explore, expand, exploit, exterminate through a full economy/tech/combat loop
2. **Cooperative factions** — every online player strengthens their faction's output
3. **Persistent campaigns** — colonies, tech, fleets persist across sessions
4. **Zero friction** — SSH in, pick faction, manage your empire
5. **Standby mode** — simulation pauses when nobody's online

### In-Universe Lore

> Replicator-Colonel-Carter coded this galactic conquest simulator and embedded
> it in SGC terminals as part of the Stargwent universe. What started as a
> strategic planning tool became the most-played program on the base network.

### Multiplayer Model

Each player joins 1 faction. Each online player adds to faction output (production, research). Any faction member can manage any colony, design ships, command fleets. Last-write-wins for sliders; build/research orders are additive. More players online = stronger faction.

---

## Architecture

### Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| Language | Go 1.26 | Goroutines, single binary, fast |
| SSH Server | [Wish](https://github.com/charmbracelet/wish) | Battle-tested SSH app framework |
| TUI | [Bubbletea](https://github.com/charmbracelet/bubbletea) | Elm architecture, diff-based rendering |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal colors, layout |
| Persistence | SQLite (modernc.org/sqlite) | No CGO, pure Go, WAL mode |
| Stats API | net/http | REST JSON endpoint |

### Goroutine Architecture

```
main goroutine
  ├── Engine.Run() goroutine (10 ticks/sec, owns simulation state)
  ├── Wish SSH server goroutine
  │     ├── Session 1 (Bubbletea model, reads Engine.Snapshot())
  │     ├── Session 2 ...
  │     └── ... (per-session idle timeout monitors)
  ├── HTTP stats API goroutine (:8080/stats)
  └── ConnLimiter (rate limiting + session caps)
```

### Concurrency Model

- **Engine** is single-writer: holds `sync.RWMutex`, tick loop takes write lock
- **Sessions** are readers: `Snapshot()` takes read lock, returns immutable copy
- **Action queue**: players enqueue actions (mutex-protected append), engine drains each tick
- Galaxy is immutable after generation — no synchronization needed

### Project Structure

```
ssh-gatewars/
  cmd/server/main.go                    -- entry point, flags, Wish setup, signal handling
  internal/
    server/
      ssh.go                            -- SSH key extraction, Server struct
      session.go                        -- per-player Bubbletea model, ~10 states
      limiter.go                        -- connection rate limiting, session caps
    faction/
      factions.go                       -- 5 factions with MOO-style racial traits
    gamedata/
      planettypes.go                    -- 13 planet types, 5 sizes, 5 mineral richness
      techtree.go                       -- 6 tech trees, tier costs, miniaturization
      hulls.go                          -- 4 hull sizes (Al'kesh → City-ship)
      components.go                     -- weapons, shields, engines, computers
    simulation/
      engine.go                         -- tick loop, economy phases, FactionState
      galaxy.go                         -- procedural star system generation, gate network
      colony.go                         -- population growth, factory building, 5 sliders
      tech.go                           -- research accumulation, tier unlocking
      campaign.go                       -- campaign lifecycle
      actions.go                        -- player action queue
      snapshot.go                       -- immutable snapshot types for renderers
      fleet.go                          -- fleet movement (stub for Layer 2)
      ship.go                           -- ship design (stub for Layer 2)
      combat.go                         -- tactical combat (stub for Layer 3)
      diplomacy.go                      -- treaties, council (stub for Layer 5)
      vec.go                            -- Vec2 math helpers
    render/
      galaxymap.go                      -- star system map with gate lines, faction colors
      systemview.go                     -- system detail (planet info, colony summary)
      colonyview.go                     -- colony slider management, build queue
      splash.go                         -- faction selection with Stargate ASCII art
      hud.go                            -- bottom status bar: campaign, resources, controls
      scoreboard.go                     -- faction standings comparison table
      techview.go                       -- tech tree browser + allocation sliders
      diplomacyview.go                  -- faction relations matrix
      shipdesign.go                     -- ship designer (stub for Layer 2)
      combatview.go                     -- tactical combat view (stub for Layer 3)
      frame.go                          -- PixelBuffer/HalfBlockCell (kept for combat view)
      starfield.go                      -- procedural starfield (kept for combat view)
      viewport.go                       -- world-to-screen mapping (kept for combat view)
    player/
      identity.go                       -- SQLite persistence (players table)
    httpapi/
      stats.go                          -- HTTP JSON stats API (/stats endpoint)
  db/migrations/
    001_moo_schema.sql                  -- complete schema reference
  go.mod, go.sum
  README.md, CHANGELOG.md, DEVELOPMENT.md, LICENSE
```

---

## The Galaxy

### Procedural Generation

Each campaign generates a galaxy from a seed (reproducible with `--seed`):

1. **5 homeworlds** placed in a pentagonal arrangement with jitter (Earth, Chulak, Othala, Vorash, Cal Mah)
2. **Dakara** placed near center (Ancient superweapon, Huge/Ultra-Rich)
3. **Remaining systems** (default ~44) placed via rejection sampling with minimum distance
4. **Gate network**: k=3 nearest neighbors per system, bidirectional, connectivity ensured via BFS
5. **Planets**: 90% of systems have a planet with random type, size, and mineral richness

### Star Types

| Type | Color | Frequency |
|------|-------|-----------|
| Yellow | #FFE030 | 35% |
| Red | #FF6030 | 25% |
| Blue | #6090FF | 15% |
| White | #FFFFFF | 15% |
| Binary | #FFCC00 | 10% |

### Planet Types (13)

**Habitable** (colonizable immediately):

| Type | Pop Modifier | Color |
|------|-------------|-------|
| Terran | 1.0x | Green |
| Ocean | 0.9x | Blue |
| Jungle | 0.9x | Dark green |
| Arid | 0.8x | Tan |
| Steppe | 0.8x | Yellow-green |
| Tundra | 0.7x | Light blue |
| Desert | 0.6x | Gold |

**Hostile** (requires tech to colonize):

| Type | Pop Modifier | Requires |
|------|-------------|----------|
| Barren | 0.4x | Controlled Environment |
| Volcanic | 0.3x | Controlled Environment |
| Toxic | 0.3x | Controlled Environment |
| Dead | 0.3x | Controlled Environment |
| Inferno | 0.2x | Advanced Environment |
| Radiated | 0.2x | Advanced Environment |

### Planet Sizes

| Size | Max Pop | Frequency |
|------|---------|-----------|
| Tiny | 2 | 10% |
| Small | 4 | 20% |
| Medium | 7 | 30% |
| Large | 10 | 25% |
| Huge | 14 | 15% |

### Mineral Richness

| Richness | Production Mult | Frequency |
|----------|----------------|-----------|
| Ultra-Poor | 0.33x | 5% |
| Poor | 0.50x | 15% |
| Normal | 1.00x | 45% |
| Rich | 1.33x | 20% |
| Ultra-Rich | 1.67x | 15% |

### Special Systems

- **Dakara**: Ancient superweapon system. Huge Terran planet, Ultra-Rich minerals. Central position.
- **Artifact worlds**: ~5% chance per system. Ancient outpost with research bonuses.

---

## Colony Economy

### The 5 Sliders

Every colony has 5 production sliders that sum to 100%:

```
┌─ COLONY: Earth ────────────────────────────────┐
│                                                 │
│  Population: 5.0 / 7    Factories: 10 / 50     │
│  Output: 5.0/s          Waste: 0.1              │
│                                                 │
│  PRODUCTION SLIDERS                             │
│                                                 │
│> Ship       [░░░░░░░░░░░░░░░░░░░░]   0%  0.0/s │
│  Defense    [██░░░░░░░░░░░░░░░░░░]  10%  0.5/s │
│  Industry   [██████░░░░░░░░░░░░░░]  30%  1.5/s │
│  Ecology    [████░░░░░░░░░░░░░░░░]  20%  1.0/s │
│  Research   [████████░░░░░░░░░░░░]  40%  2.0/s │
│                                                 │
└─────────────────────────────────────────────────┘
```

### Production Math (per second)

```
baseOutput   = min(population, factories) * mineralMultiplier * factionProdMod
onlineBonus  = 1.0 + min(onlinePlayers * 0.05, 0.50)  // +5%/player, cap +50%
totalOutput  = baseOutput * onlineBonus

shipOutput     = totalOutput * sliderShip / 100
defenseOutput  = totalOutput * sliderDefense / 100
industryOutput = totalOutput * sliderIndustry / 100
ecologyOutput  = totalOutput * sliderEcology / 100
researchOutput = totalOutput * sliderResearch / 100
```

### Slider Effects

| Slider | Target | Cost per Unit | Notes |
|--------|--------|---------------|-------|
| Ship | Build queue | Varies by design | Ships built from queue |
| Defense | Missile bases | 50 per base | Unlimited bases |
| Industry | Factories | 10 per factory | Cap = population * 10 (+faction bonus) |
| Ecology | Waste cleanup, terraform | — | Cleans waste first, then increases max pop |
| Research | Faction RP pool | — | Split by tech allocation across 6 trees |

### Population Growth

Logistic growth model: fast when population is low, slowing near max capacity.

```
growth = 0.5% * population * (1 - population/maxPop) * factionGrowthMod
```

### Starting Colony (each homeworld)

- Population: 5.0
- Max Pop: 7 (Terran Medium)
- Factories: 10
- Max Factories: 50
- Naquadah: 200
- Default sliders: Ship 0%, Defense 10%, Industry 30%, Ecology 20%, Research 40%

---

## Technology

### 6 Tech Trees

| Tree | Color | Description |
|------|-------|-------------|
| SGC Systems | Blue | Colony infrastructure, controlled environments |
| Goa'uld Engineering | Red | Factory efficiency, production bonuses |
| Asgard Shields | Teal | Shield technology, defensive systems |
| Ancient Knowledge | Gold | Terraforming, advanced colonization |
| Hyperdrive Tech | Purple | Fleet speed, warp capabilities |
| Weapons | Orange | Ship weapons, attack power |

### Research Mechanics

- All colonies' Research slider output → faction RP pool
- RP pool split by 6-way allocation (adjustable via tech tree screen)
- Tier unlock when accumulated RP exceeds cost: exponential scaling starting at 100
- **Miniaturization**: each tier above a component's required tier reduces size/cost by 5%, minimum 50%
- Faction trait modifiers apply (Asgard +50% research, Goa'uld -15%, etc.)

---

## Ship Design & Fleets

> Layer 2 — not yet implemented, stubs in place

### 4 Hull Sizes

| Hull | SG Name | Space | HP | Base Cost |
|------|---------|-------|----|-----------|
| Small | Al'kesh | 20 | 3 | 10 |
| Medium | Ha'tak | 60 | 18 | 50 |
| Large | O'Neill-class | 120 | 60 | 200 |
| Huge | City-ship | 250 | 150 | 600 |

### Fleet Movement

- **Stargate**: instant travel between connected systems
- **Hyperspace**: slower, can reach any system (speed depends on engine tech)
- Max 6 active ship designs per faction

---

## Combat

> Layer 3 — not yet implemented, stubs in place

### Stack-Based Tactical Combat

- Ships grouped by design into stacks
- 1 combat round per engine tick (10 rounds/sec)
- Each stack fires at opposing stacks: `damage = weaponDPS * attackMod / defenseMod`
- Shields absorb first, then armor, then hull HP
- Auto-resolve option computes all rounds instantly
- Typical combat: 5-30 seconds real-time
- Half-block pixel rendering for visual combat (infrastructure preserved from v0.6)

---

## Diplomacy & Victory

> Layer 5 — not yet implemented, stubs in place

### Treaty System

| Status | Effect |
|--------|--------|
| None | Default between all factions |
| War | Open hostility, fleet combat enabled |
| NAP | Non-Aggression Pact — no attacks |
| Trade | Trade agreement — shared income bonus |
| Alliance | Full alliance — shared vision, coordinated defense |

### Victory Conditions

- **Conquest**: one faction controls all colonies
- **Galactic Council**: periodic vote, 2/3 population majority wins a peaceful victory
- Campaign ends → new galaxy generated automatically

---

## Visual Design

### Rendering Approaches

| Screen | Method | Key Visual |
|--------|--------|------------|
| Faction Select | Character grid | Stargate ASCII art + faction list with traits |
| Galaxy Map | Character grid | Stars as colored dots, gate links as dim lines, selected system bracketed |
| System View | Character grid | Planet info (type, size, minerals), colony summary |
| Colony Manage | Character grid | 5 slider bars `[████░░░░]`, production rates, build queue |
| Tech Tree | Character grid | 6 colored bars, tier levels, RP progress |
| Diplomacy | Character grid | 5x5 relation matrix, colored status indicators |
| Scoreboard | Character grid | Faction comparison table with key metrics |
| Combat View | Half-block pixels | Ship sprites, beams, explosions (Layer 3) |

### Color Palette

```
Tau'ri:      fg=#4A90D9  bg=#0A1520   — Steel blue
Goa'uld:     fg=#CC2222  bg=#180000   — Red
Asgard:      fg=#40E0D0  bg=#001414   — Cyan teal
Tok'ra:      fg=#C850C0  bg=#100014   — Purple magenta
Jaffa:       fg=#E8C820  bg=#151400   — Bright yellow

Star types:  Yellow #FFE030, Red #FF6030, Blue #6090FF, White #FFFFFF, Binary #FFCC00
Gate links:  #333333 (dim dots)
Selection:   #FFFFFF (bright brackets)
```

### Frame Rate & Bandwidth

- Simulation: 10 ticks/sec
- Rendering: 15 FPS per session
- Economy updates: 1/sec (every 10th tick)
- Bandwidth: ~1-3 KB/s per player (character-grid rendering)

---

## Security

### SSH Layer
- No shell access — Bubbletea model only
- Accept all SSH keys for identity, no password auth
- Host key auto-generated, stored outside repo

### Rate Limiting
- Max 500 concurrent sessions total
- Max 10 sessions per SSH key
- 10 new connections/sec globally
- 30 min idle timeout (view-only sessions exempt)

### Input Validation
- Whitelist-only key handling (arrows, hjkl, Enter, Esc, Tab, t/d/s/?, q, 1-5)
- SSH commands matched against fixed registry ("scoreboard", "galaxy")
- Action queue validated: slider sums must equal 100, faction ownership verified

### Data Protection
- Parameterized SQL queries only
- SHA256 fingerprints only (no raw keys, no PII)
- SQLite with WAL mode, 0600 file permissions

---

## Implementation Layers

### Layer 1: Galaxy + Colonies + Economy (Done)
- [x] Procedural galaxy generation (Poisson disc placement, k-nearest gate network)
- [x] Colony struct with 5 sliders, population growth, factory building
- [x] Tick engine with economy phases (production, industry, ecology, research, defense)
- [x] Galaxy map rendering (character grid with gate lines and faction coloring)
- [x] System view + colony management screens
- [x] Faction selection with MOO-style traits
- [x] Session state machine (10 states)
- [x] Database schema + player persistence
- [x] HTTP stats API

### Layer 2: Ship Design + Fleets (Planned)
- [ ] Ship designer screen + design validation
- [ ] Fleet creation, movement (gate instant / hyperspace timed)
- [ ] Fleet command overlay on galaxy map
- [ ] Ship build queue integration with colony Ship slider

### Layer 3: Combat (Planned)
- [ ] Stack-based combat engine (round per tick)
- [ ] Combat detection (opposing fleets at same system)
- [ ] Colony bombardment + ground invasion
- [ ] Half-block pixel combat view (ship sprites, beams, explosions)

### Layer 4: Technology (Planned)
- [ ] 50+ tech definitions across 6 trees
- [ ] Component gating (require tech tier to unlock)
- [ ] Miniaturization (cheaper/smaller components at higher tiers)
- [ ] Colony improvements (factory cap, terraforming, controlled environments)

### Layer 5: Diplomacy + Victory (Planned)
- [ ] Treaty system (NAP, trade, alliance, war)
- [ ] Proposal/vote system within factions
- [ ] Galactic Council vote (2/3 population wins)
- [ ] Victory detection + campaign regeneration

### Layer 6: Polish + Specials (Planned)
- [ ] Dakara system (guarded, major bonus)
- [ ] Artifact worlds, planet specials
- [ ] Tok'ra espionage (enemy intel, sabotage)
- [ ] Starbases at colonies
- [ ] Balance tuning

---

## Verification Checklist

1. `go build ./cmd/server` — compiles
2. `go vet ./...` — no issues
3. SSH in → faction select → see galaxy map with procedural stars
4. Navigate systems, enter owned colony, adjust sliders → see production numbers change
5. Research accumulates across tech trees with allocation adjustments
6. Scoreboard shows correct faction statistics
7. Disconnect and reconnect → faction persists from SQLite
8. Multiple concurrent SSH sessions → player counts update in real-time
