# SSH GateWars

A persistent multiplayer roguelike set in the Stargate SG-1 universe — playable entirely over SSH.

```
ssh -p 2222 localhost
```

Explore planets through a Stargate network, fight Jaffa with ranged and melee combat, discover Ancient technology, chat with other players, and bring loot back to SGC — all in your terminal.

## Features

- **SSH multiplayer** — No client install. Connect with any terminal emulator.
- **Stargate network** — Dial 7-symbol gate addresses to travel between worlds. Discover new addresses through exploration and intel.
- **Procedural planets** — Each gate address generates a unique world with biome, threat level, enemies, and loot. Persistent named planets (SGC, Abydos, Chulak) coexist with infinite procedural ones.
- **Real-time world** — 10 Hz simulation. Enemies patrol and hunt independently. The world is alive whether you're moving or not.
- **Ranged & melee combat** — `f` to aim and fire ranged weapons (P-90, staff weapons, zat). Bump into enemies for melee. Line of sight, cover system, and projectile physics.
- **Cover system** — Hide behind walls, crates, and consoles to reduce incoming fire. Cover values per tile type.
- **25+ weapons** — Earth (P-90, M9, shotgun, C-4), Goa'uld (staff weapon, zat, kara kesh), Ancient (drones, hand weapon, ARG), Asgard (plasma beam), Ori (staff, stun).
- **20+ enemy types** — Jaffa patrols, Serpent Guards, Kull Warriors, Replicator swarms, Ori soldiers, Unas, crystal entities, Sodan warriors.
- **Star map** — `m` to browse the entire gate network. Astroterm-inspired star field with constellation lines connecting named planets. Stars colored by biome, sized by threat. Pan, zoom, and dial directly from the map.
- **DHD interface** — Circular dial-home device with 39 glyphs in concentric rings. Locked chevrons light up as you dial.
- **Chat system** — devzat-inspired. Global #ops channel, planet-scoped #local, SG team channels, DMs. Walter NPC announces gate events. Slash commands (/help, /tune, /roster, /team, /me, /kree).
- **SG teams** — Create named teams, invite players, team chat channel, coordinated exploration.
- **SGC home base** — Gate room, armory, briefing room, infirmary, mess hall. Gear up between missions.
- **Player persistence** — Character stats, inventory, and discovered addresses saved to SQLite.
- **Connection security** — Token-bucket rate limiting, per-key session caps, graceful degradation under load.

## Quick Start

**Requirements:** Go 1.22+

```bash
make build
./gatewars

# Or build and run in one step
make run
```

**Connect:**
```bash
ssh -p 2222 localhost
```

### Server Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `2222` | SSH listen port |
| `--seed` | `0` | World seed (0 = random) |
| `--db` | `gatewars.db` | SQLite database path |
| `--max-sessions` | `200` | Maximum concurrent connections |

## Controls

| Key | Action |
|-----|--------|
| `WASD` / `HJKL` / Arrows | Move (bump into enemies for melee) |
| `f` | Aim mode (ranged fire) |
| `r` | Reload weapon |
| `e` | Interact (loot, NPCs, consoles) |
| `g` | Dial stargate (when adjacent) |
| `i` | Inventory |
| `m` | Star map (gate network browser) |
| `A` | Address book |
| `c` | Chat (cycle: hidden/compact/expanded) |
| `Tab` | Player list |
| `?` | Help |
| `q` / `Esc` | Quit / back |

### Chat Commands

| Command | Description |
|---------|-------------|
| `/help` | Show command list |
| `/tune #channel` | Switch active channel |
| `/roster` | List online players |
| `/who` | Players on current channel |
| `/callsign <name>` | Change display name |
| `/me <action>` | Emote |
| `@name message` | Direct message |
| `/team create <n>` | Create SG team |
| `/team invite <n>` | Invite to team |
| `/kree` | KREE! |
| `/indeed` | Indeed. |

## Architecture

```
cmd/server/         Entry point, flag parsing, server bootstrap
internal/
  core/             Configuration, geometry types, session metadata
  simulation/       Game engine (10Hz tick loop, planet instances, enemy AI, projectiles)
  gamedata/         Tiles, biomes, items, enemies, gate addresses, factions
  world/            Tile maps, procedural generation, SGC layout, stargate, LOS
  entity/           Characters, enemies, inventory
  combat/           Melee damage, ranged attacks, projectiles, cover system, LOS
  chat/             Message hub, channels, Walter NPC, slash commands, teams
  server/           SSH server, identity, rate limiter
  store/            SQLite persistence (characters, inventory, addresses, chat, teams)
  tui/              Bubbletea TUI model, views, keybinds
```

**Concurrency model:** Single-writer engine goroutine ticks all active planet instances at 10Hz. Each planet publishes atomic immutable snapshots. Reader sessions access snapshots lock-free. Chat hub runs as a separate goroutine with channel-based message routing.

## Tech Stack

- [Go](https://go.dev/) 1.22+
- [Wish](https://github.com/charmbracelet/wish) — SSH server
- [Bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — Pure-Go SQLite (no CGO)

## Roadmap

- [x] Phase 1 — SGC hub, gate dialing, procedural planets, bump combat, loot, persistence
- [x] Phase 2 — Chat system, SG teams, Walter NPC, multiplayer visibility
- [x] Phase 3 — Ranged combat, aim mode, cover system, projectiles, 25+ weapons, 20+ enemies
- [ ] Phase 4 — Tech tree, faction reputation, crafting, SGC lab
- [ ] Phase 5 — Named planets (Abydos, Chulak, Dakara...), missions, server-wide events
- [ ] Phase 6 — Specializations, perks, ascension, ASCII art polish

## License

Creative Commons Attribution-NonCommercial 4.0 International (CC BY-NC 4.0)
