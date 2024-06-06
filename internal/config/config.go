package config

import (
	"github.com/caarlos0/env/v11"

	// Automagically load environment variables from the .env file
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DiscordToken string `env:"DISCORD_BOT_TOKEN"`
}

func LoadEnv() (cfg Config, err error) {
	err = env.Parse(&cfg)
	return
}
