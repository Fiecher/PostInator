package botutil

import (
	"PostInator/draw"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mymmrac/telego"
)

func Initialization(token string) *telego.Bot {

	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println("Bot initialization:", err)
		os.Exit(1)
	}

	return bot
}

func GetFile(update telego.Update, bot *telego.Bot) *telego.File {
	message := update.Message
	chatID := update.Message.Chat.ID

	var fileID string
	if message.Photo != nil {
		fileID = message.Photo[len(message.Photo)-1].FileID
	} else if message.Document != nil {
		fileID = message.Document.FileID
	}

	file, err := bot.GetFile(&telego.GetFileParams{FileID: fileID})
	if err != nil {
		sendError(bot, chatID, "Getting files from Telegram: "+err.Error())
	}

	return file
}

func DownloadFile(update telego.Update, bot *telego.Bot, token string) *os.File {
	file := GetFile(update, bot)
	chatID := update.Message.Chat.ID

	localPath := filepath.Join("temp", filepath.Base(file.FilePath))
	out, err := downloadFile(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, file.FilePath), localPath)
	if err != nil {
		sendError(bot, chatID, "Downloading file: "+err.Error())
	}

	bot.SendMessage(&telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "✅ File has been successfully downloaded!",
	})

	return out
}

func GetText(update telego.Update) string {
	text := ""
	if update.Message.Caption != "" {
		text = update.Message.Caption
	} else if update.Message.Text != "" {
		text = update.Message.Text
	}
	return text
}

func SendFile(update telego.Update, bot *telego.Bot, file string) bool {
	chatID := update.Message.Chat.ID

	resizedPath, err := draw.ResizeImage(file, 2048)
	if err != nil {
		sendError(bot, chatID, "Error resizing image: "+err.Error())
	}

	fileInfo, err := os.Stat(resizedPath)
	if err != nil {
		sendError(bot, chatID, "Error getting file size: "+err.Error())
	}

	fileToSend, err := os.Open(resizedPath)
	if err != nil {
		sendError(bot, chatID, "Error opening file: "+err.Error())
	}

	defer fileToSend.Close()

	if fileInfo.Size() <= 10*1024*1024 {
		_, err = bot.SendPhoto(&telego.SendPhotoParams{
			ChatID:  telego.ChatID{ID: chatID},
			Photo:   telego.InputFile{File: fileToSend},
			Caption: "Here is your processed image.",
		})
	} else {
		_, err = bot.SendDocument(&telego.SendDocumentParams{
			ChatID:   telego.ChatID{ID: chatID},
			Document: telego.InputFile{File: fileToSend},
			Caption:  "Image is too large, sending as document.",
		})
	}

	if err != nil {
		sendError(bot, chatID, "Error sending image: "+err.Error())
		return false
	}
	return true

}

func downloadFile(url, filepath string) (*os.File, error) {
	out, err := os.Create(filepath)
	if err != nil {
		return out, err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return out, err
}

func sendError(bot *telego.Bot, chatID int64, message string) {
	fmt.Println("Error:", message)
	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "⚠️ " + message,
	})
	if err != nil {
		fmt.Println("Sending error:", err)
	}
}
