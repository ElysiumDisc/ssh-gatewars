# SSH GateWars

> A 4X strategy game inspired by Master of Orion (1993), themed as Stargate SG-1.
> Played entirely over SSH. No downloads. No accounts. Just `ssh sgc.games`

```
$ ssh sgc.games

        .-'''''-.
      .'           '.
     /    .-'''-.    \
    |   /         \   |
    |  |  GATEWARS |  |
    |   \         /   |
     \    '-...-'    /
      '.           .'
        '-......-'

  SSH GATEWARS — MASTER OF ORION
  A Stargate SG-1 4X Strategy Game
```

## What Is This?

SSH GateWars is a **multiplayer 4X strategy game** rendered over SSH. Players join one of 5 Stargate SG-1 factions and cooperatively manage their faction's galactic empire — colonizing planets, building factories, researching technology, designing ships, and conquering rivals through a procedurally generated galaxy connected by Stargates.

**It is not a game you play alone. Every online player strengthens their faction.**

- **Explore** a procedurally generated galaxy of 50+ star systems connected by Stargates
- **Colonize** planets with varying size, type, and mineral richness
- **Build** factories, missile bases, and manage 5 production sliders per colony
- **Research** across 6 tech trees to unlock new capabilities
- **Conquer** rival factions through fleet combat and diplomacy

### Multiplayer Model

Each player joins one faction. Every online player adds to faction output (production, research). Any faction member can manage any colony, design ships, command fleets. This is cooperative empire management — more players online means more production.

## The Five Factions

| # | Faction | Color | Strength | Weakness | Special |
|---|---------|-------|----------|----------|---------|
| 1 | **Tau'ri** | Steel blue | Diplomacy +30%, Research +10% | Average combat | Cheaper treaties |
| 2 | **Goa'uld** | Red | Production +20%, Attack +20% | Research -15% | Higher factory cap |
| 3 | **Asgard** | Cyan teal | Research +50% | Ground -30%, slow growth | Cheaper miniaturization |
| 4 | **Tok'ra** | Purple | Espionage +50% | Production -20% | Sabotage & intel |
| 5 | **Jaffa Free Nation** | Yellow | Ground +50%, Defense +10% | Research -20% | Cheaper invasions |

## The Galaxy

Each campaign generates a procedural galaxy:

- **50 star systems** (configurable) with Stargate connections
- **5 homeworlds**: Earth, Chulak, Othala, Vorash, Cal Mah — one per faction, Terran/Medium/Normal
- **Dakara**: center of the galaxy, Ancient superweapon, Huge/Ultra-Rich
- **~30 named planets** from SG-1 canon: Abydos, Langara, Tollana, Heliopolis, Edora, Argos...
- **Gate network**: each system connects to ~3 nearest neighbors via Stargates

### Planets

| Attribute | Values |
|-----------|--------|
| **Type** (13) | Terran, Ocean, Jungle, Arid, Steppe, Tundra, Desert (habitable) + Barren, Volcanic, Toxic, Inferno, Dead, Radiated (hostile — requires tech) |
| **Size** (5) | Tiny (pop 2), Small (4), Medium (7), Large (10), Huge (14) |
| **Minerals** (5) | Ultra-Poor (0.33x), Poor (0.5x), Normal (1x), Rich (1.33x), Ultra-Rich (1.67x) |

### Stargate Theming

| MOO Concept | Stargate Theme |
|-------------|----------------|
| Star systems | Star systems connected by Stargates |
| Hyperspace lanes | Stargate network (instant) + Hyperspace (slow, any-to-any) |
| 10 Races | 5 Factions: Tau'ri, Goa'uld, Asgard, Tok'ra, Jaffa |
| BC (money) | Naquadah |
| 4 hull sizes | Al'kesh, Ha'tak, O'Neill-class, City-ship |
| 6 tech trees | SGC Systems, Goa'uld Engineering, Asgard Shields, Ancient Knowledge, Hyperdrive Tech, Weapons |
| Galactic Council | Galactic Council (same) |
| Orion + Guardian | Dakara (Ancient superweapon) + Guardian fleet |

## Colony Management

Each colony has **5 production sliders** (sum to 100%):

| Slider | Effect |
|--------|--------|
| **Ship** | Builds ships from the build queue |
| **Defense** | Builds missile bases (50 production each) |
| **Industry** | Builds factories (10 production each) — more factories = more total output |
| **Ecology** | Cleans waste, then terraforms (increases max population) |
| **Research** | Feeds the faction research pool across 6 tech trees |

### Production Formula

```
baseOutput = min(population, factories) * mineralMultiplier * factionBonus
onlineBonus = 1.0 + min(onlinePlayers * 5%, cap 50%)
totalOutput = baseOutput * onlineBonus
```

Population grows logistically — fast when low, slowing near the planet's max capacity.

## Controls

### Galaxy Map (Hub)

| Key | Action |
|-----|--------|
| Arrow keys / `hjkl` | Navigate between star systems |
| `Enter` | Open system view |
| `t` | Tech tree / research allocation |
| `d` | Diplomacy view |
| `s` | Ship designer |
| `Tab` | Scoreboard |
| `?` | Help |
| `q` | Disconnect |

### System View

| Key | Action |
|-----|--------|
| Arrow keys / `hjkl` | Navigate to adjacent systems |
| `Enter` | Manage colony (if owned) |
| `Esc` | Back to galaxy map |

### Colony Management

| Key | Action |
|-----|--------|
| `Up` / `Down` | Select production slider |
| `Left` / `Right` | Adjust slider (-5% / +5%) |
| `Esc` | Back to system view |

### Tech Tree

| Key | Action |
|-----|--------|
| `Up` / `Down` | Select tech tree |
| `Left` / `Right` | Adjust research allocation |
| `Esc` | Back to galaxy map |

## Running the Server

```bash
# Build
go build -o gatewars ./cmd/server/

# Run (generates SSH host key on first start)
./gatewars --port 2222

# Connect
ssh -p 2222 localhost
```

### Quick Faction Login

```
ssh tauri@sgc.games     # Jump straight into Tau'ri
ssh asgard@sgc.games    # Join the Asgard
ssh tokra@sgc.games     # Join the Tok'ra
```

### Multiplex Views

```
ssh sgc.games              # Galaxy map (default)
ssh sgc.games galaxy       # Galaxy map directly
ssh sgc.games scoreboard   # Faction scoreboard
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `2222` | SSH server port |
| `--host` | `0.0.0.0` | Bind address |
| `--key` | `.ssh/id_ed25519` | Host key path |
| `--db` | `gatewars.db` | SQLite database path |
| `--http` | `127.0.0.1:8080` | HTTP stats API address (empty to disable) |
| `--seed` | `0` | Galaxy seed (0 = random) |
| `--systems` | `50` | Number of star systems (20-100) |
| `--max-sessions` | `500` | Maximum concurrent SSH sessions |
| `--max-per-key` | `10` | Maximum sessions per SSH key |
| `--connect-rate` | `10` | Max new connections per second |
| `--idle-timeout` | `30m` | Idle session timeout |

## Tech Stack

- **Go 1.26** — goroutines for concurrent players, single binary
- **[Wish](https://github.com/charmbracelet/wish)** — SSH app framework
- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** — TUI with Elm architecture
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — terminal styling
- **SQLite** (modernc.org/sqlite) — player persistence, no CGO

## Implementation Layers

The game is being built incrementally:

| Layer | Status | Content |
|-------|--------|---------|
| **1. Galaxy + Colonies + Economy** | Done | Procedural galaxy, colony sliders, population growth, factory building, research |
| **2. Ship Design + Fleets** | Planned | Hull/component system, fleet movement (gate instant / hyperspace slow) |
| **3. Combat** | Planned | Stack-based tactical combat, bombardment, ground invasion |
| **4. Technology** | Planned | 50+ techs, miniaturization, component gating |
| **5. Diplomacy + Victory** | Planned | Treaties, Galactic Council, victory conditions |
| **6. Polish + Specials** | Planned | Dakara guardian, Tok'ra espionage, starbases |

## In-Universe Lore

> Replicator-Colonel-Carter coded this game. It's embedded in SGC terminals
> as part of the Stargwent universe. Players "discover" it while browsing
> classified systems.

## Part of the Stargwent Universe

SSH GateWars is a companion to [Stargwent](https://github.com/mrgamer/stargwent), a Stargate-themed card game. Same factions, same lore, same universe.

## License

CC BY-NC 4.0
