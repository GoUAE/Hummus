package whatsapp

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/gouae/hummus/internal/config"
	"github.com/gouae/hummus/internal/discord"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Bot struct {
	session          *whatsmeow.Client
	whatsappGoUAEJID string
	discordBridge    discord.Bot
	messageMutex     sync.Mutex
}

func (bot *Bot) GetEventHandler() func(interface{}) {
	return func(event interface{}) {
		switch v := event.(type) {
		case *events.Message:
			if v != nil {
				jid := v.Info.Chat

				if jid.String() == bot.whatsappGoUAEJID {
					// lock the mutex and wait for the message
					// to be sent before sending the next one
					bot.messageMutex.Lock()
					defer bot.messageMutex.Unlock()

					bot.discordBridge.PipeToDiscord(jid, bot.session, v)
				}
			}
		}
	}
}

func New(cfg config.HummusConfig, bridge discord.Bot) (waBot Bot, err error) {
	// creates a new bot instance
	waBot = Bot{
		whatsappGoUAEJID: cfg.WhatsappGoUAEJID,
		discordBridge:    bridge,
	}

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite as we did in this minimal working example
	container, err := sqlstore.New("sqlite3", "file:hummus.db?_foreign_keys=on", dbLog)
	if err != nil {
		return
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	waBot.session = whatsmeow.NewClient(deviceStore, clientLog)
	waBot.session.AddEventHandler(waBot.GetEventHandler())

	return
}

func (waBot *Bot) Run() (err error) {
	if waBot.session.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := waBot.session.GetQRChannel(context.Background())
		err = waBot.session.Connect()
		if err != nil {
			return
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal:
				// fmt.Println("QR code:", evt.Code)
			} else {
				log.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = waBot.session.Connect()
		if err != nil {
			return
		}
	}
	return
}

func (waBot *Bot) Stop() {
	waBot.session.Disconnect()
}
