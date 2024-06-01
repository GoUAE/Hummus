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

type Bot struct {
	session           *discordgo.Session
	discordBotToken   string
	webhookID         string
	webhookToken      string
	fallbackAvatarURL string
}

func New(cfg config.HummusConfig) (discordBot Bot, err error) {
	discordBot = Bot{
		discordBotToken:   cfg.DiscordBotToken,
		webhookID:         cfg.DiscordWebhookID,
		webhookToken:      cfg.DiscordWebhoolToken,
		fallbackAvatarURL: cfg.FallbackAvatarURL,
	}

	discordBot.session, err = discordgo.New("Bot " + discordBot.discordBotToken)
	if err != nil {
		return
	}

	// sets correct itents for the current session
	discordBot.session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	return
}

func (bot *Bot) Run() (err error) {
	err = bot.session.Open()
	if err != nil {
		return err
	}
	return
}

func (bot *Bot) Stop() (err error) {
	err = bot.session.Close()
	if err != nil {
		return
	}
	return
}

func (bot *Bot) PipeToDiscord(jid types.JID, client *whatsmeow.Client, v *events.Message) {
	// fallback profile picture
	profilePic := bot.fallbackAvatarURL

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
	_, err = bot.session.WebhookExecute(
		bot.webhookID,
		bot.webhookToken,
		true,
		&discordgo.WebhookParams{
			Content:   redactedText,
			Username:  v.Info.PushName,
			AvatarURL: profilePic,
			Files:     files,
		},
	)
	if err != nil {
		utils.LogError(err, "Failed to execute Discord webhook")
	}
}
