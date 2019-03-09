package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"os"
)

var LIKE = "LIKE"
var NEUTRAL = "NEUTRAL"

func ProcessChannelButtonPress(update tgbotapi.Update) {
	chatId := update.CallbackQuery.Message.Chat.ID
	messageId := update.CallbackQuery.Message.MessageID
	SetLikes(
		chatId,
		messageId,
		update.CallbackQuery.From.ID,
		update.CallbackQuery.Data,
	)
	videoId := findYoutubeVideoIDInUpdate(update, false)
	playlistUrl := buildPlaylistUrl(videoId)
	edit := tgbotapi.NewEditMessageReplyMarkup(
		chatId,
		messageId,
		buildKeyboardMarkup(chatId, messageId, playlistUrl),
	)
	BOT.Send(edit)
}

func ProcessChannelMessage(update tgbotapi.Update) {
	SaveMainChatID(update.ChannelPost.Chat.ID)

	videoId := findYoutubeVideoIDInUpdate(update, true)
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

func findYoutubeVideoIDInUpdate(update tgbotapi.Update, addToPlaylist bool) string {
	text := ""
	if update.ChannelPost != nil {
		text = update.ChannelPost.Text
	}
	if update.CallbackQuery != nil {
		text = update.CallbackQuery.Message.Text
	}

	firstVideoIdInPost := FindVideoInText(text, addToPlaylist)
	return firstVideoIdInPost
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
