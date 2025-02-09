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
	if checkError("Bot initialization", err) {
		os.Exit(1)
	}
	return bot
}

func GetFile(update telego.Update, bot *telego.Bot) *telego.File {
	message := update.Message
	chatID := message.Chat.ID

	var fileID string
	if message.Photo != nil {
		fileID = message.Photo[len(message.Photo)-1].FileID
	} else if message.Document != nil {
		fileID = message.Document.FileID
	}

	file, err := bot.GetFile(&telego.GetFileParams{FileID: fileID})
	if checkError("Getting files from Telegram", err) {
		sendError(bot, chatID, "Getting files from Telegram: "+err.Error())
	}

	return file
}

func DownloadFile(update telego.Update, bot *telego.Bot, token string) *os.File {
	file := GetFile(update, bot)
	chatID := update.Message.Chat.ID

	localPath := filepath.Join("temp", filepath.Base(file.FilePath))
	out, err := fetchFile(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, file.FilePath), localPath)
	if checkError("Downloading file", err) {
		sendError(bot, chatID, "Downloading file: "+err.Error())
	}

	bot.SendMessage(&telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "✅ File has been successfully downloaded!",
	})

	return out
}

func GetText(update telego.Update) string {
	if update.Message.Caption != "" {
		return update.Message.Caption
	}
	return update.Message.Text
}

func SendFile(update telego.Update, bot *telego.Bot, file string) bool {
	defer clearPhotos()
	chatID := update.Message.Chat.ID

	resizedPath, err := draw.ResizeImage(file, 2048)
	if checkError("Resizing image", err) {
		sendError(bot, chatID, "Resizing image: "+err.Error())
	}

	fileInfo, err := os.Stat(resizedPath)
	if checkError("Getting file size", err) {
		sendError(bot, chatID, "Getting file size: "+err.Error())
	}

	fileToSend, err := os.Open(resizedPath)
	if checkError("Opening file", err) {
		sendError(bot, chatID, "Opening file: "+err.Error())
	}
	defer fileToSend.Close()

	var sendErr error
	if fileInfo.Size() <= 10*1024*1024 {
		_, sendErr = bot.SendPhoto(&telego.SendPhotoParams{
			ChatID: telego.ChatID{ID: chatID},
			Photo:  telego.InputFile{File: fileToSend},
		})
	} else {
		_, sendErr = bot.SendDocument(&telego.SendDocumentParams{
			ChatID:   telego.ChatID{ID: chatID},
			Document: telego.InputFile{File: fileToSend},
		})
	}

	if checkError("Error sending image", sendErr) {
		sendError(bot, chatID, "Error sending image: "+sendErr.Error())
		return false
	}
	return true
}

func fetchFile(url, filepath string) (*os.File, error) {
	out, err := os.Create(filepath)
	if checkError("Creating file", err) {
		return nil, err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if checkError("Downloading file", err) {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return out, err
}

func clearPhotos() {
	tempDir := "temp"
	files, err := os.ReadDir(tempDir)
	if checkError("Reading photos directory", err) {
		return
	}

	for _, file := range files {
		if checkError("Deleting file", os.Remove(filepath.Join(tempDir, file.Name()))) {
			fmt.Println("Error! Deleting file:", file.Name())
		}
	}
}

func sendError(bot *telego.Bot, chatID int64, message string) {
	fmt.Println("Error!", message)
	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   "⚠️ " + message,
	})
	checkError("Sending error", err)
}

func checkError(msg string, err error) bool {
	if err != nil {
		fmt.Println(msg+":", err)
		return true
	}
	return false
}
