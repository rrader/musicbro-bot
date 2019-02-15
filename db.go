package main

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"strconv"
	"strings"
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
