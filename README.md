# GateWars v3 — Ancient Defense Network

A multiplayer cooperative tower defense game played over SSH. Themed around the Ancient Control Chair from Stargate Atlantis.

## The Game

The galaxy is being invaded by replicators. You are an Ancient chair operator on Atlantis.

- **Deploy your control chair** to a planet under siege
- **Your drones fire automatically** at incoming replicator swarms
- **Choose a faction** — Ancient (drone swarms, shields, Ascension Pulse) or Ori (devastating firepower, Prior's Wrath)
- **Set drone tactics** — Spread, Focus, or Perimeter targeting during defense
- **Hold the planet** for 5 minutes per defender to liberate it
- **Cooperate** — multiple players can deploy chairs on the same planet
- **Upgrade** your chair and drones at the Ancient Throne with ZPM earned from kills
- **Explore the galaxy** — Astrologic star map, Stargate network tube map, or sensor list
- **Upgrade the gate network** — spend ZPM to upgrade Stargate links for defense bonuses
- **Send resources** — shield boosts, drone boosts, or ZPM gifts to planets being defended
- **Galaxy events** — replicator surges (2x spawns, 2x ZPM), planet bounties, liberation milestones
- **New Game+** — liberate the entire galaxy to trigger a harder cycle

When you disconnect, your planet progress resets — but your upgrades persist (roguelite loop).

## Play

```bash
ssh -p 2222 yourserver.com
```

No client install required. Any terminal with SSH.

## Run a Server

```bash
# Build
make build

# Run (random galaxy seed)
make run

# Dev mode (fixed seed, 20 planets)
make dev
```

### Flags

```
--port 2222        SSH listen port
--host 0.0.0.0     Listen address
--key path         SSH host key path
--db path          SQLite database path
--seed N           Galaxy seed (0=random)
--planets N        Number of planets (default 50)
--max-sessions N   Max concurrent players (default 200)
--max-per-key N    Max sessions per SSH key (default 3)
```

## Controls

### Atlantis (Hub)
- `t` — Open upgrade throne (spend ZPM on chair levels & drone tiers)
- `g` — Browse galaxy sensor list
- `a` — Astrologic star map (2D galaxy projection)
- `n` — Stargate network tube map (upgrades & transfers)
- `c` — Open chat
- `q` — Disconnect

### Throne (Upgrades)
- `↑/↓` — Navigate upgrade options (chair level, drone tiers, faction switch, reset)
- `Enter` — Purchase upgrade or switch faction
- `q` — Back to Atlantis

### Galaxy Browser
- `↑/↓` — Navigate planet list
- `Enter` — Deploy to selected planet
- `a` / `n` — Switch to Astro or Network view
- `q` — Back to Atlantis

### Astrologic View (Star Map)
- `←→↑↓` — Pan camera
- `+/-` — Zoom in/out
- `Tab` — Cycle planet selection
- `Enter` — Deploy to selected planet
- `g` / `n` — Switch to Galaxy or Network view
- `q` — Back to Atlantis

### Stargate Network (Tube Map)
- `↑/↓` — Navigate stations
- `Enter` — Deploy to selected planet
- `u` — Upgrade gate link (spend ZPM for defense bonuses)
- `s` — Send resources to planet (shield/drones/ZPM)
- `g` / `a` — Switch to Galaxy or Astro view
- `q` — Back to Atlantis

### Defense (Gameplay)
- `1` — Spread tactic (target nearest)
- `2` — Focus tactic (target strongest)
- `3` — Perimeter tactic (target closest to center)
- `Tab` — Toggle chat panel
- `c` — Open chat input
- `q` — Retreat to Atlantis

### Chat
- Type message + `Enter` — Send
- `Esc` — Close chat input
- `/help` — List all commands
- `/roster` — Online players
- `/team create <name>` — Form a team
- `@callsign msg` — Direct message
- `/indeed` — Indeed.

## Architecture

- **Go 1.22+** with the Charm stack (Wish SSH + Bubbletea TUI + Lipgloss)
- **SQLite** persistence (modernc.org/sqlite, no CGO)
- **10Hz engine** tick loop with per-planet defense instances
- **Single-writer chat hub** with channel-based message routing
- **Radial defense view** — chair at center, concentric defense rings, replicators approach from all directions, drones intercept
- **Astrologic star map** — 2D galaxy projection with constellation lines, planet positions, region labels, pan/zoom
- **Stargate network tube map** — subway-style topology with named routes, gate upgrades, resource transfers
- **SGC terminal aesthetic** — True Color Ancient cyan + ZPM gold palette, animated splash, atmospheric rendering

## Tech Stack

| Component | Library |
|-----------|---------|
| SSH Server | charmbracelet/wish |
| TUI Framework | charmbracelet/bubbletea |
| Styling | charmbracelet/lipgloss |
| Database | modernc.org/sqlite |

## License

CC BY-NC 4.0
