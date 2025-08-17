package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendToTelegram(message string) {
	// Replace with your bot token from BotFather
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	// Replace with your own Telegram user ID
	userID := int64(291408730) // <-- Change this to your actual user ID

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	msg := tgbotapi.NewMessage(userID, message)

	_, err = bot.Send(msg)
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

	log.Println("Message sent to Saved Messages successfully!")
}
