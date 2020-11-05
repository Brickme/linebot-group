// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				switch {
				case event.Source.GroupID != "":
					//In the group
					if strings.EqualFold(message.Text, "bye") {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(" Bye bye!")).Do(); err != nil {
							log.Print(err)
						}
						bot.LeaveGroup(event.Source.GroupID).Do()
					} else {
						//Response with get member profile
						if profile, err := bot.GetGroupMemberProfile(event.Source.GroupID, event.Source.UserID).Do(); err == nil {
							sendUserProfile(*profile, event)
						}
					}

				case event.Source.RoomID != "":
					//In the room
					if strings.EqualFold(message.Text, "bye") {
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(" Bye bye!")).Do(); err != nil {
							log.Print(err)
						}
						bot.LeaveRoom(event.Source.RoomID).Do()
					} else {
						//Response with get member profile
						if profile, err := bot.GetRoomMemberProfile(event.Source.RoomID, event.Source.UserID).Do(); err == nil {
							sendUserProfile(*profile, event)
						}
					}
				default:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(" Hello there :"+message.Text+" OK!")).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		case linebot.EventTypeJoin:
			// If join into a Group
			if event.Source.GroupID != "" {
				if groupRes, err := bot.GetGroupSummary(event.Source.GroupID).Do(); err == nil {
					if goupMemberResult, err := bot.GetGroupMemberCount(event.Source.GroupID).Do(); err == nil {
						retString := fmt.Sprintf("Thanks for letting me join this group, the name of this group is: %s, there are: %d people in total\n", groupRes.GroupName, goupMemberResult.Count)
						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(retString), linebot.NewImageMessage(groupRes.PictureURL, groupRes.PictureURL)).Do(); err != nil {
							//Reply fail.
							log.Print(err)
						}
					} else {
						//GetGroupMemberCount fail.
						log.Printf("GetGroupMemberCount:%x", err)
					}
				} else {
					//GetGroupSummary fail/.
					log.Printf("GetGroupSummary:%x", err)
				}
			} else if event.Source.RoomID != "" {
				// If join into a Room
				if goupMemberResult, err := bot.GetRoomMemberCount(event.Source.RoomID).Do(); err == nil {
					retString := fmt.Sprintf("Thanks for letting me join this chat room, the name of this chat room has: %d people\n", goupMemberResult.Count)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(retString)).Do(); err != nil {
						//Reply fail.
						log.Print(err)
					}
				} else {
					//GetRoomMemberCount fail.
					log.Printf("GetRoomMemberCount:%x", err)
				}
			}
		}
	}
}

func sendUserProfile(user linebot.UserProfileResponse, event *linebot.Event) {
	retString := fmt.Sprintf("Hello user %s, your ID is %s, your language is %s, and your status is: %s\n", user.DisplayName, user.UserID, user.Language, user.StatusMessage)
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(retString), linebot.NewImageMessage(user.PictureURL, user.PictureURL)).Do(); err != nil {
		//Reply fail.
		log.Print(err)
	}
}
