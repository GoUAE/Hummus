package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gouae/hummus/internal/config"
	"github.com/gouae/hummus/internal/discord"
	"github.com/gouae/hummus/internal/utils"
	"github.com/gouae/hummus/internal/whatsapp"

	// autoload loads all the environment variables from the .env file
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		utils.LogError(err, "Failed to load config from environment")
		return
	}

	discordBot, err := discord.New(cfg)
	if err != nil {
		utils.LogError(err, "Failed to start the discord bot")
		return
	}

	// Signal handling to gracefully shutdown the bot
	stop := make(chan os.Signal, 1)

	// channel to listen for SIGINT and SIGTERM events (CTRL-C etc.)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	err = discordBot.Run()
	if err != nil {
		return
	}

	waBot, err := whatsapp.New(cfg, discordBot)
	if err != nil {
		utils.LogError(err, "Failed to initialize the whatsapp bot")
		utils.LogError(discordBot.Stop(), "Error stopping discord bot")
		return
	}

	err = waBot.Run()
	if err != nil {
		utils.LogError(err, "Failed to start the whatsapp bridge")
		utils.LogError(discordBot.Stop(), "Error stopping discord bot")
		return
	}

	<-stop // Wait for a termination signal

	log.Println("\nShutting down the bot...")

	waBot.Stop()
	err = discordBot.Stop()
	if err != nil {
		utils.LogError(discordBot.Stop(), "Error stopping discord bot")
		return
	}

	log.Println("Bot stopped gracefully.")
}
