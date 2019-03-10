package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/nu7hatch/gouuid"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

var DB *badger.DB = nil

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
				if prevValue == "" {
					likesNum = 0
				} else {
					users = strings.Split(prevValue, ",")
				}
			}
		}

		likesNum = len(users) - 1

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return likesNum
}

func SetLikes(chatId int64, messageId int, userId int, name string) {
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

		isLiked := false
		userIdStr := strconv.Itoa(userId)
		for _, user := range users {
			if user == userIdStr {
				// user has already liked this post: UNLIKE
				isLiked = true
			}
		}

		var newValue string
		if isLiked {
			// unlike
			for _, user := range users {
				if user == "" {
					continue
				}

				if user != userIdStr {
					newValue = fmt.Sprintf("%s,%s", user, newValue)
				}
			}
		} else {
			// like
			newValue = fmt.Sprintf("%s,%s", userIdStr, prevValue)
		}

		var err error
		if newValue == "" {
			err = txn.Delete([]byte(key))
		} else {
			err = txn.Set([]byte(key), []byte(newValue))
		}
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

// schedule
type ScheduledMessage struct {
	Id          string
	Message     string
	Timestamp   int64
	IsPublished bool
}

func GetScheduledList(chatId int64) []ScheduledMessage {
	list := make([]ScheduledMessage, 0)

	err := DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(fmt.Sprintf("scheduled:%d", chatId))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.Value()
			if err != nil {
				continue
			}

			msg := ScheduledMessage{}
			err = json.Unmarshal(val, &msg)
			if err != nil {
				continue
			}

			list = append(list, msg)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Timestamp < list[j].Timestamp
	})
	return list
}

func GetScheduledMessage(msgId string) ScheduledMessage {
	msg := ScheduledMessage{}

	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(msgId))
		if err != nil {
			return nil
		}
		val, _ := item.Value()
		_ = json.Unmarshal(val, &msg)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return msg
}

func GetDraftsList(chatId int64) []ScheduledMessage {
	list := make([]ScheduledMessage, 0)

	err := DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(fmt.Sprintf("draft:%d", chatId))
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			log.Printf("Found draft %s", key)
			val, err := item.Value()
			if err != nil {
				continue
			}

			msg := ScheduledMessage{}
			err = json.Unmarshal(val, &msg)
			if err != nil {
				continue
			}

			list = append(list, msg)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return list
}

func SaveScheduledMessage(msg ScheduledMessage) {
	err := DB.Update(func(txn *badger.Txn) error {
		jsonstr, _ := json.Marshal(msg)
		var err error
		if msg.IsPublished {
			err = txn.Set([]byte(msg.Id), jsonstr)
		} else {
			err = txn.SetWithTTL([]byte(msg.Id), jsonstr, 24*time.Hour)
		}
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func RemoveScheduledMessage(msgId string) {
	err := DB.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(msgId))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func GenerateScheduledMessageID(chatId int64) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("scheduled:%d:%s", chatId, u.String())
}

func GenerateScheduledDraftMessageID(chatId int64) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("draft:%d:%s", chatId, u.String())
}

// Chat ID

func SaveMainChatID(chatId int64) {
	err := DB.Update(func(txn *badger.Txn) error {
		chatIdStr := strconv.FormatInt(chatId, 10)
		err := txn.Set([]byte("mainChatId"), []byte(chatIdStr))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func GetMainChatID() int64 {
	var chatId int64
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("mainChatId"))
		if err != nil {
			return nil
		}
		val, _ := item.Value()
		chatId, err = strconv.ParseInt(string(val), 10, 0)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return chatId
}

// Scheduling process

func SaveSchedulingCurrentlyMsg(chatId int64, msgId string) {
	err := DB.Update(func(txn *badger.Txn) error {
		err := txn.SetWithTTL([]byte(fmt.Sprintf("currentlyScheduling:%d", chatId)), []byte(msgId), time.Hour)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func GetSchedulingCurrentlyMsg(chatId int64) string {
	var msgId string
	err := DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("currentlyScheduling:%d", chatId)))
		if err != nil {
			return nil
		}
		val, _ := item.Value()
		msgId = string(val)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return msgId
}
