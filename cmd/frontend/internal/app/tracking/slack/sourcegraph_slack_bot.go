package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
)

var signupsWebhookURL = env.Get("SLACK_SIGNUPS_BOT_HOOK", "", "Webhook for posting signup notifications to the Slack #bot-signups channel.")
var commentsWebhookURL = env.Get("SLACK_COMMENTS_BOT_HOOK", "", "Webhook for posting comment notifications to the Slack #comments channel.")

type payload struct {
	Attachments []*attachment `json:"attachments,omitempty"`
}
type attachment struct {
	Fallback  string   `json:"fallback"`
	Color     string   `json:"color"`
	Title     string   `json:"title"`
	ThumbURL  string   `json:"thumb_url"`
	Fields    []*field `json:"fields"`
	Timestamp int64    `json:"ts"`
}
type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NotifyOnSignup posts a message to the Slack channel #bot-signups
// when a user signs up for Sourcegraph
func NotifyOnSignup(actor *actor.Actor, hubSpotProps *hubspot.ContactProperties, response *hubspot.ContactResponse) error {
	if signupsWebhookURL == "" {
		return errors.New("Slack Webhook URL not defined")
	}

	var links []string
	if hubSpotProps.LookerLink != "" {
		links = append(links, fmt.Sprintf("<%s|View on Looker>", hubSpotProps.LookerLink))
	}
	if response != nil {
		links = append(links, fmt.Sprintf("<https://app.hubspot.com/contacts/2762526/contact/%v|View on HubSpot>", response.VID))
	}

	payload := &payload{
		Attachments: []*attachment{
			&attachment{
				Fallback: fmt.Sprintf("%s just signed up!", actor.Email),
				Title:    fmt.Sprintf("%s just signed up!", actor.Email),
				Color:    "good",
				ThumbURL: actor.AvatarURL,
				Fields: []*field{
					&field{
						Title: "User profile links",
						Value: strings.Join(links, ", "),
						Short: false,
					},
				},
			},
		},
	}

	return post(payload, signupsWebhookURL)
}

// NotifyOnComment posts a message to the Slack channel #comments
// when a user posts a reply to a thread
func NotifyOnComment(authorName string, authorEmail string, repoRemoteURI string, recipients string, commentURL string) error {
	return notifyOnComments("replied to a thread", authorName, authorEmail, repoRemoteURI, recipients, commentURL)
}

// NotifyOnThread posts a message to the Slack channel #comments
// when a user creates a thread
func NotifyOnThread(authorName string, authorEmail string, repoRemoteURI string, recipients string, commentURL string) error {
	return notifyOnComments("created a thread", authorName, authorEmail, repoRemoteURI, recipients, commentURL)
}

func notifyOnComments(actionText string, authorName string, authorEmail string, repoRemoteURI string, recipients string, commentURL string) error {
	if commentsWebhookURL == "" {
		return errors.New("Slack Webhook URL not defined")
	}

	payload := &payload{
		Attachments: []*attachment{
			&attachment{
				Fallback: fmt.Sprintf("%s just %s!", authorName, actionText),
				Title:    fmt.Sprintf("%s just %s!", authorName, actionText),
				Color:    "good",
				Fields: []*field{
					&field{
						Title: "User email",
						Value: authorEmail,
						Short: true,
					},
					&field{
						Title: "Repo",
						Value: repoRemoteURI,
						Short: true,
					},
					&field{
						Title: "All recipients",
						Value: recipients,
						Short: true,
					},
				},
			},
		},
	}
	if strings.HasPrefix(repoRemoteURI, "github.com/sourcegraph") {
		payload.Attachments[0].Fields = append(payload.Attachments[0].Fields, &field{
			Title: "Deep link (SG only)",
			Value: fmt.Sprintf("<%s|View in Sourcegraph>", commentURL),
			Short: true,
		})
	}

	return post(payload, commentsWebhookURL)
}

func post(payload *payload, webhookURL string) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "slack.post")
	}

	req, err := http.NewRequest("POST", webhookURL, strings.NewReader(string(payloadJSON)))
	if err != nil {
		return errors.Wrap(err, "slack.post")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "slack.post")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return errors.Wrap(fmt.Errorf("Code %v: %s", resp.StatusCode, buf.String()), "slack.post")
	}
	return nil
}
