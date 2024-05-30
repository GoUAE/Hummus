package discord

import (
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gouae/hummus/internal/config"
	"github.com/gouae/hummus/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func PipeToDiscord(jid types.JID, client *whatsmeow.Client, v *events.Message) {
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

	files := utils.WhatsappFilesToDiscordFiles(client, v.Message)

	// TODO: check other message types (audio, document, stickers etc.) and handle them properly.

	var newMessageContent string
	var quotedMessageContent string

	// Check if v.Message exists and is not nil
	if msg := v.Message; msg != nil {
		newMessageContent = utils.ReplaceJIDsWithPushNames(
			client,
			utils.GetJIDs(msg),
			utils.GetMessageCaption(msg),
		)

		if etm := msg.ExtendedTextMessage; etm != nil {
			if etmContext := etm.ContextInfo; etmContext != nil {
				quotedMessage := etmContext.GetQuotedMessage()
				if quotedMessage != nil {
					quotedMessageContent = utils.ReplaceJIDsWithPushNames(
						client,
						utils.GetJIDs(quotedMessage),
						utils.GetMessageCaption(quotedMessage),
					)
					lines := strings.Split(quotedMessageContent, "\n")

					// Iterate through each line and add the suffix
					var processedLines []string
					for _, line := range lines {
						processedLine := "> " + line
						processedLines = append(processedLines, processedLine)
					}

					// Join the processed lines back with newline separators
					quotedMessageContent = strings.Join(processedLines, "\n") + "\n\n"
				}
			}
		}

	}

	messageContent := quotedMessageContent + newMessageContent

	// Regex pattern with capturing mention
	regex := regexp.MustCompile(`(?i)@([a-zA-Z0-9]+)`)

	// INFO: Catches all unreplaces phone numbers and redacts them
	redactedText := regex.ReplaceAllString(messageContent, "`@[REDACTED]`")

	if redactedText == "" && files == nil {
		return
	}
	_, err = config.DG.WebhookExecute(
		config.HummusConfig.DiscordWebhookID,
		config.HummusConfig.DiscordWebhoolToken,
		true,
		&discordgo.WebhookParams{
			Content:   redactedText,
			Username:  v.Info.PushName,
			AvatarURL: profilePic,
			Files:     files,
		},
	)
	if err != nil {
		log.Println(err.Error())
	}
}
