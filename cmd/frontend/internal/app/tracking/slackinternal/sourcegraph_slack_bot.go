// Package slackinternal is used to send notifications to Sourcegraph. Unlike
// package slack, these notifications would never be consumed by external users
// or customers, and will only ever be sent to specific Sourcegraph webhooks.
package slackinternal

import (
	"fmt"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/slack"
)

var signupsWebhookURL = env.Get("SLACK_SIGNUPS_BOT_HOOK", "", "Webhook for posting signup notifications to the Slack #bot-signups channel.")

// NotifyOnSignup posts a message to the Slack channel #bot-signups
// when a user signs up for Sourcegraph
func NotifyOnSignup(avatarURL, email string, hubSpotProps *hubspot.ContactProperties, response *hubspot.ContactResponse) error {
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
				Fallback: fmt.Sprintf("%s just signed up!", email),
				Title:    fmt.Sprintf("%s just signed up!", email),
				Color:    "good",
				ThumbURL: avatarURL,
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

	return slack.Post(payload, signupsWebhookURL)
}
