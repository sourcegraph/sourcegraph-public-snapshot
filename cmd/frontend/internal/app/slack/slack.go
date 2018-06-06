// Package slack is used to send notifications of an organization's activity to a
// given Slack webhook. In contrast with package slackinternal, this package contains
// notifications that external users and customers should also be able to receive.
package slack

import (
	"fmt"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/slack"
)

// New is declared so that users of this package don't need to also import pkg/slack
var New = slack.New

// User is an interface for accessing a Sourcegraph user's profile data
type User interface {
	Username() string
	DisplayName() *string
	AvatarURL() *string
}

// NotifyOnInvite posts a message to the defined Slack channel
// when a user invites another user to join their org
func NotifyOnInvite(c *slack.Client, user User, userEmail string, org *types.Org, inviteEmail string) {
	displayNameText := userEmail
	if user.DisplayName() != nil {
		displayNameText = *user.DisplayName()
	}
	usernameText := ""
	if user.Username() != "" {
		usernameText = fmt.Sprintf("(@%s) ", user.Username())
	}

	text := fmt.Sprintf("*%s* %sjust invited %s to join *<https://sourcegraph.com/organizations/%s/settings|%s>*", displayNameText, usernameText, inviteEmail, org.Name, org.Name)

	payload := &slack.Payload{
		Attachments: []*slack.Attachment{
			{
				Fallback:   text,
				Color:      "#F96316",
				Text:       text,
				MarkdownIn: []string{"text"},
			},
		},
	}

	if user.AvatarURL() != nil {
		payload.Attachments[0].ThumbURL = *user.AvatarURL()
	}

	// First, send the notification to the Client's webhookURL
	err := slack.Post(payload, c.WebhookURL)
	if err != nil {
		log15.Error("slack.NotifyOnInvite failed", "error", err)
	}

	// Next, if the action was by an external Sourcegraph customer, also send the
	// notification to the Sourcegraph-internal webhook
	if c.AlsoSendToSourcegraph && slack.SourcegraphOrgWebhookURL != "" && org.Name != "Sourcegraph" {
		err := slack.Post(payload, slack.SourcegraphOrgWebhookURL)
		if err != nil {
			log15.Error("slack.NotifyOnInvite failed", "error", err)
		}
	}
}

// NotifyOnAcceptedInvite posts a message to the defined Slack channel
// when an invited user accepts their invite to join an org
func NotifyOnAcceptedInvite(c *slack.Client, user User, userEmail string, org *types.Org) {
	displayNameText := userEmail
	if user.DisplayName() != nil {
		displayNameText = *user.DisplayName()
	}
	usernameText := ""
	if user.Username() != "" {
		usernameText = fmt.Sprintf("(@%s) ", user.Username())
	}

	text := fmt.Sprintf("*%s* %sjust accepted their invitation to join *<https://sourcegraph.com/organizations/%s/settings|%s>*", displayNameText, usernameText, org.Name, org.Name)

	payload := &slack.Payload{
		Attachments: []*slack.Attachment{
			{
				Fallback:   text,
				Color:      "#B114F7",
				Text:       text,
				MarkdownIn: []string{"text"},
			},
		},
	}

	if user.AvatarURL() != nil {
		payload.Attachments[0].ThumbURL = *user.AvatarURL()
	}

	// First, send the notification to the Client's webhookURL
	err := slack.Post(payload, c.WebhookURL)
	if err != nil {
		log15.Error("slack.NotifyOnAcceptedInvite failed", "error", err)
	}

	// Next, if the action was by an external Sourcegraph customer, also send the
	// notification to the Sourcegraph-internal webhook
	if c.AlsoSendToSourcegraph && slack.SourcegraphOrgWebhookURL != "" && org.Name != "Sourcegraph" {
		err := slack.Post(payload, slack.SourcegraphOrgWebhookURL)
		if err != nil {
			log15.Error("slack.NotifyOnAcceptedInvite failed", "error", err)
		}
	}
}
