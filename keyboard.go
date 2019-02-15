package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/url"
	"os"
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
	videoId := findYoutubeVideoID(update)
	playlistUrl := buildPlaylistUrl(videoId)
	edit := tgbotapi.NewEditMessageReplyMarkup(
		chatId,
		messageId,
		buildKeyboardMarkup(chatId, messageId, playlistUrl),
	)
	BOT.Send(edit)
}

func ProcessMessage(update tgbotapi.Update) {
	videoId := findYoutubeVideoID(update)
	playlistUrl := buildPlaylistUrl(videoId)

	if videoId != "" {
		markup := buildKeyboardMarkup(
			update.ChannelPost.Chat.ID,
			update.ChannelPost.MessageID,
			playlistUrl,
		)

		edit := tgbotapi.NewEditMessageReplyMarkup(
			update.ChannelPost.Chat.ID,
			update.ChannelPost.MessageID,
			markup,
		)
		BOT.Send(edit)
	}
}

func buildPlaylistUrl(videoId string) string {
	playlist := os.Getenv("YOUTUBE_PLAYLIST")
	return fmt.Sprintf(
		"https://www.youtube.com/watch?list=%s&v=%s",
		playlist,
		videoId,
	)
}

func findYoutubeVideoID(update tgbotapi.Update) string {
	videoId := ""
	text := ""
	var entities *[]tgbotapi.MessageEntity
	if update.ChannelPost != nil {
		entities = update.ChannelPost.Entities
		text = update.ChannelPost.Text
	}
	if update.CallbackQuery != nil {
		entities = update.CallbackQuery.Message.Entities
		text = update.CallbackQuery.Message.Text
	}

	if entities != nil {
		for _, entity := range *entities {
			if entity.Type == "url" {
				urlStr := string(([]rune(text))[entity.Offset : entity.Offset+entity.Length])

				u, err := url.Parse(urlStr)
				if err != nil {
					continue
				}

				if strings.Contains(u.Host, "youtube.com") {
					log.Printf("YouTube: %s", urlStr)
					videoId = u.Query().Get("v")
				}
			}
		}
	}
	return videoId
}

func buildKeyboardMarkup(chatId int64, messageId int, link string) tgbotapi.InlineKeyboardMarkup {
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
			{
				{
					Text: "▶️ Open in playlist",
					URL:  &link,
				},
			},
		},
	}
	return markup
}
