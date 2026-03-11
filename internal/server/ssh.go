package server

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"

	"ssh-gatewars/internal/core"
)

// ModelFactory creates a Bubbletea model for a new session.
type ModelFactory func(session *core.SessionInfo) tea.Model

// RejectModelFactory creates a model for rejected connections.
type RejectModelFactory func(reason string) tea.Model

// SSHServerConfig holds the dependencies for NewSSHServer.
type SSHServerConfig struct {
	Cfg            core.GameConfig
	Limiter        *ConnLimiter
	NewModel       ModelFactory
	NewRejectModel RejectModelFactory
}

// NewSSHServer creates and configures the Wish SSH server.
func NewSSHServer(scfg SSHServerConfig) (*ssh.Server, error) {
	handler := makeHandler(scfg)

	srv, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", scfg.Cfg.Host, scfg.Cfg.Port)),
		wish.WithHostKeyPath(scfg.Cfg.KeyPath),
		wish.WithPublicKeyAuth(func(_ ssh.Context, _ ssh.PublicKey) bool {
			return true
		}),
		wish.WithPasswordAuth(func(_ ssh.Context, _ string) bool {
			return true
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(handler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func makeHandler(scfg SSHServerConfig) bubbletea.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		sshKey := SessionKey(s)

		if !scfg.Limiter.TryConnect(sshKey) {
			log.Warn("connection rejected", "key", sshKey[:20], "reason", "rate limit")
			return scfg.NewRejectModel("Server is busy. Try again shortly."), nil
		}

		pty, _, _ := s.Pty()
		width := pty.Window.Width
		height := pty.Window.Height
		if width == 0 {
			width = 80
		}
		if height == 0 {
			height = 24
		}

		displayName := DisplayName(s)
		info := core.NewSessionInfo(sshKey, displayName, width, height)

		model := scfg.NewModel(info)
		return model, []tea.ProgramOption{tea.WithAltScreen()}
	}
}

// ListenAndServe starts the SSH server and blocks until context is cancelled.
func ListenAndServe(ctx context.Context, srv *ssh.Server) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("SSH server listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return srv.Close()
	case err := <-errCh:
		return err
	}
}
