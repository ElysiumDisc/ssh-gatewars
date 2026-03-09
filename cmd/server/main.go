package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
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

	"ssh-gatewars/internal/httpapi"
	"ssh-gatewars/internal/player"
	"ssh-gatewars/internal/server"
	"ssh-gatewars/internal/simulation"
)

func main() {
	port := flag.Int("port", 2222, "SSH server port")
	host := flag.String("host", "0.0.0.0", "SSH server host")
	keyPath := flag.String("key", ".ssh/id_ed25519", "SSH host key path")
	dbPath := flag.String("db", "gatewars.db", "SQLite database path")
	httpAddr := flag.String("http", "127.0.0.1:8080", "HTTP stats API address (empty to disable)")
	maxSessions := flag.Int("max-sessions", 500, "Maximum concurrent SSH sessions")
	maxPerKey := flag.Int("max-per-key", 10, "Maximum sessions per SSH key")
	connectRate := flag.Float64("connect-rate", 10, "Max new connections per second")
	idleTimeout := flag.Duration("idle-timeout", 30*time.Minute, "Idle session timeout")
	seed := flag.Int64("seed", 0, "Galaxy seed (0 = random)")
	systemCount := flag.Int("systems", 50, "Number of star systems")
	flag.Parse()

	store, err := player.NewStore(*dbPath)
	if err != nil {
		log.Fatal("Failed to open database", "error", err)
	}
	defer store.Close()

	limiter := server.NewConnLimiter(*maxSessions, *maxPerKey, *connectRate)

	engine := simulation.NewEngine(*seed, *systemCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go engine.Run(ctx)

	srv, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", *host, *port)),
		wish.WithHostKeyPath(*keyPath),
		wish.WithPublicKeyAuth(func(_ ssh.Context, _ ssh.PublicKey) bool {
			return true
		}),
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(
				makeHandler(engine, store, limiter, *idleTimeout),
				termenv.ANSI256,
			),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("Failed to create server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	log.Info("Starting SSH GateWars — Galactic Conquest",
		"host", *host, "port", *port, "systems", engine.GalaxySystemCount())
	log.Info("Connect with:", "command", fmt.Sprintf("ssh -p %d localhost", *port))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			select {
			case <-done:
			default:
				log.Error("Server error", "error", err)
			}
		}
	}()

	if *httpAddr != "" {
		statsAPI := httpapi.NewStatsServer(engine)
		httpSrv := &http.Server{Addr: *httpAddr, Handler: statsAPI.Handler()}
		go func() {
			log.Info("Starting HTTP stats API", "addr", *httpAddr)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error("HTTP server error", "error", err)
			}
		}()
	}

	<-done
	log.Info("Shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Shutdown error", "error", err)
	}
}

func makeHandler(engine *simulation.Engine, store *player.Store, limiter *server.ConnLimiter, idleTimeout time.Duration) bubbletea.ProgramHandler {
	return func(s ssh.Session) *tea.Program {
		sshKey := server.SessionKey(s)

		if !limiter.TryConnect(sshKey) {
			fmt.Fprintln(s, "Server is full or rate limited. Try again shortly.")
			s.Close()
			return nil
		}

		renderer := bubbletea.MakeRenderer(s)
		sshUser := s.User()
		sshCmd := s.Command()

		model := server.NewModel(engine, renderer, sshKey, store, sshUser, sshCmd)

		go func() {
			<-s.Context().Done()
			engine.UnregisterPlayer(sshKey)
			limiter.Disconnect(sshKey)
		}()

		go func() {
			ticker := time.NewTicker(60 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-s.Context().Done():
					return
				case <-ticker.C:
					if !model.IsViewOnly() && time.Since(model.LastActivity()) > idleTimeout {
						log.Info("Idle timeout", "key", sshKey)
						s.Close()
						return
					}
				}
			}
		}()

		opts := bubbletea.MakeOptions(s)
		opts = append(opts, tea.WithAltScreen())
		return tea.NewProgram(model, opts...)
	}
}
