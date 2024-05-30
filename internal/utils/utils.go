package utils

import (
	"bytes"
	"fmt"
	"log"
	"mime"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func ConvertWebpToX(inputBuffer *bytes.Buffer, outformat string) (*bytes.Buffer, error) {
	// Create temporary files for input and output
	inputFile, err := os.CreateTemp("", "input-*.webp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp input file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.CreateTemp("", "output-*."+outformat)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output file: %v", err)
	}
	defer outputFile.Close()

	// Write the input buffer to the temporary input file
	_, err = inputFile.Write(inputBuffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write to temp input file: %v", err)
	}

	// Close the input file to flush the data
	inputFile.Close()

	// Run imagemagick "convert" command to convert the input file to res
	cmd := exec.Command("convert", inputFile.Name(), outputFile.Name())
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("magick command failed: %v", err)
	}

	// Read the output file into a buffer
	outputData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read temp output file: %v", err)
	}

	outputBuffer := bytes.NewBuffer(outputData)

	return outputBuffer, nil
}

func ReplaceFirst(input, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(input)
	if loc == nil {
		return input
	}
	return input[:loc[0]] + replacement + input[loc[1]:]
}

func AppendJIDs(message any, userJIDStrings []string) []string {
	messageValue := reflect.ValueOf(message)

	// Check if message is a pointer and its underlying value is not nil
	if messageValue.Kind() == reflect.Ptr && !messageValue.IsNil() {
		messageElem := messageValue.Elem()

		// Check if messageElem has a field named "ContextInfo"
		contextInfoField := messageElem.FieldByName("ContextInfo")
		if contextInfoField.IsValid() && !contextInfoField.IsNil() {
			// Get the actual ContextInfo value
			contextInfoValue := contextInfoField.Interface()

			// Assert the type of contextInfoValue to *ContextInfo
			if contextInfo, ok := contextInfoValue.(*waE2E.ContextInfo); ok {
				// Append the mentioned JIDs to the userJIDStrings slice
				userJIDStrings = append(userJIDStrings, contextInfo.GetMentionedJID()...)
			}
		}
	}

	return userJIDStrings
}

func GetJIDs(msg *waE2E.Message) (userJIDs []types.JID) {
	userJIDStrings := []string{}

	// Check if msg.ExtendedTextMessage exists and is not nil
	if etm := msg.ExtendedTextMessage; etm != nil {
		// Check if msg.ExtendedTextMessage.ContextInfo exists and is not nil
		if etmContext := etm.ContextInfo; etmContext != nil {
			// use these strings to get PushNames
			userJIDStrings = etmContext.GetMentionedJID()
		}
	}

	userJIDStrings = AppendJIDs(msg.GetImageMessage(), userJIDStrings)
	userJIDStrings = AppendJIDs(msg.GetVideoMessage(), userJIDStrings)
	userJIDStrings = AppendJIDs(msg.GetAudioMessage(), userJIDStrings)
	userJIDStrings = AppendJIDs(msg.GetDocumentMessage(), userJIDStrings)

	// find actual JIDs
	for _, userJIDString := range userJIDStrings {
		jid, err := types.ParseJID(userJIDString)
		if err == nil {
			userJIDs = append(userJIDs, jid)
		}
	}

	return
}

func ReplaceJIDsWithPushNames(client *whatsmeow.Client, userJIDs []types.JID, message string) string {
	for _, userJID := range userJIDs {
		// fallback
		replaceWith := "`@[REDACTED]`"
		user, err := client.Store.Contacts.GetContact(userJID)
		if err == nil {
			replaceWith = user.PushName
		}
		message = ReplaceFirst(
			message,
			`(?i)@([a-zA-Z0-9]+)`,
			"`"+replaceWith+"`",
		)
	}
	return message
}

func GetMessageCaption(msg *waE2E.Message) string {
	return msg.GetConversation() +
		msg.ExtendedTextMessage.GetText() +
		msg.ImageMessage.GetCaption() +
		msg.VideoMessage.GetCaption() +
		msg.DocumentMessage.GetCaption()
}

// TODO: cleanup this function
func WhatsappFilesToDiscordFiles(client *whatsmeow.Client, msg *waE2E.Message) (files []*discordgo.File) {
	if msg == nil {
		return
	}

	imageMessage := msg.GetImageMessage()
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
		files = append(files, &discordgo.File{
			Name:        "image." + extension,
			ContentType: mimeType,
			Reader:      bytes.NewBuffer(data),
		})
	}

	audioMessage := msg.GetAudioMessage()
	if audioMessage != nil {

		mimeType := "audio/mp3"

		mimeTypeRef := audioMessage.Mimetype
		if mimeTypeRef != nil {
			mimeType = *mimeTypeRef
		}

		extension := "mp3"
		extensions, _ := mime.ExtensionsByType(mimeType)
		if len(extensions) > 0 {
			extension = extensions[0]
		}

		data, _ := client.Download(audioMessage)
		files = append(files, &discordgo.File{
			Name:        "audio." + extension,
			ContentType: mimeType,
			Reader:      bytes.NewBuffer(data),
		})
	}

	documentMessage := msg.GetDocumentMessage()
	if documentMessage != nil {

		mimeType := "text/unknown"

		mimeTypeRef := documentMessage.Mimetype
		if mimeTypeRef != nil {
			mimeType = *mimeTypeRef
		}

		extension := "unknown"
		extensions, _ := mime.ExtensionsByType(mimeType)
		if len(extensions) > 0 {
			extension = extensions[0]
		}

		data, _ := client.Download(documentMessage)
		files = append(files, &discordgo.File{
			Name:        "document." + extension,
			ContentType: mimeType,
			Reader:      bytes.NewBuffer(data),
		})
	}

	videoMessage := msg.GetVideoMessage()
	if videoMessage != nil {

		mimeType := "video/mp4"

		mimeTypeRef := videoMessage.Mimetype
		if mimeTypeRef != nil {
			mimeType = *mimeTypeRef
		}

		extension := "mp4"

		data, _ := client.Download(videoMessage)
		files = append(files, &discordgo.File{
			Name:        "videoMessage." + extension,
			ContentType: mimeType,
			Reader:      bytes.NewBuffer(data),
		})
	}

	stickerMessage := msg.GetStickerMessage()

	if stickerMessage != nil {
		mimeType := "image/png"

		mimeTypeRef := stickerMessage.Mimetype
		if mimeTypeRef != nil {
			mimeType = *mimeTypeRef
		}

		extension := "png"
		extensions, _ := mime.ExtensionsByType(mimeType)
		if len(extensions) > 0 {
			extension = extensions[0]
		}

		data, _ := client.Download(stickerMessage)

		reader := bytes.NewBuffer(data)
		var err error

		if stickerMessage.GetIsAnimated() {
			mimeType = "image/gif"
			extension = "gif"
			reader, err = ConvertWebpToX(reader, "gif")
			if err != nil {
				log.Println("Failed to convert to gif", err.Error())
			}
		} else {
			reader, err = ConvertWebpToX(reader, "png")
			if err != nil {
				log.Println("Failed to convert to gif", err.Error())
			}
		}

		files = append(files, &discordgo.File{
			Name:        "image." + extension,
			ContentType: mimeType,
			Reader:      reader,
		})
	}

	return
}
