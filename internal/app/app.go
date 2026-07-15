package app

import "forum/internal/config"

type Application struct {
	Config config.Config
}

type Dependencies struct {
	Config config.Config
}
