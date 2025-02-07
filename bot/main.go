package main

import (
	"fmt"
	"os"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	botToken := os.Getenv("TOKEN")

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updates, _ := bot.UpdatesViaLongPolling(nil)

	defer bot.StopLongPolling()

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			text := update.Message.Text

			if text == "" {
				continue
			}

			if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
				photo := update.Message.Photo[len(update.Message.Photo)-1]

				file := telego.InputFile{FileID: photo.FileID}

				document := tu.Document(
					tu.ID(update.Message.Chat.ID),
					file,
				).WithCaption("Look")

				_, err = bot.SendDocument(document)
				if err != nil {
					fmt.Println("Sending file:", err)
					bot.SendMessage(&telego.SendMessageParams{
						ChatID: telego.ChatID{ID: chatID},
						Text:   err.Error(),
					})
				}
			} else {
				runes := []rune(text)
				text = string(runes)

				_, err := bot.SendMessage(&telego.SendMessageParams{
					ChatID: telego.ChatID{ID: chatID},
					Text:   text,
				})
				if err != nil {
					fmt.Println("Sending text:", err)
					bot.SendMessage(&telego.SendMessageParams{
						ChatID: telego.ChatID{ID: chatID},
						Text:   err.Error(),
					})
				}

				document := tu.Document(
					tu.ID(update.Message.Chat.ID),
					tu.File(mustOpen("C:\\Users\\user\\OneDrive\\Изображения\\Фоновые изображения рабочего стола\\AutumnBG_alena.jpg")),
				).WithCaption("Look")

				_, err = bot.SendDocument(document)
				if err != nil {
					fmt.Println("Sending file:", err)
					bot.SendMessage(&telego.SendMessageParams{
						ChatID: telego.ChatID{ID: chatID},
						Text:   err.Error(),
					})
				}
			}
		}
	}
}

func mustOpen(filename string) *os.File {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return file
}
