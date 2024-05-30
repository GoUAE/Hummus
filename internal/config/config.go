package config

import (
	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v11"
)

type hummusConfig struct {
	DiscordBotToken     string `env:"DISCORD_BOT_TOKEN"`
	WhatsappGoUAEJID    string `env:"WA_GOUAE_JID"`
	DiscordWebhookID    string `env:"DISCORD_WEBHOOK_ID"`
	DiscordWebhoolToken string `env:"DISCORD_WEBHOOK_TOKEN"`
	// DiscordChannelID    string `env:"DISCORD_CHANNEL_ID"`
}

// shared session
var (
	HummusConfig hummusConfig
	DG           *discordgo.Session
)

func LoadEnvIntoConfig() error {
	return env.Parse(&HummusConfig)
}
