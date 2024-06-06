package discord

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	dg "github.com/bwmarrin/discordgo"
	"go.mau.fi/whatsmeow/types/events"
)

type Bot struct {
	session *dg.Session
	token   string

	Log *log.Logger
}

func New(token string) Bot {
	bot := Bot{token: token}

	bot.Log = log.New(os.Stdout, "[Discord] ", log.LstdFlags)

	bot.session, _ = dg.New("Bot " + bot.token)
	bot.session.Identify.Intents = dg.IntentsGuildMessages | dg.IntentsMessageContent | dg.IntentsGuildWebhooks

	return bot
}

func (bot *Bot) Channel(channelID string) (chan *events.Message, error) {
	var webHook *dg.Webhook

	existingHooks, err := bot.session.ChannelWebhooks(channelID)
	if err != nil {
		return nil, err
	}

	// Check if we have an existing whatsapp bridge hook
	for _, hook := range existingHooks {
		if hook.Name == "Hummus Whatsapp Bridge" {
			webHook = hook
		}
	}

	// Otherwise create a new webhook
	if webHook == nil {
		webHook, err = bot.session.WebhookCreate(channelID, "Hummus Whatsapp Bridge", "")
		if err != nil {
			return nil, err
		}
	}

	ch := make(chan *events.Message, 1)
	go bot.Pipe(webHook, ch)

	return ch, nil
}

// TODO: Refactor PipeToDiscord to use channels & solve existing edge-cases
func (bot *Bot) Pipe(webHook *dg.Webhook, ch chan *events.Message) {
	for v := range ch {
		_, err := bot.session.WebhookExecute(
			webHook.ID,
			webHook.Token,
			true,
			&discordgo.WebhookParams{ /* TODO */ },
		)

		if err != nil {
			bot.Log.Println("Failed to send message to discord, ", err)
		}
	}
}

func (bot *Bot) Run() error {
	return bot.session.Open()
}

func (bot *Bot) Stop() error {
	return bot.session.Close()
}
