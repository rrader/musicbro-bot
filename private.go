package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/now"
	"log"
	"strings"
	"time"
)

var SCHEDULE = "S"
var HELP = "L"
var REMOVE = "M"
var RIGHT_NOW = "R"

const HELP_STRING = "сделай /drafts чтоб посмотреть черновики,\n/list чтоб посмотреть список запланированных отправок"

func ProcessPrivateMessage(update tgbotapi.Update) {
	if update.Message.IsCommand() {
		if update.Message.Command() == "drafts" {
			drafts := GetDraftsList(GetMainChatID())
			for _, draft := range drafts {
				markup := buildKeyboardMarkupPrivate(
					draft,
				)

				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					draft.Message,
				)
				msg.ReplyMarkup = markup

				BOT.Send(msg)
			}
		}
		if update.Message.Command() == "list" {
			drafts := GetScheduledList(GetMainChatID())
			for _, draft := range drafts {
				markup := buildKeyboardMarkupPrivate(
					draft,
				)

				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					draft.Message,
				)
				msg.ReplyMarkup = markup

				BOT.Send(msg)
			}
		}
		if update.Message.Command() == "help" {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				HELP_STRING,
			)

			BOT.Send(msg)
		}
		return
	}

	currentlyScheduling := GetSchedulingCurrentlyMsg(update.Message.Chat.ID)
	if currentlyScheduling != "" {
		currentMessage := GetScheduledMessage(currentlyScheduling)
		if currentMessage.Id == "" {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"такого нет",
			)
			BOT.Send(msg)
			SaveSchedulingCurrentlyMsg(update.Message.Chat.ID, "")
			return
		}
		dt, err := now.Parse(update.Message.Text)
		if err != nil {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"не получилось, попробуй еще раз",
			)

			BOT.Send(msg)
			return
		}
		if time.Now().Unix() > dt.Unix() {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"это уже в прошлом, попробуй еще раз",
			)

			BOT.Send(msg)
			return
		}

		RemoveScheduledMessage(currentMessage.Id)
		currentMessage.Timestamp = dt.Unix()

		currentMessage.Id = GenerateScheduledMessageID(GetMainChatID())
		SaveScheduledMessage(currentMessage)

		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf("ладн. будет отправлено в %s", dt),
		)

		BOT.Send(msg)

		SaveSchedulingCurrentlyMsg(update.Message.Chat.ID, "")

		markup := buildKeyboardMarkupPrivate(
			currentMessage,
		)

		msg = tgbotapi.NewMessage(
			update.Message.Chat.ID,
			currentMessage.Message,
		)
		msg.ReplyMarkup = markup

		BOT.Send(msg)
		return
	}

	draft := ScheduledMessage{
		Id:          GenerateScheduledDraftMessageID(GetMainChatID()),
		IsPublished: false,
		Message:     update.Message.Text,
	}
	SaveScheduledMessage(draft)

	markup := buildKeyboardMarkupPrivate(
		draft,
	)

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"Сохранил черновик (черновики живут сутки), что-то сделать?",
	)
	msg.ReplyMarkup = markup

	BOT.Send(msg)
}

func buildKeyboardMarkupPrivate(draft ScheduledMessage) tgbotapi.InlineKeyboardMarkup {
	draftId := draft.Id
	scheduleText := "Запланировать отправку"
	if draft.Timestamp > 0 {
		scheduleText = fmt.Sprintf("Будет отправлено: %s (перепланировать)", time.Unix(draft.Timestamp, 0))
	}

	schedule := fmt.Sprintf("%s#%s", SCHEDULE, draftId)
	send_now := fmt.Sprintf("%s#%s", RIGHT_NOW, draftId)
	help := fmt.Sprintf("%s#%s", HELP, draftId)
	remove := fmt.Sprintf("%s#%s", REMOVE, draftId)

	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         scheduleText,
					CallbackData: &schedule,
				},
			},
			{
				{
					Text:         "Отправляй прям сейчас!",
					CallbackData: &send_now,
				},
			},
			{

				{
					Text:         "Удали",
					CallbackData: &remove,
				},
			},
			{

				{
					Text:         "Помощь",
					CallbackData: &help,
				},
			},
		},
	}
	return markup
}

func ProcessPrivateButtonPress(update tgbotapi.Update) {
	parts := strings.Split(update.CallbackQuery.Data, "#")
	if len(parts) != 2 {
		return
	}

	if parts[0] == HELP {
		msg := tgbotapi.NewMessage(
			update.CallbackQuery.Message.Chat.ID,
			HELP_STRING,
		)

		BOT.Send(msg)
	}
	if parts[0] == SCHEDULE {
		SaveSchedulingCurrentlyMsg(update.CallbackQuery.Message.Chat.ID, parts[1])
		msg := tgbotapi.NewMessage(
			update.CallbackQuery.Message.Chat.ID,
			fmt.Sprintf("Введи дату (например %s)", now.BeginningOfMinute()),
		)

		BOT.Send(msg)
	}
	if parts[0] == REMOVE {
		RemoveScheduledMessage(parts[1])
		msg := tgbotapi.NewMessage(
			update.CallbackQuery.Message.Chat.ID,
			"удалил",
		)

		BOT.Send(msg)
	}
	if parts[0] == RIGHT_NOW {
		publishDraft(parts[1])
		msg := tgbotapi.NewMessage(
			update.CallbackQuery.Message.Chat.ID,
			"запощено!",
		)
		BOT.Send(msg)
	}

	editMsg := tgbotapi.NewEditMessageReplyMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		},
	)
	BOT.Send(editMsg)
}

func publishDraft(draftId string) {
	scheduledMessage := GetScheduledMessage(draftId)
	if scheduledMessage.Id == "" {
		return
	}
	videoId := FindVideoInText(scheduledMessage.Message, true)
	playlistUrl := buildPlaylistUrl(videoId)
	msg := tgbotapi.NewMessage(
		GetMainChatID(),
		scheduledMessage.Message,
	)
	resp, err := BOT.Send(msg)
	if err != nil {
		log.Fatal(err)
	}
	if videoId != "" {
		edit := tgbotapi.NewEditMessageReplyMarkup(
			resp.Chat.ID,
			resp.MessageID,
			buildKeyboardMarkup(
				GetMainChatID(),
				resp.MessageID,
				playlistUrl,
			),
		)
		BOT.Send(edit)
	}
	RemoveScheduledMessage(draftId)
}

func doSchedule() {
	msges := GetScheduledList(GetMainChatID())
	for _, msg := range msges {
		if time.Now().Unix() > msg.Timestamp {
			publishDraft(msg.Id)
		}
	}
}

func Scheduler() {
	for {
		go doSchedule()
		time.Sleep(time.Minute)
	}
}
