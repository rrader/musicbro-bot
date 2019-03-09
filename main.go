package main

import (
	"github.com/dgraph-io/badger"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/rand"
	"os"
	"time"
)

var BOT *tgbotapi.BotAPI = nil

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// init bot
	var err error = nil
	BOT, err = tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	// init db
	opts := badger.DefaultOptions
	opts.Dir = "/tmp/musicbro_db"
	opts.ValueDir = "/tmp/musicbro_db"
	DB, err = badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// init youtube
	ConnectYoutube()

	BOT.Debug = true

	log.Printf("Authorized on account %s", BOT.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	go Scheduler()
	updates, err := BOT.GetUpdatesChan(u)

	for update := range updates {
		if update.ChannelPost != nil {
			ProcessChannelMessage(update)
		}
		if update.CallbackQuery != nil {
			if update.CallbackQuery.Message.Chat.Type == "private" {
				ProcessPrivateButtonPress(update)
			} else {
				ProcessChannelButtonPress(update)
			}
		}
		if update.Message != nil {
			if update.Message.Chat.Type == "private" {
				ProcessPrivateMessage(update)
			}
		}
	}
}
