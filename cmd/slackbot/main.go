package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func handleAppMention(event *slackevents.AppMentionEvent) {
	log.Printf("App mention event: %+v", event)
	// Add your predicate and actions here for app mention events
}

func handleMessage(event *slackevents.MessageEvent) {
	log.Printf("Message event: %+v", event)
	// Add your predicate and actions here for message events
}
func main() {
	// token := os.Getenv("SLACK_BOT_TOKEN") // Replace this with the actual token or read it from an environment variable or configuration file

	token := "xoxp-5068316257251-5081072338465-5068433831763-d5cae101951172587edbc778a19bfc73"
	api := slack.New(token,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)

	// conversations, _, err := api.GetConversations(&slack.GetConversationsParameters{})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for _, conversation := range conversations {
	// 	imarshal, _ := json.Marshal(conversation)
	// 	fmt.Printf("%v\n", string(marshal))
	// }

	history, err := api.GetConversationHistory(&slack.GetConversationHistoryParameters{ChannelID: "C0522SY342Y"})
	if err != nil {
		return
	}

	jsonify, _ := json.Marshal(history)
	fmt.Printf(string(jsonify))

	// api.GetConversationReplies()

	// Add your event handling code here
	// h.HandleEvents("app_mention", func(event *socketmode.Event, client *socketmode.Client) {
	// handleAppMention(event.Data.(*slackevents.AppMentionEvent))
	// })
	// h.HandleEvents("message", func(event *socketmode.Event, client *socketmode.Client) {
	// 	handleMessage(event.Data.(*slackevents.MessageEvent))
	// })
	// if err := sm.Run(); err != nil {
	// 	log.Fatalf("socketmode.Run() failed: %v", err)
	// }
}
