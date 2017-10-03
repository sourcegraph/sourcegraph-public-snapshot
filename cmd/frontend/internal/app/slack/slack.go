package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"

	"github.com/pkg/errors"
)

type Client struct {
	webhookURL string
}

func New(webhookURL string) *Client {
	return &Client{webhookURL: webhookURL}
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
	if c.webhookURL == "" {
		return errors.New("slack: webhookURL is empty")
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "slack: marshal json")
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewReader(payloadJSON))
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
	actor *actor.Actor,
	org *sourcegraph.Org,
	orgRepo *sourcegraph.OrgRepo,
	thread *sourcegraph.Thread,
	comment *sourcegraph.Comment,
	recipients []string,
	deepURL string,
	threadTitle string,
) error {
	return c.notifyOnComments(false, actor, org, orgRepo, thread, comment, recipients, deepURL, threadTitle)
}

// NotifyOnThread posts a message to the defined Slack channel
// when a user creates a thread
func (c *Client) NotifyOnThread(
	actor *actor.Actor,
	org *sourcegraph.Org,
	orgRepo *sourcegraph.OrgRepo,
	thread *sourcegraph.Thread,
	comment *sourcegraph.Comment,
	recipients []string,
	deepURL string,
) error {
	return c.notifyOnComments(true, actor, org, orgRepo, thread, comment, recipients, deepURL, "")
}

func (c *Client) notifyOnComments(
	isNewThread bool,
	actor *actor.Actor,
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
		// TODO: remove this check if webhook URLs are stored for every org, rather
		// than just the one constant for Sourcegraph
		if org.Name == "Sourcegraph" {
			color = "warning"
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

	payload := &Payload{
		Attachments: []*Attachment{
			&Attachment{
				ThumbURL:   actor.AvatarURL,
				AuthorIcon: actor.AvatarURL,
				AuthorName: fmt.Sprintf("%s %s", actor.Login, actionText),
				AuthorLink: deepURL,
				Fallback:   fmt.Sprintf("%s (%s) just <%s|%s>!", actor.Login, actor.Email, deepURL, actionText),
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
						Title: "Thread participants",
						Value: strings.Join(recipients, ", "),
						Short: true,
					},
				},
				Text:       text,
				MarkdownIn: []string{"text"},
			},
		},
	}

	return c.Post(payload)
}

// NotifyOnInvite posts a message to the defined Slack channel
// when a user invites another user to join their org
func (c *Client) NotifyOnInvite(actor *actor.Actor, org *sourcegraph.Org, email string) error {
	text := fmt.Sprintf("*%s* (%s) just invited %s to join *<https://sourcegraph.com/settings/teams/%s|%s>*", actor.Login, actor.Email, email, org.Name, org.Name)

	payload := &Payload{
		Attachments: []*Attachment{
			&Attachment{
				ThumbURL:   actor.AvatarURL,
				Fallback:   text,
				Color:      "#F96316",
				Text:       text,
				MarkdownIn: []string{"text"},
			},
		},
	}

	return c.Post(payload)
}

// NotifyOnAcceptedInvite posts a message to the defined Slack channel
// when an invited user accepts their invite to join an org
func (c *Client) NotifyOnAcceptedInvite(actor *actor.Actor, org *sourcegraph.Org) error {
	text := fmt.Sprintf("*%s* (%s) just accepted their invitation to join *<https://sourcegraph.com/settings/teams/%s|%s>*", actor.Login, actor.Email, org.Name, org.Name)

	payload := &Payload{
		Attachments: []*Attachment{
			&Attachment{
				ThumbURL:   actor.AvatarURL,
				Fallback:   text,
				Color:      "#B114F7",
				Text:       text,
				MarkdownIn: []string{"text"},
			},
		},
	}

	return c.Post(payload)
}
