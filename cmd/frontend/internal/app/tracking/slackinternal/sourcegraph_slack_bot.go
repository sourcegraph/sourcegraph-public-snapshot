package slackinternal

import (
	"fmt"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
)

var signupsWebhookURL = env.Get("SLACK_SIGNUPS_BOT_HOOK", "", "Webhook for posting signup notifications to the Slack #bot-signups channel.")

// NotifyOnSignup posts a message to the Slack channel #bot-signups
// when a user signs up for Sourcegraph
func NotifyOnSignup(actor *actor.Actor, hubSpotProps *hubspot.ContactProperties, response *hubspot.ContactResponse) error {
	client := slack.New(signupsWebhookURL)

	var links []string
	if hubSpotProps.LookerLink != "" {
		links = append(links, fmt.Sprintf("<%s|View on Looker>", hubSpotProps.LookerLink))
	}
	if response != nil {
		links = append(links, fmt.Sprintf("<https://app.hubspot.com/contacts/2762526/contact/%v|View on HubSpot>", response.VID))
	}

	payload := &slack.Payload{
		Attachments: []*slack.Attachment{
			&slack.Attachment{
				Fallback: fmt.Sprintf("%s just signed up!", actor.Email),
				Title:    fmt.Sprintf("%s just signed up!", actor.Email),
				Color:    "good",
				ThumbURL: actor.AvatarURL,
				Fields: []*slack.Field{
					&slack.Field{
						Title: "User profile links",
						Value: strings.Join(links, ", "),
						Short: false,
					},
				},
			},
		},
	}

	return client.Post(payload)
}
