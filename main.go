package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gouae/hummus/internal/config"
	"github.com/gouae/hummus/internal/whatsapp"
	_ "github.com/joho/godotenv/autoload"
)

func RunDiscordBot() {
	var err error
	config.DG, err = discordgo.New("Bot " + config.HummusConfig.DiscordBotToken)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	// sets correct itents
	config.DG.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	err = config.DG.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	config.DG.Close()
}

func main() {
	if err := config.LoadEnvIntoConfig(); err != nil {
		log.Printf("%+v\n", err)
	}

	// runs the discord bot on a goroutine
	go RunDiscordBot()

	// Signal handling to gracefully shutdown the bot
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		whatsapp.RunWhatsappBot()
	}()

	<-stop // Wait for a termination signal

	log.Println("\nShutting down the bot...")

	if config.DG != nil {
		err := config.DG.Close()
		if err != nil {
			log.Println("Error closing Discord session,", err)
		}
	}

	log.Println("Bot stopped gracefully.")
}
