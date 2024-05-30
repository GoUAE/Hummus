package whatsapp

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gouae/hummus/internal/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func GetEventHandler(client *whatsmeow.Client) func(interface{}) {
	return func(event interface{}) {
		switch v := event.(type) {
		case *events.Message:
			if v != nil {
				jid := v.Info.Chat

				if jid.String() == config.HummusConfig.WhatsappGoUAEJID {
					// fallback profile picture
					profilePic := "https://wearedesigners.golang.ae/dist/img/gouae-mascot.png"

					userProfilePicInfo, err := client.
						GetProfilePictureInfo(
							v.Info.Sender,
							&whatsmeow.GetProfilePictureParams{},
						)
					if err != nil {
						log.Println(err.Error())
					}

					if userProfilePicInfo != nil {
						profilePic = userProfilePicInfo.URL
					}

					// TODO: check other message types (audio, document, stickers etc.) and handle them properly.
					imageMessage := v.Message.GetImageMessage()
					stickerMessage := v.Message.GetStickerMessage()
					files := []*discordgo.File{}
					var embeds []*discordgo.MessageEmbed

					var imageEmbed discordgo.MessageEmbedImage
					if imageMessage != nil {

						mimeType := "image/jpg"

						mimeTypeRef := imageMessage.Mimetype
						if mimeTypeRef != nil {
							mimeType = *mimeTypeRef
						}

						extension, hasExtension := strings.CutPrefix(mimeType, "image/")
						if !hasExtension {
							extension = "jpg"
						}

						data, _ := client.Download(imageMessage)
						imageEmbed = discordgo.MessageEmbedImage{
							URL:    "attachment://image." + extension,
							Width:  int(v.Message.ImageMessage.GetWidth()),
							Height: int(v.Message.ImageMessage.GetHeight()),
						}
						files = append(files, &discordgo.File{
							Name:        "image." + extension,
							ContentType: mimeType,
							Reader:      bytes.NewBuffer(data),
						})
						embeds = append(embeds, &discordgo.MessageEmbed{
							Type:  discordgo.EmbedTypeImage,
							Image: &imageEmbed,
						})
					}

					// FIXME: fix sticker messages (maybe save them, convert to .gif from .webp and then attach it?)
					if stickerMessage != nil {
						mimeType := "image/webp"

						mimeTypeRef := stickerMessage.Mimetype
						if mimeTypeRef != nil {
							mimeType = *mimeTypeRef
						}

						extension, hasExtension := strings.CutPrefix(mimeType, "image/")
						if !hasExtension {
							extension = "webp"
						}

						data, _ := client.Download(stickerMessage)
						imageEmbed = discordgo.MessageEmbedImage{
							URL:    "attachment://image." + extension,
							Width:  int(v.Message.ImageMessage.GetWidth()),
							Height: int(v.Message.ImageMessage.GetHeight()),
						}
						files = append(files, &discordgo.File{
							Name:        "image." + extension,
							ContentType: mimeType,
							Reader:      bytes.NewBuffer(data),
						})
						embeds = append(embeds, &discordgo.MessageEmbed{
							Type:  discordgo.EmbedTypeImage,
							Image: &imageEmbed,
						})
					}

					newMessageContent := v.Message.GetConversation() +
						v.Message.ExtendedTextMessage.GetText() +
						v.Message.ImageMessage.GetCaption()

					var quotedMessage *waE2E.Message // Declare the final variable as nil initially

					// Check if v.Message exists and is not nil
					if msg := v.Message; msg != nil {
						// Check if v.Message.ExtendedTextMessage exists and is not nil
						if etm := msg.ExtendedTextMessage; etm != nil {
							// Check if v.Message.ExtendedTextMessage.ContextInfo exists and is not nil
							if etmContext := etm.ContextInfo; etmContext != nil {
								// use these strings to get PushNames
								userJIDStrings := etmContext.GetMentionedJID()

								var userJIDs []types.JID

								// find actual JIDs
								for _, userJIDString := range userJIDStrings {
									jid, err := types.ParseJID(userJIDString)
									if err == nil {
										userJIDs = append(userJIDs, jid)
									}
								}

								// Regex pattern with capturing mention
								regex := regexp.MustCompile(`(?i)@([a-zA-Z0-9]+)`)

								// fmt.Println(userInfoMap)

								for _, userJID := range userJIDs {
									// fallback
									replaceWith := "`@[REDACTED]`"
									user, err := client.Store.Contacts.GetContact(userJID)
									if err == nil {
										replaceWith = user.PushName
									}
									newMessageContent = regex.ReplaceAllString(newMessageContent, "`"+replaceWith+"`")
								}

								// Finally, access the GetQuotedMessage() method
								quotedMessage = etmContext.GetQuotedMessage()
							}
						}
					}

					replyContent := ""

					if quotedMessage != nil {
						replyContent = quotedMessage.GetConversation() +
							quotedMessage.ExtendedTextMessage.GetText() +
							quotedMessage.ImageMessage.GetCaption()

						lines := strings.Split(replyContent, "\n")

						// Iterate through each line and add the suffix
						var processedLines []string
						for _, line := range lines {
							processedLine := "> " + line
							processedLines = append(processedLines, processedLine)
						}

						// Join the processed lines back with newline separators
						replyContent = strings.Join(processedLines, "\n") + "\n\n"
					}

					messageContent := replyContent + newMessageContent

					// Regex pattern with capturing mention
					regex := regexp.MustCompile(`(?i)@([a-zA-Z0-9]+)`)

					// FIXME: Replace mentions in replyContent with redacted for now
					// (until we figure out how to get the JID and eventually name for each mention)
					redactedText := regex.ReplaceAllString(messageContent, "`@[REDACTED]`")

					_, err = config.DG.WebhookExecute(
						config.HummusConfig.DiscordWebhookID,
						config.HummusConfig.DiscordWebhoolToken,
						true,
						&discordgo.WebhookParams{
							Content:   redactedText,
							Username:  v.Info.PushName,
							AvatarURL: profilePic,
							Embeds:    embeds,
							Files:     files,
						},
					)
					if err != nil {
						log.Println(err.Error())
					}
				}
			}
		}
	}
}

func RunWhatsappBot() {
	// TODO: replace panics with proper error handling
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite as we did in this minimal working example
	container, err := sqlstore.New("sqlite3", "file:hummus.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(GetEventHandler(client))

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
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
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}