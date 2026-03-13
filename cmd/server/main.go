package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"ssh-gatewars/internal/chat"
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/engine"
	"ssh-gatewars/internal/server"
	"ssh-gatewars/internal/store"
	"ssh-gatewars/internal/tui"
)

func main() {
	cfg := core.DefaultConfig()

	flag.IntVar(&cfg.Port, "port", cfg.Port, "SSH listen port")
	flag.StringVar(&cfg.Host, "host", cfg.Host, "SSH listen host")
	flag.StringVar(&cfg.KeyPath, "key", cfg.KeyPath, "SSH host key path")
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "SQLite database path")
	flag.IntVar(&cfg.MaxSessions, "max-sessions", cfg.MaxSessions, "max concurrent SSH sessions")
	flag.IntVar(&cfg.MaxPerKey, "max-per-key", cfg.MaxPerKey, "max sessions per SSH key")
	flag.Int64Var(&cfg.Seed, "seed", cfg.Seed, "galaxy seed (0=random)")
	flag.IntVar(&cfg.NumPlanets, "planets", cfg.NumPlanets, "number of planets in galaxy")
	flag.Parse()

	log.Info("GateWars v3.0 starting",
		"port", cfg.Port,
		"seed", cfg.Seed,
		"planets", cfg.NumPlanets,
		"max_players", cfg.MaxSessions,
	)

	// Open player store.
	playerStore, err := store.NewPlayerStore(cfg.DBPath)
	if err != nil {
		log.Fatal("failed to open database", "err", err)
	}
	defer playerStore.Close()

	// Create game engine.
	eng := engine.NewEngine(cfg)

	// Create chat hub.
	chatHub := chat.NewHub(playerStore)

	// Wire engine game events → chat hub (proper type mapping).
	go func() {
		for ev := range eng.GameEvents {
			var chatType chat.GameEventType
			switch ev.Type {
			case engine.GamePlayerDeploy:
				chatType = chat.GamePlayerDeploy
			case engine.GamePlayerRetreat:
				chatType = chat.GamePlayerRetreat
			case engine.GamePlanetLiberated:
				chatType = chat.GamePlanetLiberated
			case engine.GamePlanetFailed:
				chatType = chat.GamePlanetFailed
			case engine.GamePlayerConnect:
				chatType = chat.GamePlayerConnect
			case engine.GamePlayerDisconnect:
				chatType = chat.GamePlayerDisconnect
			case engine.GameSurgeStart:
				chatType = chat.GameSurgeStart
			case engine.GameSurgeEnd:
				chatType = chat.GameSurgeEnd
			case engine.GameMilestone:
				chatType = chat.GameMilestone
			case engine.GameGalaxyReset:
				chatType = chat.GameGalaxyReset
			default:
				continue
			}
			chatHub.GameEvents <- chat.GameEvent{
				Type:       chatType,
				Callsign:   ev.Callsign,
				PlanetName: ev.PlanetName,
				Extra:      ev.Extra,
			}
		}
	}()

	// Create connection limiter.
	limiter := server.NewConnLimiter(cfg.MaxSessions, cfg.MaxPerKey, cfg.ConnectRate)

	// Create SSH server.
	sshSrv, err := server.NewSSHServer(server.SSHServerConfig{
		Cfg:     cfg,
		Limiter: limiter,
		NewModel: func(session *core.SessionInfo) tea.Model {
			return tui.NewModel(tui.ModelConfig{
				Engine:  eng,
				Store:   playerStore,
				ChatHub: chatHub,
				Session: session,
				Width:   session.TermWidth,
				Height:  session.TermHeight,
			})
		},
		NewRejectModel: func(reason string) tea.Model {
			return tui.NewRejectModel(reason)
		},
	})
	if err != nil {
		log.Fatal("failed to create SSH server", "err", err)
	}

	// Start engine.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Run(ctx)
	log.Info("engine started", "tick_rate", cfg.TickRate, "planets", cfg.NumPlanets)

	// Start chat hub.
	go chatHub.Run(ctx.Done())
	log.Info("chat hub started")

	// Handle shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nshutting down...")
		cancel()
		sshSrv.Close()
	}()

	// Start SSH server (blocks).
	log.Info("SSH server listening", "addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err := sshSrv.ListenAndServe(); err != nil {
		select {
		case <-ctx.Done():
			// Expected shutdown.
		default:
			log.Fatal("SSH server error", "err", err)
		}
	}

	log.Info("GateWars stopped")
}
