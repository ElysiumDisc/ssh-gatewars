# SSH GateWars

> A persistent space war between five factions, fought across the Stargate Network.
> Always running. Always beautiful. Your presence fuels the fight.
> No downloads. No accounts. Just `ssh sgc.games`

```
$ ssh sgc.games

  ╔═══════════════════════════════════════════════╗
  ║                                               ║
  ║   SGC TACTICAL NETWORK                        ║
  ║   Developed by Lt. Col. Carter & SGC Staff    ║
  ║                                               ║
  ║   "I was just modeling hyperspace fleet        ║
  ║    trajectories. The game part was an          ║
  ║    accident." — Dr. Samantha Carter            ║
  ║                                               ║
  ║   The war is already in progress.              ║
  ║   Choose your side.                            ║
  ║                                               ║
  ╚═══════════════════════════════════════════════╝
```

## What Is This?

SSH GateWars is an **ambient persistent faction war** rendered in the terminal over SSH. Five factions from the Stargate universe clash in an eternal space battle. Ships spawn, fight, and die — endlessly, beautifully, whether anyone is watching or not.

**It is not a game you play. It is a war you join.**

- **Just watching?** Your faction spawns more ships. You are fuel.
- **Interacting?** You trigger your faction's unique power. You are fire.
- **Many players online?** Your faction surges. Numbers win wars.

## The Five Fleets

| Faction | Color | Ships | Style |
|---------|-------|-------|-------|
| **Tau'ri** | Steel blue | F-302 interceptors `->` | Tight formation clusters, focused fire |
| **Goa'uld** | Gold amber | Ha'tak motherships `{=>` | Slow advancing wall, shield matrix |
| **Jaffa Rebellion** | Bright yellow | Al'kesh bombers `>>` | Accelerating charges, battle fury |
| **Lucian Alliance** | Purple | Cargo runners `~>` | Erratic weaving, unpredictable spawns |
| **Asgard** | Cyan | O'Neill warships `*->` | Few but powerful, ion cannon beams |

## Controls

| Key | Action |
|-----|--------|
| `Space` | Activate faction power (shared cooldown) |
| `1-5` | Focus spawns toward a sector |
| `Tab` | Cycle views (battlefield / scoreboard / stats) |
| `?` | Help overlay |
| `q` | Disconnect |

## Multiplex Views

Open multiple terminals for different perspectives of the war:

```
ssh sgc.games              # Battlefield (default, full controls)
ssh sgc.games scoreboard   # Live faction scoreboard
ssh sgc.games network      # Stargate network territory map
ssh sgc.games stats        # Your personal stats + season history
```

Same SSH key = same player. Multiple sessions don't give extra spawn bonuses.

## Running the Server

```bash
# Build
go build -o gatewars ./cmd/server/

# Run (generates SSH host key on first start)
./gatewars --port 2222

# Connect
ssh -p 2222 localhost
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `2222` | SSH server port |
| `--host` | `0.0.0.0` | Bind address |
| `--key` | `.ssh/id_ed25519` | Host key path |

## Tech Stack

- **Go** — goroutines for concurrent players, single binary
- **[Wish](https://github.com/charmbracelet/wish)** — SSH app framework
- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** — TUI with Elm architecture
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — terminal styling
- **SQLite** — player persistence, season history

## Part of the Stargwent Universe

SSH GateWars is a companion to [Stargwent](https://github.com/mrgamer/stargwent), a Stargate-themed card game. Same factions, same lore, same universe.

## License

TBD
