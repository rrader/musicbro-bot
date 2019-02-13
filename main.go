package main

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var DB *badger.DB = nil
var BOT *tgbotapi.BotAPI = nil
var LIKE string = "LIKE"
var NEUTRAL string = "NEUTRAL"

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

	BOT.Debug = true

	log.Printf("Authorized on account %s", BOT.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := BOT.GetUpdatesChan(u)

	for update := range updates {
		if update.ChannelPost != nil {
			processMessage(update)
		}
		if update.CallbackQuery != nil {

			// increment
			currentValue := GetLikes(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				update.CallbackQuery.Data,
			)
			SetLikes(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				update.CallbackQuery.From.ID,
				update.CallbackQuery.Data,
				currentValue + 1,
			)

			likes := GetLikes(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				LIKE,
			)
			neutrals := GetLikes(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				NEUTRAL,
			)
			markup := buildKeyboardMarkup(likes, neutrals)

			edit := tgbotapi.NewEditMessageReplyMarkup(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				markup,
			)
			BOT.Send(edit)
		}
	}
}

func GetLikes(chatId int64, messageId int, name string) int {
	var likesNum int
	err := DB.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("%s:%d:%d", name, chatId, messageId)
		item, err := txn.Get([]byte(key))
		if err != nil {
			likesNum = 0
			return nil
		}

		var users []string
		var prevValue string
		if item != nil {
			val, _ := item.Value()
			if val != nil {
				prevValue = string(val)
				users = strings.Split(prevValue, ",")
			}
		}

		likesNum = len(users)

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return likesNum
}

func SetLikes(chatId int64, messageId int, userId int, name string, value int) {
	err := DB.Update(func(txn *badger.Txn) error {
		key := fmt.Sprintf("%s:%d:%d", name, chatId, messageId)

		item, _ := txn.Get([]byte(key))

		var users []string
		var prevValue string
		if item != nil {
			val, _ := item.Value()
			if val != nil {
				prevValue = string(val)
				users = strings.Split(prevValue, ",")
			}
		}

		userIdStr := strconv.Itoa(userId)
		for _, user := range users {
			if user == userIdStr {
				// user has already liked this post
				return nil
			}
		}

		newValue := fmt.Sprintf("%s,%s", prevValue, userIdStr)
		err := txn.Set([]byte(key),[]byte(newValue))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func processMessage(update tgbotapi.Update) {
	if update.ChannelPost.Entities != nil {
		for _, entity := range *update.ChannelPost.Entities {
			if entity.Type == "url" {
				urlStr := string(([]rune(update.ChannelPost.Text))[entity.Offset : entity.Offset+entity.Length])

				u, err := url.Parse(urlStr)
				if err != nil {
					continue
				}

				if strings.Contains(u.Host, "youtube.com") {
					log.Printf("YouTube: %s", urlStr)

					markup := buildKeyboardMarkup(0, 0)

					edit := tgbotapi.NewEditMessageReplyMarkup(
						update.ChannelPost.Chat.ID,
						update.ChannelPost.MessageID,
						markup,
					)
					BOT.Send(edit)
				}
			}
		}
	}
}

func buildKeyboardMarkup(btnLikeVal int, btnNeutralVal int) tgbotapi.InlineKeyboardMarkup {
	btnLikeText := fmt.Sprintf("❤ (%d)", btnLikeVal)
	btnNeutralText := fmt.Sprintf("😐 (%d)", btnNeutralVal)

	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         btnLikeText,
					CallbackData: &LIKE,
				},
				{
					Text:         btnNeutralText,
					CallbackData: &NEUTRAL,
				},
			},
		},
	}
	return markup
}
