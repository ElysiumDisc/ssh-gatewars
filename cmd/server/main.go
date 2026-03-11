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
	"ssh-gatewars/internal/server"
	"ssh-gatewars/internal/simulation"
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
	flag.Float64Var(&cfg.ConnectRate, "connect-rate", cfg.ConnectRate, "max new connections/sec")
	flag.Int64Var(&cfg.Seed, "seed", cfg.Seed, "world seed (0=random)")
	flag.Parse()

	log.Info("GateWars v2.0 starting",
		"port", cfg.Port,
		"seed", cfg.Seed,
		"max_players", cfg.MaxSessions,
	)

	// Open player store.
	playerStore, err := store.NewPlayerStore(cfg.DBPath)
	if err != nil {
		log.Fatal("failed to open database", "err", err)
	}
	defer playerStore.Close()

	// Create simulation engine.
	engine := simulation.NewEngine(cfg)

	// Create chat hub.
	chatHub := chat.NewHub(playerStore)

	// Wire engine game events → chat hub.
	engine.GameEvents = chatHub.GameEvents

	// Create connection limiter.
	limiter := server.NewConnLimiter(cfg.MaxSessions, cfg.MaxPerKey, cfg.ConnectRate)

	// Create SSH server with model factories.
	sshSrv, err := server.NewSSHServer(server.SSHServerConfig{
		Cfg:     cfg,
		Limiter: limiter,
		NewModel: func(session *core.SessionInfo) tea.Model {
			return tui.NewModel(tui.ModelConfig{
				Engine:  engine,
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

	go engine.Run(ctx)
	log.Info("simulation engine started", "tick_rate", cfg.TickRate)

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
