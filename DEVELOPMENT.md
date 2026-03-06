# SSH GateWars — Development Guide

> v0.1.0 — Phase 0+1 Complete, Phase 2 In Progress

---

## Table of Contents

1. [Dev Environment Setup](#dev-environment-setup)
2. [Core Concept](#core-concept)
3. [Architecture](#architecture)
4. [The Five Fleets](#the-five-fleets)
5. [Simulation Engine](#simulation-engine)
6. [Player Presence System](#player-presence-system)
7. [Interaction System](#interaction-system)
8. [SGC Command System — Planets](#sgc-command-system--planets)
9. [Visual Design](#visual-design)
10. [Seasons & Scoring](#seasons--scoring)
11. [The Stargwent Bridge](#the-stargwent-bridge)
12. [Security](#security)
13. [Roadmap](#roadmap)

---

## Dev Environment Setup

Fresh Ubuntu laptop? Follow these steps to get everything running.

### Prerequisites

- Ubuntu (or any Debian-based Linux)
- Git installed (`sudo apt install git`)
- SSH key configured for GitHub (`ssh -T git@github.com` should work)

### Step 1: Install Go (no sudo needed)

Go installs to two folders in your home directory:

```
~/go-sdk/go/    <-- GOROOT: the Go compiler, tools, standard library
~/go/           <-- GOPATH: where Go downloads modules and builds binaries
```

Download and extract Go (check https://go.dev/dl/ for latest version):

```bash
# Download Go
wget https://go.dev/dl/go1.26.1.linux-amd64.tar.gz -O /tmp/go.tar.gz

# Extract to ~/go-sdk (no sudo needed)
mkdir -p ~/go-sdk
tar -C ~/go-sdk -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

# Create GOPATH directory
mkdir -p ~/go
```

### Step 2: Add Go to your PATH

Add these lines to the bottom of `~/.bashrc`:

```bash
# Go
export PATH=$HOME/go-sdk/go/bin:$HOME/go/bin:$PATH
export GOPATH=$HOME/go
```

Then reload:

```bash
source ~/.bashrc
```

### Step 3: Verify Go works

```bash
go version
# Should print: go version go1.26.1 linux/amd64
```

### Step 4: Clone the project

```bash
cd ~
git clone git@github.com:ElysiumDisc/ssh-gatewars.git
cd ssh-gatewars
```

### Step 5: Download dependencies

```bash
go mod download
```

This pulls all Charm libraries (Wish, Bubbletea, Lipgloss) into `~/go/pkg/mod/`.

### Step 6: Build the server

```bash
go build -o gatewars ./cmd/server/
```

This creates a single `gatewars` binary (~9.5 MB) in the project root.

### Step 7: Run the server

```bash
# First run auto-generates an SSH host key in .ssh/
./gatewars --port 2222
```

You should see:

```
INFO Starting SSH GateWars host=0.0.0.0 port=2222
INFO Connect with: command="ssh -p 2222 localhost"
```

### Step 8: Connect and test

Open a **second terminal** and connect:

```bash
ssh -p 2222 localhost
```

First time you'll get a host key prompt — type `yes`. You should see the faction selection screen.

Press `1-5` to pick a faction and enter the battlefield. Press `q` to disconnect.

### Step 9: Stop the server

Back in the server terminal, press `Ctrl+C` for graceful shutdown.

### Quick Reference

| What | Command |
|------|---------|
| Build | `go build -o gatewars ./cmd/server/` |
| Run | `./gatewars --port 2222` |
| Connect | `ssh -p 2222 localhost` |
| Run tests | `go test ./...` |
| Format code | `gofmt -w .` |
| Tidy deps | `go mod tidy` |
| Add a dep | `go get github.com/some/package@latest` |

### Directory Structure After Setup

```
~/
├── go-sdk/go/              <-- Go compiler (GOROOT) — don't modify
│   ├── bin/go, gofmt
│   ├── src/                <-- Go standard library source
│   └── pkg/                <-- Compiled standard library
│
├── go/                     <-- Go workspace (GOPATH)
│   ├── bin/                <-- Installed Go binaries (tools)
│   └── pkg/mod/            <-- Downloaded module cache
│       ├── github.com/charmbracelet/wish@v1.4.7/
│       ├── github.com/charmbracelet/bubbletea@v1.3.10/
│       └── ...
│
└── ssh-gatewars/           <-- This project
    ├── cmd/server/main.go
    ├── internal/...
    ├── go.mod, go.sum
    ├── gatewars             <-- Compiled binary (gitignored)
    └── .ssh/id_ed25519     <-- Auto-generated host key (gitignored)
```

### Troubleshooting

**`go: command not found`** — Your PATH isn't set. Run `source ~/.bashrc` or check that the export lines are in `~/.bashrc`.

**`ssh: connect to host localhost port 2222: Connection refused`** — Server isn't running. Start it with `./gatewars --port 2222`.

**`WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED`** — The host key changed (maybe you deleted `.ssh/`). Fix with: `ssh-keygen -R "[localhost]:2222"`

**`go mod download` is slow** — First download pulls ~50MB of deps. Subsequent builds use the cache in `~/go/pkg/mod/`.

**Port already in use** — Another instance is running, or another service uses 2222. Use a different port: `./gatewars --port 3333` and `ssh -p 3333 localhost`.

---

## Core Concept

SSH GateWars is a **persistent ambient space battle** rendered entirely in the terminal over SSH. Five factions wage an eternal war across a shared starfield. Ships spawn, fly, fight, and die — endlessly, whether anyone is watching or not.

### Design Pillars

1. **Beauty first** — the terminal output must be gorgeous. Smooth movement, color, particle effects.
2. **Zero friction** — SSH in, pick faction, you're in. No tutorial needed.
3. **Presence is participation** — being connected IS playing.
4. **Faction identity** — each fleet looks, moves, and fights differently.
5. **Standalone and bridged** — works via SSH, also accessible from Stargwent.

### In-Universe Lore

> **SGC TACTICAL NETWORK — CLASSIFIED**
>
> *"After the Battle of Antarctica, I started thinking about fleet coordination
> across the gate network. We needed a way to model multi-faction engagements
> in real-time — something any SGC terminal could connect to. So I built it."*
> — **Dr. Samantha Carter**, Chief Science Officer, SGC
>
> Originally developed as a tactical simulation by **Lt. Col. Samantha Carter**
> and the **SGC technical staff** to model fleet movements across the Stargate
> Network, the system quickly evolved beyond its original purpose. What started
> as a dry strategic planning tool became something the entire base couldn't
> stop watching — a living, breathing war simulation that ran 24/7 on the SGC's
> Deep Space Telemetry servers.
>
> Technicians on the night shift started "dialing in" from their terminals to
> watch the fleets clash. Then someone figured out you could bias your faction's
> spawns by pressing a key. Then Walter started keeping score between shifts.
> Before long, half the SGC was running it in a corner of their screen during
> downtime, quietly rooting for their faction.
>
> General Landry authorized it as "morale software." Sgt. Siler insisted on
> keeping the Lucian Alliance faction in because "you need a wildcard."
> Teal'c has never once switched off Jaffa Rebellion.
>
> Now declassified and connected to the public Stargate Network, anyone with a
> terminal can dial in. The simulation never stops. The war never ends.

---

## Architecture

### Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| Language | Go | Goroutines, single binary, fast |
| SSH Server | [Wish](https://github.com/charmbracelet/wish) | Battle-tested SSH app framework |
| TUI | [Bubbletea](https://github.com/charmbracelet/bubbletea) | Elm architecture, diff-based rendering |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal colors, layout |
| Persistence | SQLite | Players, seasons, stats |
| Stats API | net/http | REST for Stargwent integration |

### Goroutine Architecture

```
main goroutine
  ├── Engine.Run() goroutine (10 ticks/sec, owns simulation state)
  ├── Wish SSH server goroutine
  │     ├── Session 1 (Bubbletea model, reads Engine.Snapshot())
  │     ├── Session 2 ...
  │     └── ...
  └── HTTP stats API (future)
```

### Concurrency Model

- **Engine** is single-writer: holds `sync.RWMutex`, tick loop takes write lock
- **Sessions** are readers: `Snapshot()` takes read lock, returns immutable copy
- Player registration/power activation use brief mutex-protected methods
- Starfield is immutable after generation — no synchronization needed

### Key Design Decisions

1. **One simulation, many viewers** — simulation runs once, all players observe same state
2. **Simulation runs without players** — war is always in progress when you connect
3. **Diff-based rendering** — Bubbletea diffs `View()` output, sends only changes over SSH
4. **Per-session renderer** — `bubbletea.MakeRenderer(s)` for correct SSH color profiles
5. **Snapshot-based rendering** — sessions get immutable copies, no shared mutable state
6. **Shared cooldowns** — faction powers have one timer per faction, any player triggers

### Project Structure

```
ssh-gatewars/
  cmd/server/main.go               -- entry point, Wish setup, signal handling
  internal/
    server/
      ssh.go                        -- SSH key extraction, server struct
      session.go                    -- per-player Bubbletea model, state machine
    simulation/
      engine.go                     -- tick loop, spatial hash, spawn, move, combat
      ship.go                       -- ship entity, Vec2, distance helpers
      territory.go                  -- zone ownership, territory calculation
      faction_power.go              -- active abilities, cooldowns (Phase 2)
    render/
      starfield.go                  -- procedural background generation
      viewport.go                   -- world-to-screen coordinate mapping
      frame.go                      -- frame builder: starfield + ships + effects
      hud.go                        -- bottom status bar
      effects.go                    -- explosions, beams (Phase 2)
    faction/
      factions.go                   -- 5 faction definitions, colors, symbols, stats
    player/
      identity.go                   -- SQLite persistence (Phase 1.5)
  db/migrations/
    001_initial.sql                 -- schema (Phase 1.5)
  .gitignore
  go.mod, go.sum
  README.md, CHANGELOG.md, DEVELOPMENT.md
```

---

## The Five Fleets

Each faction has a **distinct visual identity**, **movement pattern**, and **unique power**. You should be able to tell factions apart at a glance by shape, color, and behavior.

### Tau'ri — The Coordinated

```
Color:    Steel blue (#4A90D9)
Ships:    F-302 interceptors  ->  <-  /\  \/
Stats:    HP 100, Damage 10, Speed 2.0
Spawn:    Normal rate
```

- **Movement**: Fly in **formation** — cluster toward nearby allies
- **Passive**: Formation Bonus — +6% damage per nearby ally within 3 tiles (max +30%)
- **Active**: Coordinated Strike — all ships lock onto same target for 5s (45s cooldown)
- **Weakness**: Area attacks scatter whole formations
- **Fantasy**: *"We fight as one. Our discipline is our strength."*

### Goa'uld — The Shielded

```
Color:    Gold amber (#D4A017)
Ships:    Ha'tak motherships  {=>  <=}  /^\  \v/
Stats:    HP 120, Damage 8, Speed 1.2
Spawn:    Normal rate
```

- **Movement**: Slow advancing **wall**, low steering rate
- **Passive**: Shield Matrix — ships behind front row take 50% less damage
- **Active**: Bombardment — ships pause, deal 2x damage for 5s (45s cooldown)
- **Weakness**: Slow. Fast factions run circles around them
- **Fantasy**: *"Kneel before your god. We advance, unstoppable."*

### Jaffa Rebellion — The Zealous

```
Color:    Bright yellow (#E8C820)
Ships:    Al'kesh bombers  >>  <<  ^^  vv
Stats:    HP 80, Damage 9, Speed 2.5
Spawn:    Normal rate
```

- **Movement**: **Accelerate** as they approach enemies
- **Passive**: Battle Fury — speed increases as distance to enemy decreases
- **Active**: Kree! — all ships surge at max speed for 8s, ignore damage (45s cooldown)
- **Weakness**: Fragile at range. Kited easily
- **Fantasy**: *"Freedom! We charge, and we do not stop."*

### Lucian Alliance — The Unpredictable

```
Color:    Purple (#C850C0)
Ships:    Modified cargo runners  ~>  <~  ~^  ~v
Stats:    HP 90, Damage 10, Speed 2.2
Spawn:    25% chance to double each wave
```

- **Movement**: **Erratic** — random velocity noise each tick
- **Passive**: Smuggler's Fortune — 25% chance for double spawns
- **Active**: Kassa Rush — +50% attack speed for 10s, but take +25% damage (45s cooldown)
- **Weakness**: Inconsistent. Bad luck = thin fleets
- **Fantasy**: *"You can't predict us. We can't predict us. That's the point."*

### Asgard — The Superior

```
Color:    Cyan (#40E0D0)
Ships:    O'Neill warships  *->  <-*  *|*  *|*
Stats:    HP 300, Damage 30, Speed 1.5
Spawn:    1/3 normal rate (3x stats compensate)
```

- **Movement**: Wide spacing, steer away from allies, precise steering
- **Passive**: Technological Supremacy — each ship worth 3 standard ships
- **Active**: Ion Cannon — piercing beam across screen, hits all enemies in line (30s cooldown)
- **Weakness**: Every ship lost is a major blow. Overwhelmed by numbers
- **Fantasy**: *"We are few. We are enough."*

### Balance Philosophy

Asymmetric strengths in a soft rock-paper-scissors cycle:

```
Tau'ri formations  ──strong vs──>  Jaffa charges (focused fire melts rushers)
Jaffa charges      ──strong vs──>  Goa'uld walls (speed flanks slow ships)
Goa'uld walls      ──strong vs──>  Asgard elites (attrition grinds down few)
Asgard elites      ──strong vs──>  Lucian swarms (beam clears chaff)
Lucian swarms      ──strong vs──>  Tau'ri formations (erratic dodges focus)
```

Player count, active powers, and timing matter more than faction matchup.

---

## Simulation Engine

### Core Loop (10 ticks/sec)

```
Every tick:
  1. SPAWN ships at gates (rate scales with connected players)
  2. MOVE all ships (steering AI + faction-specific behaviors)
  3. REBUILD spatial hash (20x20 grid for O(n) queries)
  4. COMBAT resolution (attack range 5 tiles, faction modifiers)
  5. PROCESS explosions (4-frame death animation)
  6. REMOVE dead ships
  7. TERRITORY recalculation (every 10th tick = 1 second)
  8. EXPIRE notifications
```

### Ship Entity

```go
Ship {
    ID        uint64
    Faction   int
    X, Y      float64     // world position
    VX, VY    float64     // velocity
    HP        float32
    MaxHP     float32
    Damage    float32
    Speed     float32
    State     Alive | Exploding | Dead
    SpawnTick uint64
    Boosted   bool
}
```

### Spatial Hash

- World divided into 20x20-unit cells (20 columns x 10 rows)
- Rebuilt every tick: O(n) insert all alive ships
- Range queries check local cell + 8 neighbors
- Keeps combat resolution O(n) instead of O(n^2)

### Combat Modifiers

| Modifier | Effect |
|----------|--------|
| Tau'ri formation | +6% per ally within 3 tiles (max +30%) |
| Goa'uld shield | Rear ships take 50% less damage |
| Underdog bonus | +15% damage if <50% players of leader |
| Player presence | +1 ship per cycle per player (cap 16) |

### Gate Positions

5 gates arranged in regular pentagon within 400x200 world (radius ~80 from center):

```
        Asgard (top)
       /            \
  Goa'uld          Jaffa
      |              |
  Lucian           Tau'ri
```

### Resource Caps

- Max 2000 ships total
- Spawn interval floor: 0.2 seconds
- Underdog bonus: +15% (not cumulative)

---

## Player Presence System

### Passive Contribution

| Connected Players | Effect |
|-------------------|--------|
| 0 | Base spawn rate. AI equilibrium. |
| 1-5 | +1 ship per spawn cycle per player |
| 6-15 | Diminishing returns approaching cap |
| 16+ | Spawn capped. Excess boosts ship HP +2% each |

**Underdog bonus**: If a faction has <50% players of the leading faction, their ships deal +15% damage.

### Multiplex Sessions

Players can open **multiple SSH sessions** for different views:

```
ssh sgc.games              # Battlefield (primary, full controls)
ssh sgc.games scoreboard   # Live faction scoreboard
ssh sgc.games network      # Stargate network territory overview
ssh sgc.games stats        # Personal stats + season history
```

**Same SSH key = 1 player** for spawn rate. Multiple sessions don't stack bonuses. First session is "primary" (can trigger powers), others are view-only.

### HUD Layout

```
┌─ TAU'RI ── 3 online ── 47 ships ── ████░░░░░░ 22% territory ──────┐
│  Kills: 123  |  Deaths: 98  |  Power: READY [SPACE]                │
│  [1-5] Focus sector  |  [Tab] Views  |  [?] Help  |  [q] Quit     │
└────────────────────────────────────────────────────────────────────-┘
```

---

## Interaction System

### Controls

| Key | Action | Effect |
|-----|--------|--------|
| `Space` | Rally | Trigger faction's active power (shared cooldown) |
| `1-5` | Focus sector | Bias new ship spawns toward a sector |
| `Tab` | Cycle view | Battlefield / scoreboard / network / stats |
| `?` | Help | Controls overlay + faction power description |
| `q` | Disconnect | Close the gate |

### Faction Powers (Phase 2)

Shared cooldown per faction. Any player can trigger. All faction members see notification: *"A fellow TAU'RI has rallied the fleet!"*

| Faction | Power | Duration | Cooldown |
|---------|-------|----------|----------|
| Tau'ri | Coordinated Strike (all lock same target) | 5s | 45s |
| Goa'uld | Bombardment (stop + 2x damage) | 5s | 45s |
| Jaffa | Kree! (max speed, ignore damage) | 8s | 45s |
| Lucian | Kassa Rush (+50% attack, +25% damage taken) | 10s | 45s |
| Asgard | Ion Cannon (piercing beam) | instant | 30s |

### Sector Focus

Keys 1-5 vote for a sector. New ships bias heading toward the voted sector proportionally. Creates emergent coordination without chat.

```
      [2] Asgard sector
     /                \
[1] Goa'uld    [3] Jaffa
    |                |
[4] Lucian    [5] Tau'ri
```

---

## SGC Command System — Planets (Phase 2.5)

SSH supports passing commands: `ssh sgc.games <planet>`. Each planet grants a faction bonus while connected through it.

### Connection Flow

```
$ ssh sgc.games dakara

  Chevron 1... encoded.
  Chevron 2... encoded.
  ...
  Chevron 7... LOCKED.

  ┌───────────────────────────────────┐
  │  WORMHOLE ESTABLISHED            │
  │  Destination: DAKARA             │
  │                                   │
  │  "The weapons of the Ancients    │
  │   belong to all free Jaffa."     │
  │              — Teal'c            │
  │                                   │
  │  Bonus: +25% rally power duration │
  │  Capacity: 12/20 Jaffa connected │
  └───────────────────────────────────┘
```

### Known Planets

| Gate Address | Planet | Bonus |
|-------------|--------|-------|
| `abydos` | Abydos | +15% spawn rate |
| `chulak` | Chulak | +20% ship speed |
| `dakara` | Dakara | +25% rally duration |
| `tollana` | Tollana | +20% ship damage |
| `langara` | Langara | +30% damage, -15% HP |
| `tartarus` | Tartarus | +10% all stats, 2x spawn cooldown |
| `atlantis` | Atlantis | +1 extra ship per wave |
| `p3x-888` | Unas Homeworld | +40% ship HP |
| `nox` | Nox Homeworld | Ships cloak 3s after spawning |
| `ori` | Ori Galaxy | +50% damage, +25% damage taken |

### Secret Planets

| Gate Address | Bonus | Discovery |
|-------------|-------|-----------|
| `kheb` | Rally cooldown halved | Tooltip hint season 3+ |
| `p4x-639` | Ships respawn once | Rare "time loop" event |
| `erebus` | Enemy ships near fleet move 30% slower | Hidden in stats page |
| `vis-uban` | +15% everything | Ancient text puzzle in starfield |
| `othala` | Every 5th ship spawns double-strength | Only if faction holds >40% territory |
| `celestis` | Rally triggers twice | Type address backwards: `sitscelec` |
| `praclarush` | Gate pulses, damages nearby enemies | Appears in MOTD once per season |

### Planet Capacity

| Tier | Capacity | Examples |
|------|----------|---------|
| Common | Unlimited | abydos, chulak, p3x-888 |
| Contested | 20 per faction | tollana, dakara, atlantis |
| Rare | 5 per faction | tartarus, ori, langara |
| Secret | 3 per faction | kheb, othala, celestis |

### Superweapons — SGC Command Unlocks

Certain SGC commands unlock **faction superweapons** — devastating one-time abilities that change the course of the war. These are rare, earned through play, and announced to the entire battlefield when activated.

| Command | Weapon | Effect | Unlock Condition |
|---------|--------|--------|-----------------|
| `ssh sgc.games antarctica` | **Ancient Drone Swarm** | Launches a wave of golden drones from your gate. Each drone (`.` yellow) homes in on the nearest enemy and deals 5x damage on impact. 20 drones total. | Faction holds >50% territory for 5+ consecutive minutes |
| `ssh sgc.games dakara-weapon` | **Dakara Superweapon** | Massive energy wave radiates outward from center. All enemy ships within 80 tiles of world center take 50% of their max HP as damage. Visual: expanding ring of white `=` characters. | Faction controls Dakara planet + >3 players connected through it |
| `ssh sgc.games ori-beam` | **Ori Satellite Beam** | A sustained vertical beam sweeps slowly across the battlefield (left to right over 10 seconds). Destroys any enemy ship it touches. Visual: bright white column `║` moving across screen. | Connect through Ori planet during The Reckoning phase (day 6-7) |
| `ssh sgc.games replicator-swarm` | **Replicator Outbreak** | Spawns 30 neutral replicator ships (`%` grey) that attack ALL factions equally — including yours. Chaotic but clears the field. Ships that survive gain +20% HP. | Faction is in last place for 10+ minutes (desperation move) |
| `ssh sgc.games asgard-legacy` | **Asgard Core Overload** | One random Asgard ship sacrifices itself in a massive explosion (15-tile radius). All enemy ships in range are destroyed. All friendly ships in range are fully healed. | Asgard-only. Requires 5+ Asgard ships alive. |
| `ssh sgc.games horizon` | **Horizon Weapon** | Fires 6 gate-buster warheads at each enemy gate. Each warhead destroys ships within 5 tiles of the gate and disables spawning for that gate for 15 seconds. | Tau'ri-only. Requires coordinated strike to have been used 3 times this season. |
| `ssh sgc.games kull` | **Kull Warriors** | Spawns 5 Kull warrior ships (black `##`, HP 500, damage 20, slow). Nearly unkillable — they have 80% damage resistance. Last 30 seconds then self-destruct. | Goa'uld-only. Requires >8 Goa'uld players online simultaneously. |
| `ssh sgc.games hatak-fleet` | **Ha'tak Armada** | All Goa'uld ships temporarily gain shield visual `[=]` and take 75% less damage for 15 seconds. New ships spawn at 3x rate during this period. | Goa'uld-only. Available once per season when losing by >15% territory. |

**Superweapon rules**:
- Each superweapon has a **once-per-season** cooldown (or specific unlock conditions)
- Activation is announced to ALL players: `"THE TAU'RI HAVE ACTIVATED THE ANCIENT DRONE SWARM!"`
- Visual effects are dramatic and full-screen — everyone sees the destruction
- Superweapons are balanced by rarity and unlock conditions, not by being weak
- The player who triggers it gets credit in season stats ("Superweapon Operator" award)

### Utility Commands

| Command | What It Does |
|---------|-------------|
| `ssh sgc.games` | Join the war |
| `ssh sgc.games help` | Planet list, faction info, controls |
| `ssh sgc.games status` | Your stats: faction, hours, rallies, loyalty |
| `ssh sgc.games planets` | Live planet status: who's where |
| `ssh sgc.games scores` | Faction leaderboard + season standings |
| `ssh sgc.games lore` | In-universe backstory |
| `ssh sgc.games weapons` | Superweapon status: which are available, cooldowns, unlock progress |

---

## Visual Design

### Starfield

- Procedurally generated, static background (400x200 world)
- `.` at 4%, `*` at 0.5%, `+` at 0.1% density
- Territory tinting: dim background color per faction ownership

### Ship Rendering

Faction-specific characters that change with movement direction:

```
Tau'ri:   ->  <-  /\  \/      (2 char, blue)
Goa'uld:  {=>  <=}  /^\  \v/   (3 char, gold)
Jaffa:    >>  <<  ^^  vv       (2 char, yellow)
Lucian:   ~>  <~  ~^  ~v      (2 char, purple)
Asgard:   *->  <-*  *|*  *|*  (3 char, cyan — bigger = more powerful)
```

### Combat Effects

```
Laser fire:   -> - - - ◇       (brief flash between ships)
Explosion:    * → -X- → .:. → . .  (4 frames, ~400ms)
Shield:       [=}  ->  (=}  ->  [=}  (Goa'uld absorbing hit)
Ion Cannon:   *-> ═══════════════>   (cyan beam, 1 second)
Kree! charge: >>  >>  >> becomes >>>>>>>  (ships streak)
```

### Color Palette

```
Tau'ri:    fg=#4A90D9  bg=#0A1520   — Steel blue
Goa'uld:   fg=#D4A017  bg=#151000   — Gold amber
Jaffa:     fg=#E8C820  bg=#151400   — Bright yellow
Lucian:    fg=#C850C0  bg=#100014   — Purple magenta
Asgard:    fg=#40E0D0  bg=#001414   — Cyan teal
Explosions: #FF4444 → #FF8844 → #FFAA44 → dim
Stars:     #555555 (dim) / #AAAAAA (bright) / #FFFFFF (rare)
```

### Terminal Adaptation

| Size | Experience |
|------|-----------|
| 60x15 (minimum) | Compact, minimal HUD |
| 120x40 (typical) | Full HUD, 2-char ships |
| 200x60+ (large) | Spacious, full detail |

### Frame Rate & Bandwidth

- Simulation: 10 ticks/sec
- Rendering: 15 FPS per session
- Bandwidth: ~2-5 KB/s per player (Bubbletea diff rendering)

---

## Seasons & Scoring (Phase 3)

### Season Structure (7 days)

```
Day 1-2:  THE GATHERING  — 20% territory each, initial frontlines
Day 3-5:  THE WAR        — Peak battles, territory swings
Day 6-7:  THE RECKONING  — +50% spawn rates, last hour doubles
```

### Victory

Factions ranked by territory % at season end.

```
Season 14 Results:
  1st  TAU'RI           34.2%  ████████████████░░░░░░
  2nd  JAFFA REBELLION  24.1%  ████████████░░░░░░░░░░
  3rd  GOA'ULD          19.8%  █████████░░░░░░░░░░░░░
  4th  LUCIAN ALLIANCE  13.7%  ██████░░░░░░░░░░░░░░░░
  5th  ASGARD            8.2%  ████░░░░░░░░░░░░░░░░░░
```

### Awards

- **MVP**: Most hours connected + rallies triggered
- **Most Dedicated**: Never switched factions
- **Underdog Award**: Fought outnumbered all season

### Player Stats

- Connection time per faction/season
- Rallies triggered
- Sectors focused
- Faction loyalty streak
- Seasons participated

---

## Running the Server

### Local Development

```bash
cd ~/ssh-gatewars

# Build
go build -o gatewars ./cmd/server/

# Run (auto-generates .ssh/id_ed25519 host key on first start)
./gatewars --port 2222
```

The server starts three things:
1. **Simulation engine** — runs in background goroutine, 10 ticks/sec, war is always happening
2. **SSH server** — listens on port 2222, accepts connections
3. *(future)* **HTTP API** — port 8080, read-only stats for Stargwent

### Connecting

```bash
# From another terminal (or another machine on your LAN)
ssh -p 2222 localhost

# From another machine on your network
ssh -p 2222 your-ip-address
```

### Server Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `2222` | SSH listen port |
| `--host` | `0.0.0.0` | Bind address (0.0.0.0 = all interfaces) |
| `--key` | `.ssh/id_ed25519` | Host key path |

### Production Deployment

For the real `sgc.games` server:

```bash
# Build a release binary
go build -o gatewars ./cmd/server/

# Run on standard SSH port (needs root or capabilities)
sudo setcap 'cap_net_bind_service=+ep' ./gatewars
./gatewars --port 22 --key /etc/ssh-gatewars/host_key

# Or run behind a reverse proxy on port 2222
# and point sgc.games:22 -> localhost:2222
```

Systemd service file (save as `/etc/systemd/system/gatewars.service`):

```ini
[Unit]
Description=SSH GateWars
After=network.target

[Service]
Type=simple
User=gatewars
WorkingDirectory=/opt/ssh-gatewars
ExecStart=/opt/ssh-gatewars/gatewars --port 2222 --key /opt/ssh-gatewars/.ssh/host_key
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

---

## The Stargwent Bridge (Phase 4)

### The Big Picture

SSH GateWars and Stargwent are **two separate projects** that connect through a bridge:

```
┌─────────────────────┐              ┌─────────────────────┐
│  STARGWENT          │              │  SSH GATEWARS        │
│  (Python/Pygame)    │              │  (Go/Wish)           │
│                     │   HTTP API   │                      │
│  Stats menu ───────────────────>   │  GET /api/status     │
│  "DIAL IN" button   │              │  GET /api/player/:id │
│       │             │              │                      │
│       │ subprocess  │   SSH        │                      │
│       └──────────────────────────> │  Port 2222           │
│                     │              │                      │
│  Card game          │              │  Space battle sim    │
│  (offline/LAN)      │              │  (persistent server) │
└─────────────────────┘              └─────────────────────┘
```

They share:
- Same 5 factions (Tau'ri, Goa'uld, Jaffa, Lucian, Asgard)
- Same color schemes and lore
- Same universe — playing as Tau'ri in both feels connected

### How the Hook Works

#### 1. Stats API (GateWars side)

GateWars runs an HTTP server alongside the SSH server, exposing read-only JSON:

```
GET http://sgc.games:8080/api/status
{
  "season": 14,
  "day": 5,
  "phase": "war",
  "factions": {
    "tauri":  {"territory": 34.2, "players": 12, "ships": 847},
    "goauld": {"territory": 19.8, "players": 8,  "ships": 621},
    "jaffa":  {"territory": 24.1, "players": 5,  "ships": 512},
    "lucian": {"territory": 13.7, "players": 7,  "ships": 398},
    "asgard": {"territory": 8.2,  "players": 3,  "ships": 127}
  }
}

GET http://sgc.games:8080/api/player/{ssh_key_fingerprint}
{
  "faction": "tauri",
  "seasons_played": 12,
  "total_hours": 847,
  "rallies": 23,
  "loyalty_streak": 8
}
```

File: `internal/api/stats_api.go` — simple `net/http` handler, reads from engine snapshot + SQLite.

#### 2. Stats Tab (Stargwent side)

Stargwent's Stats menu gets a **"GateWars" tab** (7th tab) that:
- Fetches `GET /api/status` on open
- Shows current season, phase, faction territories with colored bars
- Shows your personal stats if you have an SSH key configured
- Has a "DIAL IN" button

File to modify: `stargwent/stats_screen.py` (add tab) + new `stargwent/gatewars_stats.py`

#### 3. The Hidden Chevron (Stargwent side)

Inside Stargwent's **Rule Compendium**, the bottom-center chevron is special:
- Pulses with a faint amber glow (like an active Stargate chevron)
- Hover tooltip: *"SGC Tactical Network — Click to dial in. ssh sgc.games"*
- Clicking triggers:
  1. All 7 chevrons light up in sequence (animation)
  2. Brief "kawoosh" screen flash
  3. System terminal opens with SSH command

File to modify: `stargwent/compendium_screen.py` (add chevron interaction)

#### 4. Launching SSH (Stargwent side)

The "DIAL IN" action opens a system terminal with the SSH command:

```python
import subprocess
import sys
import shutil

def dial_in(host="sgc.games", port=22):
    """Launch SSH GateWars in a system terminal."""
    ssh_cmd = f"ssh -p {port} {host}"

    if sys.platform == "linux":
        # Try common terminal emulators
        for term in ["gnome-terminal", "konsole", "xfce4-terminal", "xterm"]:
            if shutil.which(term):
                if term == "gnome-terminal":
                    subprocess.Popen([term, "--", "bash", "-c", ssh_cmd])
                else:
                    subprocess.Popen([term, "-e", ssh_cmd])
                return
    elif sys.platform == "darwin":
        # macOS
        subprocess.Popen(["open", "-a", "Terminal", ssh_cmd])
    elif sys.platform == "win32":
        # Windows Terminal
        subprocess.Popen(["wt", "ssh", f"-p {port}", host])

    # Fallback: just print the command
    print(f"Run this in a terminal: {ssh_cmd}")
```

#### 5. Web Build (Emscripten)

On web builds, the chevron either:
- **Doesn't appear** (cleanest — can't SSH from a browser)
- **Opens a link** to a web terminal proxy (stretch goal — xterm.js + WebSocket bridge to SSH)

Guard with `if sys.platform != "emscripten":` in Stargwent.

### Implementation Order

1. **First**: Build the HTTP stats API in GateWars (`internal/api/stats_api.go`)
2. **Second**: Add the GateWars stats tab in Stargwent (read-only, just fetches JSON)
3. **Third**: Add the "DIAL IN" button that launches SSH
4. **Fourth**: Add the hidden chevron in the Rule Compendium with animation
5. **Future**: Cosmetic rewards (card backs for winning faction players)

---

## Security

### SSH Layer
- **No shell access** — Bubbletea model only, no exec/PTY/forwarding
- Accept all SSH keys for identity, no password auth
- Host key auto-generated, stored outside repo

### Rate Limiting
- Max 500 concurrent sessions total
- Max 10 sessions per SSH key
- 10 new connections/sec globally
- 30 min idle timeout (view-only sessions exempt)
- 2000 max ships in simulation

### Input Validation
- Whitelist-only key handling (q, space, 1-5, tab, ?)
- SSH commands matched against fixed registry (no shell exec)

### Data Protection
- Parameterized SQL queries only (never string interpolation)
- SHA256 fingerprints only (no raw keys, no PII)
- SQLite with WAL mode, 0600 file permissions

---

## Roadmap

### Done

- [x] Phase 0: SSH server, starfield, moving ships, terminal resize
- [x] Phase 0+1: Simulation engine, combat, factions, spawning, territory, HUD

### In Progress

- [ ] Phase 2: Faction powers, sector focus, tab views, multiplex sessions

### Planned

- [ ] Phase 1.5: SQLite persistence (player identity across sessions)
- [ ] Phase 2.5: Planet system (SSH command routing, bonuses, capacity)
- [ ] Phase 3: Seasons & scoring (7-day cycles, leaderboards)
- [ ] Phase 4: Stargwent bridge (chevron UI, stats API)
- [ ] Phase 5: Polish (parallax starfield, beam effects, wormhole animation)

---

## Open Questions

- **Faction switching**: Between seasons? Mid-season? Never? Loyalty should matter but trapping in a losing faction isn't fun.
- **Chat**: Minimal faction-only chat? Predefined emotes? Or zero communication (presence-only)?
- **Mobile**: Termux (Android) and Blink (iOS) support SSH. Test on small screens.
- **Bot policy**: Auto-connect bots — fun programming challenge or unfair?
- **Domain**: `sgc.games` — Stargate Command, clean and thematic.
