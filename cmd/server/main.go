package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"

	"ssh-gatewars/internal/render"
	"ssh-gatewars/internal/server"
	"ssh-gatewars/internal/simulation"
)

func main() {
	port := flag.Int("port", 2222, "SSH server port")
	host := flag.String("host", "0.0.0.0", "SSH server host")
	keyPath := flag.String("key", ".ssh/id_ed25519", "SSH host key path")
	flag.Parse()

	// Create simulation engine
	engine := simulation.NewEngine()

	// Create shared starfield
	starfield := render.NewStarfield(simulation.WorldW, simulation.WorldH, 42)

	// Start simulation in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go engine.Run(ctx)

	// SSH server
	srv, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", *host, *port)),
		wish.WithHostKeyPath(*keyPath),
		wish.WithPublicKeyAuth(func(_ ssh.Context, _ ssh.PublicKey) bool {
			return true // accept all keys for identity
		}),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(
				makeHandler(engine, starfield),
				termenv.ANSI256,
			),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("Failed to create server", "error", err)
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	log.Info("Starting SSH GateWars", "host", *host, "port", *port)
	log.Info("Connect with:", "command", fmt.Sprintf("ssh -p %d localhost", *port))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Server error", "error", err)
		}
	}()

	<-done
	log.Info("Shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Shutdown error", "error", err)
	}
}

func makeHandler(engine *simulation.Engine, starfield *render.Starfield) bubbletea.ProgramHandler {
	return func(s ssh.Session) *tea.Program {
		renderer := bubbletea.MakeRenderer(s)
		sshKey := server.SessionKey(s)

		model := server.NewModel(engine, starfield, renderer, sshKey)

		// Cleanup on session end
		go func() {
			<-s.Context().Done()
			model.Cleanup()
		}()

		opts := bubbletea.MakeOptions(s)
		opts = append(opts, tea.WithAltScreen())
		return tea.NewProgram(model, opts...)
	}
}
