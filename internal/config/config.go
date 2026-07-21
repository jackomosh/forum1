package config

import (
	"os"
	"time"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Security SecurityConfig
}

type AppConfig struct {
	Name        string
	Environment string
	BaseURL     string
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Driver string
	Path   string
	DSN    string
}

type SessionConfig struct {
	CookieName string
	Duration   time.Duration
	Secure     bool
	HTTPOnly   bool
	SameSite   string
}

type SecurityConfig struct {
	PasswordMinLength int
	CSRFTokenName     string
}

func Default() Config {
	port := getenv("FORUM_PORT", "8089")
	databasePath := getenv("FORUM_DB_PATH", "forum.db")

	return Config{
		App: AppConfig{
			Name:        "Dev Forum",
			Environment: "development",
			BaseURL:     "http://localhost:8089",
		},
		Server: ServerConfig{
			Host:            "",
			Port:            port,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Database: DatabaseConfig{
			Driver: "sqlite3",
			Path:   databasePath,
		},
		Session: SessionConfig{
			CookieName: "forum_session",
			Duration:   7 * 24 * time.Hour,
			Secure:     false,
			HTTPOnly:   true,
			SameSite:   "lax",
		},
		Security: SecurityConfig{
			PasswordMinLength: 8,
			CSRFTokenName:     "csrf_token",
		},
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
