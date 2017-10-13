package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	log15 "gopkg.in/inconshreveable/log15.v2"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"github.com/pkg/errors"
)

var sourcegraphOrgWebhookURL = env.Get("SLACK_COMMENTS_BOT_HOOK", "", "Webhook for dogfooding notifications from an organization-level Slack bot.")

// Client is capable of posting a message to a Slack webhook
type Client struct {
	webhookURL            *string
	alsoSendToSourcegraph bool
}

// New creates a new Slack client
func New(webhookURL *string, alsoSendToSourcegraph bool) *Client {
	return &Client{webhookURL: webhookURL, alsoSendToSourcegraph: alsoSendToSourcegraph}
}

// User is an interface for accessing a Sourcegraph user's profile data
type User interface {
	Email() string
	Username() *string
	DisplayName() *string
	AvatarURL() *string
}

// Payload is the wrapper for a Slack message, defined at:
// https://api.slack.com/docs/message-formatting
type Payload struct {
	Attachments []*Attachment `json:"attachments,omitempty"`
}

// Attachment is a Slack message attachment, defined at:
// https://api.slack.com/docs/message-attachments
type Attachment struct {
	AuthorIcon string   `json:"author_icon,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	Color      string   `json:"color"`
	Fallback   string   `json:"fallback"`
	Fields     []*Field `json:"fields"`
	Footer     string   `json:"footer"`
	MarkdownIn []string `json:"mrkdwn_in"`
	ThumbURL   string   `json:"thumb_url"`
	Text       string   `json:"text,omitempty"`
	Timestamp  int64    `json:"ts"`
	Title      string   `json:"title"`
	TitleLink  string   `json:"title_link,omitempty"`
}

// Field is a single item within an attachment, defined at:
// https://api.slack.com/docs/message-attachments
type Field struct {
	Short bool   `json:"short"`
	Title string `json:"title"`
	Value string `json:"value"`
}

// Post sends payload to a Slack channel defined by the client's webhookURL
func (c *Client) Post(payload *Payload) error {
	if c.alsoSendToSourcegraph && sourcegraphOrgWebhookURL == "" {
		return errors.New("slack: env var SLACK_COMMENTS_BOT_HOOK not set")
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "slack: marshal json")
	}

	var errs []error
	if c.webhookURL != nil && *c.webhookURL != "" {
		if err := c.post(payloadJSON, *c.webhookURL); err != nil {
			errs = append(errs, err)
		}
	}
	if sourcegraphOrgWebhookURL != "" {
		if err := c.post(payloadJSON, sourcegraphOrgWebhookURL); err != nil {
			errs = append(errs, err)
		}
	}
	if errs != nil {
		return fmt.Errorf("%q", errs)
	}

	return nil
}

func (c *Client) post(payloadJSON []byte, webhookURL string) error {
	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(payloadJSON))
	if err != nil {
		return errors.Wrap(err, "slack: create post request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "slack: http request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("slack: %s failed with %d %s", payloadJSON, resp.StatusCode, string(body))
	}
	return nil
}

// NotifyOnComment posts a message to the defined Slack channel
// when a user posts a reply to a thread
func (c *Client) NotifyOnComment(
	user User,
	org *sourcegraph.Org,
	orgRepo *sourcegraph.OrgRepo,
	thread *sourcegraph.Thread,
	comment *sourcegraph.Comment,
	recipients []string,
	deepURL string,
	threadTitle string,
) {
	err := c.notifyOnComments(false, user, org, orgRepo, thread, comment, recipients, deepURL, threadTitle)
	if err != nil {
		log15.Error("slack.NotifyOnComment failed", "error", err)
	}
}

// NotifyOnThread posts a message to the defined Slack channel
// when a user creates a thread
func (c *Client) NotifyOnThread(
	user User,
	org *sourcegraph.Org,
	orgRepo *sourcegraph.OrgRepo,
	thread *sourcegraph.Thread,
	comment *sourcegraph.Comment,
	recipients []string,
	deepURL string,
) {
	err := c.notifyOnComments(true, user, org, orgRepo, thread, comment, recipients, deepURL, "")
	if err != nil {
		log15.Error("slack.NotifyOnThread failed", "error", err)
	}
}

func (c *Client) notifyOnComments(
	isNewThread bool,
	user User,
	org *sourcegraph.Org,
	orgRepo *sourcegraph.OrgRepo,
	thread *sourcegraph.Thread,
	comment *sourcegraph.Comment,
	recipients []string,
	deepURL string,
	threadTitle string,
) error {
	color := "good"
	actionText := "created a thread"
	if !isNewThread {
		color = "warning"
		// TODO: remove this check if webhook URLs are stored for every org, rather
		// than just the one constant for Sourcegraph
		if org.Name == "Sourcegraph" {
			if len(threadTitle) > 75 {
				threadTitle = threadTitle[0:75] + "..."
			}
			actionText = fmt.Sprintf("replied to a thread: \"%s\"", threadTitle)
		} else {
			actionText = fmt.Sprintf("replied to a thread")
		}
	}
	// TODO: remove this check if webhook URLs are stored for every org, rather
	// than just the one constant for Sourcegraph
	text := "_only Sourcegraph org comments visible_"
	if org.Name == "Sourcegraph" {
		text = comment.Contents
	}

	displayNameText := user.Email()
	if user.DisplayName() != nil {
		displayNameText = *user.DisplayName()
	}
	usernameText := ""
	if user.Username() != nil {
		usernameText = fmt.Sprintf("(@%s) ", *user.Username())
	}
	payload := &Payload{
		Attachments: []*Attachment{
			&Attachment{
				AuthorName: fmt.Sprintf("%s %s%s", displayNameText, usernameText, actionText),
				AuthorLink: deepURL,
				Fallback:   fmt.Sprintf("%s %s<%s|%s>!", displayNameText, usernameText, deepURL, actionText),
				Color:      color,
				Fields: []*Field{
					&Field{
						Title: "Path",
						Value: fmt.Sprintf("<%s|%s/%s (lines %d–%d)>",
							deepURL,
							orgRepo.RemoteURI,
							thread.File,
							thread.StartLine,
							thread.EndLine),
						Short: true,
					},
					&Field{
						Title: "# org members notified",
						Value: strconv.Itoa(len(recipients)),
						Short: true,
					},
				},
				Text:       text,
				MarkdownIn: []string{"text"},
			},
		},
	}

	if user.AvatarURL() != nil {
		payload.Attachments[0].ThumbURL = *user.AvatarURL()
		payload.Attachments[0].AuthorIcon = *user.AvatarURL()
	}

	return c.Post(payload)
}

// NotifyOnInvite posts a message to the defined Slack channel
// when a user invites another user to join their org
func (c *Client) NotifyOnInvite(user User, org *sourcegraph.Org, inviteEmail string) {
	displayNameText := user.Email()
	if user.DisplayName() != nil {
		displayNameText = *user.DisplayName()
	}
	usernameText := ""
	if user.Username() != nil {
		usernameText = fmt.Sprintf("(@%s) ", *user.Username())
	}

	text := fmt.Sprintf("*%s* %sjust invited %s to join *<https://sourcegraph.com/settings/teams/%s|%s>*", displayNameText, usernameText, inviteEmail, org.Name, org.Name)

	payload := &Payload{
		Attachments: []*Attachment{
			&Attachment{
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

	err := c.Post(payload)
	if err != nil {
		log15.Error("slack.NotifyOnInvite failed", "error", err)
	}
}

// NotifyOnAcceptedInvite posts a message to the defined Slack channel
// when an invited user accepts their invite to join an org
func (c *Client) NotifyOnAcceptedInvite(user User, org *sourcegraph.Org) {
	displayNameText := user.Email()
	if user.DisplayName() != nil {
		displayNameText = *user.DisplayName()
	}
	usernameText := ""
	if user.Username() != nil {
		usernameText = fmt.Sprintf("(@%s) ", *user.Username())
	}

	text := fmt.Sprintf("*%s* %sjust accepted their invitation to join *<https://sourcegraph.com/settings/teams/%s|%s>*", displayNameText, usernameText, org.Name, org.Name)

	payload := &Payload{
		Attachments: []*Attachment{
			&Attachment{
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

	err := c.Post(payload)
	if err != nil {
		log15.Error("slack.NotifyOnAcceptedInvite failed", "error", err)
	}
}
