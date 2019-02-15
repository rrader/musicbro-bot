package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/url"
	"strings"
)

var LIKE = "LIKE"
var NEUTRAL = "NEUTRAL"

func ProcessButtonPress(update tgbotapi.Update) {
	chatId := update.CallbackQuery.Message.Chat.ID
	messageId := update.CallbackQuery.Message.MessageID
	SetLikes(
		chatId,
		messageId,
		update.CallbackQuery.From.ID,
		update.CallbackQuery.Data,
	)
	edit := tgbotapi.NewEditMessageReplyMarkup(
		chatId,
		messageId,
		buildKeyboardMarkup(chatId, messageId),
	)
	BOT.Send(edit)
}

func ProcessMessage(update tgbotapi.Update) {
	shouldAddLikes := false
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

					AddVideoToPlaylist(u.Query().Get("v"))

					shouldAddLikes = true
				}
			}
		}
	}

	if shouldAddLikes {
		markup := buildKeyboardMarkup(
			update.ChannelPost.Chat.ID,
			update.ChannelPost.MessageID,
		)

		edit := tgbotapi.NewEditMessageReplyMarkup(
			update.ChannelPost.Chat.ID,
			update.ChannelPost.MessageID,
			markup,
		)
		BOT.Send(edit)
	}
}

func buildKeyboardMarkup(chatId int64, messageId int) tgbotapi.InlineKeyboardMarkup {
	btnLikeVal := GetLikes(chatId, messageId, LIKE)
	btnNeutralVal := GetLikes(chatId, messageId, NEUTRAL)

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
