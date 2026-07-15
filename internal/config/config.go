package config

import "time"

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
