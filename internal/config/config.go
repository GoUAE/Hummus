package config

import (
	"github.com/caarlos0/env/v11"
)

type HummusConfig struct {
	DiscordBotToken     string `env:"DISCORD_BOT_TOKEN"`
	WhatsappGoUAEJID    string `env:"WA_GOUAE_JID"`
	DiscordWebhookID    string `env:"DISCORD_WEBHOOK_ID"`
	DiscordWebhoolToken string `env:"DISCORD_WEBHOOK_TOKEN"`
	FallbackAvatarURL   string `env:"DISCORD_FALLBACK_AVATAR_URL"`
	// DiscordChannelID    string `env:"DISCORD_CHANNEL_ID"`
}

func LoadFromEnv() (HummusConfig, error) {
	var hummusConfig HummusConfig
	err := env.Parse(&hummusConfig)

	return hummusConfig, err
}
