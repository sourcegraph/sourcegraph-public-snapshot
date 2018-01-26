// Package slack is used to send notifications of an organization's activity to a
// given Slack webhook. In contrast with package slackinternal, this package contains
// notifications that external users and customers should also be able to receive.
package slack

import (
	"fmt"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/slack"
)

// New is declared so that users of this package don't need to also import pkg/slack
var New = slack.New

// User is an interface for accessing a Sourcegraph user's profile data
type User interface {
	Username() string
	DisplayName() *string
	AvatarURL() *string
}

// NotifyOnComment posts a message to the defined Slack channel
// when a user posts a reply to a thread
func NotifyOnComment(
	c *slack.Client,
	user User,
	userEmail string,
	org *types.Org,
	orgRepo *types.OrgRepo,
	thread *types.Thread,
	comment *types.Comment,
	recipients []string,
	deepURL string,
	threadTitle string,
) {
	// First, send the uncensored comment to the Client's webhookURL
	err := notifyOnComments(c, false, user, userEmail, org, orgRepo, thread, comment, recipients, deepURL, threadTitle, c.WebhookURL, false)
	if err != nil {
		log15.Error("slack.NotifyOnComment failed", "error", err)
	}

	// Next, if the comment was made by an external Sourcegraph customer, ALSO send the
	// comment to the Sourcegraph-internal webhook. In these instances, set censored
	// to true to ensure that the private contents of the comment remain private.
	if c.AlsoSendToSourcegraph && slack.SourcegraphOrgWebhookURL != "" && org.Name != "Sourcegraph" {
		err := notifyOnComments(c, false, user, userEmail, org, orgRepo, thread, comment, recipients, deepURL, threadTitle, &slack.SourcegraphOrgWebhookURL, true)
		if err != nil {
			log15.Error("slack.NotifyOnThread failed", "error", err)
		}
	}
}

// NotifyOnThread posts a message to the defined Slack channel
// when a user creates a thread
func NotifyOnThread(
	c *slack.Client,
	user User,
	userEmail string,
	org *types.Org,
	orgRepo *types.OrgRepo,
	thread *types.Thread,
	comment *types.Comment,
	recipients []string,
	deepURL string,
) {
	// First, send the uncensored comment to the Client's webhookURL
	err := notifyOnComments(c, true, user, userEmail, org, orgRepo, thread, comment, recipients, deepURL, "", c.WebhookURL, false)
	if err != nil {
		log15.Error("slack.NotifyOnThread failed", "error", err)
	}

	// Next, if the comment was made by an external Sourcegraph customer, ALSO send the
	// comment to the Sourcegraph-internal webhook. In these instances, set censored
	// to true to ensure that the private contents of the comment remain private.
	if c.AlsoSendToSourcegraph && slack.SourcegraphOrgWebhookURL != "" && org.Name != "Sourcegraph" {
		err := notifyOnComments(c, true, user, userEmail, org, orgRepo, thread, comment, recipients, deepURL, "", &slack.SourcegraphOrgWebhookURL, true)
		if err != nil {
			log15.Error("slack.NotifyOnThread failed", "error", err)
		}
	}
}

func notifyOnComments(
	c *slack.Client,
	isNewThread bool,
	user User,
	userEmail string,
	org *types.Org,
	orgRepo *types.OrgRepo,
	thread *types.Thread,
	comment *types.Comment,
	recipients []string,
	deepURL string,
	threadTitle string,
	webhookURL *string,
	censored bool,
) error {
	color := "good"
	actionText := "created a thread"
	if !isNewThread {
		color = "warning"
		if !censored {
			if len(threadTitle) > 75 {
				threadTitle = threadTitle[0:75] + "..."
			}
			actionText = fmt.Sprintf("replied to a thread: \"%s\"", threadTitle)
		} else {
			actionText = fmt.Sprintf("replied to a thread")
		}
	}
	text := "_private_"
	if !censored {
		text = comment.Contents
	}

	displayNameText := userEmail
	if user.DisplayName() != nil {
		displayNameText = *user.DisplayName()
	}
	usernameText := ""
	if user.Username() != "" {
		usernameText = fmt.Sprintf("(@%s) ", user.Username())
	}
	payload := &slack.Payload{
		Attachments: []*slack.Attachment{
			&slack.Attachment{
				AuthorName: fmt.Sprintf("%s %s%s", displayNameText, usernameText, actionText),
				AuthorLink: deepURL,
				Fallback:   fmt.Sprintf("%s %s<%s|%s>!", displayNameText, usernameText, deepURL, actionText),
				Color:      color,
				Fields: []*slack.Field{
					&slack.Field{
						Title: "Org",
						Value: fmt.Sprintf("`%s`\n(%d member(s) notified)", org.Name, len(recipients)),
						Short: true,
					},
				},
				Text:       text,
				MarkdownIn: []string{"text", "fields"},
			},
		},
	}

	if !censored {
		payload.Attachments[0].Fields = append([]*slack.Field{
			&slack.Field{
				Title: "Path",
				Value: fmt.Sprintf("<%s|%s/%s (lines %d–%d)>",
					deepURL,
					orgRepo.CanonicalRemoteID,
					thread.RepoRevisionPath,
					thread.StartLine,
					thread.EndLine),
				Short: true,
			},
		}, payload.Attachments[0].Fields...)
	} else {
		payload.Attachments[0].Fields = append(payload.Attachments[0].Fields,
			&slack.Field{
				Title: "IDs",
				Value: fmt.Sprintf("Comment ID: %d\nThread ID: %d", comment.ID, thread.ID),
				Short: true,
			})
	}

	if user.AvatarURL() != nil {
		payload.Attachments[0].ThumbURL = *user.AvatarURL()
		payload.Attachments[0].AuthorIcon = *user.AvatarURL()
	}

	return slack.Post(payload, webhookURL)
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
			&slack.Attachment{
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
		err := slack.Post(payload, &slack.SourcegraphOrgWebhookURL)
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
			&slack.Attachment{
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
		err := slack.Post(payload, &slack.SourcegraphOrgWebhookURL)
		if err != nil {
			log15.Error("slack.NotifyOnAcceptedInvite failed", "error", err)
		}
	}
}
