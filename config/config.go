package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
)

type AppConfig struct {
	App struct {
		Name      string
		Port      string
	}
}

var Config AppConfig

func InitConfig(DevMode bool) *AppConfig {
	if DevMode {
		if err := godotenv.Load(); err != nil {
			log.Error().Err(err).Msg("Error loading .env file")
		}
	}

	Config.App.Name = os.Getenv("APP_NAME")
	Config.App.Port = os.Getenv("PORT")

	return &Config
}