package main

import (
	"PostInator/botutil"
	"PostInator/draw"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	token := os.Getenv("TOKEN")
	botObj := botutil.Initialization(token)

	updates, _ := botObj.UpdatesViaLongPolling(nil)
	defer botObj.StopLongPolling()

	photoSent := make(map[int64]bool)

	for update := range updates {
		if update.Message != nil && (len(update.Message.Photo) > 0 || update.Message.Document != nil) {
			chatID := update.Message.Chat.ID

			if photoSent[chatID] {
				continue
			}

			out := botutil.DownloadFile(update, botObj, token)

			text := botutil.GetText(update)

			render := draw.Render(out.Name(), text)

			if botutil.SendFile(update, botObj, render) {
				photoSent[chatID] = true
			}

			clearPhotos()
		}
	}
}

func clearPhotos() {
	photosDir := "photos"
	files, err := os.ReadDir(photosDir)
	if err != nil {
		fmt.Println("Error reading photos directory:", err)
		return
	}

	for _, file := range files {
		err := os.Remove(filepath.Join(photosDir, file.Name()))
		if err != nil {
			fmt.Println("Error deleting file:", file.Name(), err)
		}
	}
}
