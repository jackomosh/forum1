package app

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"forum/internal/config"
	"forum/internal/handler"
	"forum/internal/repository"
	"forum/internal/repository/sqlite"
)

type Application struct {
	Config     config.Config
	Repository repository.Repository
	server     *http.Server
}

type Dependencies struct {
	Config config.Config
}

func Run() error {
	application, err := New(Dependencies{Config: config.Default()})
	if err != nil {
		return err
	}
	defer application.Repository.Close()

	log.Printf("server running at http://localhost:%s", application.Config.Server.Port)
	return application.ListenAndServe()
}

func New(deps Dependencies) (*Application, error) {
	cfg := deps.Config
	if cfg.Server.Port == "" {
		cfg = config.Default()
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "forum.db"
	}

	client, err := sqlite.NewClient(cfg.Database.Path)
	if err != nil {
		return nil, err
	}

	if err := applySchema(client); err != nil {
		_ = client.Close()
		return nil, err
	}

	repo := sqlite.NewRepository(client)
	renderer := handler.NewRenderer("web/templates")
	forumHandler := handler.NewForumHandler(repo, renderer, handler.Options{
		SessionCookieName: cfg.Session.CookieName,
		SessionDuration:   cfg.Session.Duration,
		SessionSecure:     cfg.Session.Secure,
		SessionSameSite:   cfg.Session.SameSite,
		PasswordMinLength: cfg.Security.PasswordMinLength,
	})

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	forumHandler.RegisterRoutes(mux)

	return &Application{
		Config:     cfg,
		Repository: repo,
		server: &http.Server{
			Addr:         serverAddress(cfg),
			Handler:      mux,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}, nil
}

func (a *Application) ListenAndServe() error {
	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func applySchema(client *sqlite.Client) error {
	schemaPath := "schema.sql"

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema %q: %w", schemaPath, err)
	}

	if _, err := client.DB().Exec(string(schema)); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

func serverAddress(cfg config.Config) string {
	if cfg.Server.Host == "" {
		return ":" + cfg.Server.Port
	}
	return net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)
}
