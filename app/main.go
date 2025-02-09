package main

import (
	"PostInator/botutil"
	"PostInator/draw"
	"os"
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
		}
	}
}
