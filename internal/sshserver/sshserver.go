package sshserver

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"

	"github.com/Froggys974/e-commerce_golang/internal/apiclient"
	"github.com/Froggys974/e-commerce_golang/internal/tui/admin"
	"github.com/Froggys974/e-commerce_golang/internal/tui/client"
)

const (
	clientAddr  = ":2222"
	adminAddr   = ":2223"
	hostKeyPath = ".ssh/id_ed25519"
)

func Run() error {
	apiURL := apiclient.BaseURLFromEnv()

	clientServer, err := newServer(clientAddr, func() tea.Model { return client.NewModel(apiURL) })
	if err != nil {
		return err
	}
	adminServer, err := newServer(adminAddr, func() tea.Model { return admin.NewModel(apiURL) })
	if err != nil {
		return err
	}

	// les deux serveurs tournent en parallele et le premier qui plante arrete tout
	errs := make(chan error, 2)
	go func() { errs <- clientServer.ListenAndServe() }()
	go func() { errs <- adminServer.ListenAndServe() }()

	log.Printf("SSH client on %s, admin on %s", clientAddr, adminAddr)
	return <-errs
}

// un nouveau modele par session donc un client api et des tokens par utilisateur
func newServer(addr string, newModel func() tea.Model) (*ssh.Server, error) {
	server, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			bubbletea.Middleware(func(sess ssh.Session) (tea.Model, []tea.ProgramOption) {
				return newModel(), nil
			}),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("création du serveur ssh %s : %w", addr, err)
	}
	return server, nil
}
