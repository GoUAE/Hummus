package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gouae/hummus/internal/config"
	"github.com/gouae/hummus/internal/discord"
	"github.com/gouae/hummus/internal/whatsapp"
)

func main() {
	cfg, err := config.LoadEnv()
	if err != nil {
		log.Fatal("Failed to load config from environment, ", err)
	}

	dc := discord.New(cfg.DiscordToken)

	wa, err := whatsapp.New()
	if err != nil {
		wa.Log.Fatal("Failed to initialize the whatsapp bot, ", err)
	}

	// Signal handling to gracefully shutdown the bot
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	if err = dc.Run(); err != nil {
		dc.Log.Fatal("Failed to start the discord bot, ", err)
	}

	if err = wa.Run(); err != nil {
		dc.Stop()
		wa.Log.Fatal("Failed to start the whatsapp bot, ", err)
	}

	ch, err := dc.Channel("1248051374969851954")

	if err != nil {
		BatchStop(&wa, &dc)
		dc.Log.Fatal("Failed to get the discord channel, ", err)
	}

	if err = wa.Bridge("GoUAE Community", ch); err != nil {
		BatchStop(&wa, &dc)
		wa.Log.Fatal("Failed to bridge the whatsapp bot, ", err)
	}

	<-stop // Wait for a termination signal

	log.Println("Shutting down the bot...")

	BatchStop(&wa, &dc)

	log.Println("Bot stopped gracefully.")
}

func BatchStop(wa *whatsapp.Bot, dc *discord.Bot) {
	wa.Stop()
	dc.Stop()
}
