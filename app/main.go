package main

import (
	"PostInator/draw"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mymmrac/telego"
)

func main() {
	botToken := os.Getenv("TOKEN")

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println("Bot initialization:", err)
		os.Exit(1)
	}

	updates, _ := bot.UpdatesViaLongPolling(nil)
	defer bot.StopLongPolling()

	for update := range updates {
		if update.Message != nil && (len(update.Message.Photo) > 0 || update.Message.Document != nil) {
			chatID := update.Message.Chat.ID
			var fileID string

			if update.Message.Photo != nil {
				fileID = update.Message.Photo[len(update.Message.Photo)-1].FileID
			} else if update.Message.Document != nil {
				fileID = update.Message.Document.FileID
			}

			file, err := bot.GetFile(&telego.GetFileParams{FileID: fileID})
			if err != nil {
				sendError(bot, chatID, "Getting files from Telegram: "+err.Error())
				continue
			}

			localPath := filepath.Join("downloads", filepath.Base(file.FilePath))
			out, err := downloadFile(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", botToken, file.FilePath), localPath)
			if err != nil {
				sendError(bot, chatID, "Downloading file: "+err.Error())
				continue
			}

			_, err = bot.SendMessage(&telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   "The file has been successfully downloaded.",
			})
			if err != nil {
				fmt.Println("Sending message:", err)
			}

			render := draw.Render(out.Name(), "test")

			// Открываем сгенерированный файл
			fileToSend, err := os.Open(render)
			if err != nil {
				sendError(bot, chatID, "Error opening file: "+err.Error())
				continue
			}
			defer fileToSend.Close()

			// Отправляем файл
			_, err = bot.SendDocument(&telego.SendDocumentParams{
				ChatID:   telego.ChatID{ID: chatID},
				Document: telego.InputFile{File: fileToSend},
				Caption:  "Here is your processed file.",
			})
			if err != nil {
				sendError(bot, chatID, "Error sending file: "+err.Error())
			}
		}
	}
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
