package whatsapp

import (
	"context"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"

	wm "go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	wmLog "go.mau.fi/whatsmeow/util/log"
)

type Bot struct {
	session  *wm.Client
	channels map[types.JID]chan *events.Message

	Log *log.Logger
}

func New() (bot Bot, err error) {
	dbLog := wmLog.Stdout("Database", "DEBUG", true)
	clientLog := wmLog.Stdout("Client", "INFO", true)

	bot.Log = log.New(os.Stdout, "[Whatsapp] ", log.LstdFlags)

	container, err := sqlstore.New("sqlite3", "file:hummus.db?_foreign_keys=on", dbLog)
	if err != nil {
		return
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return
	}

	bot.session = wm.NewClient(deviceStore, clientLog)
	bot.session.AddEventHandler(bot.EventHandler())

	return
}

func (bot *Bot) EventHandler() func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case nil: // ignore
		case *events.Message:
			jid := v.Info.Chat
			if ch, ok := bot.channels[jid]; ok {
				ch <- v
			}
		}
	}
}

func (bot *Bot) Bridge(chat string, ch chan *events.Message) (err error) {
	bot.session.Store.Contacts.GetAllContacts()

	// For now, let's solely support getting JID from groups
	groups, err := bot.session.GetJoinedGroups()
	if err != nil {
		return
	}

	for _, group := range groups {
		if group != nil && group.Name == chat {
			bot.channels[group.JID] = ch
			return
		}
	}

	return
}

func (bot *Bot) Run() error {
	if bot.session.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := bot.session.GetQRChannel(context.Background())
		if err := bot.session.Connect(); err != nil {
			return err
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				bot.Log.Println("Login event: ", evt.Event)
			}
		}
	}

	// Already logged in, just connect
	return bot.session.Connect()
}

func (bot *Bot) Stop() {
	bot.session.Disconnect()
}
