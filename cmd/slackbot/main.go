package main

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
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
	token := os.Getenv("SLACK_BOT_TOKEN") // Replace this with the actual token or read it from an environment variable or configuration file

	api := slack.New(token,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)
	sm := socketmode.New(api)
	h := socketmode.NewSocketmodeHandler(sm)

	// Add your event handling code here
	h.HandleEvents("app_mention", func(event *socketmode.Event, client *socketmode.Client) {
		handleAppMention(event.Data.(*slackevents.AppMentionEvent))
	})
	h.HandleEvents("message", func(event *socketmode.Event, client *socketmode.Client) {
		handleMessage(event.Data.(*slackevents.MessageEvent))
	})
	if err := sm.Run(); err != nil {
		log.Fatalf("socketmode.Run() failed: %v", err)
	}
}
