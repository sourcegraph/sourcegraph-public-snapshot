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

// SignupsWebhookURL and CommentsWebhookURL are the Slack endpoints
// that receive signup messages and publish them to the appropriate channels
// https://sourcegraph.slack.com/services/
var SignupsWebhookURL = env.Get("SLACK_SIGNUPS_BOT_HOOK", "", "Webhook for posting signup notifications to the Slack #bot-signups channel.")
var CommentsWebhookURL = env.Get("SLACK_COMMENTS_BOT_HOOK", "", "Webhook for posting comment notifications to the Slack #comments channel.")

type payload struct {
	Attachments []*attachment `json:"attachments"`
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
	if SignupsWebhookURL == "" {
		return errors.New("Slack Webhook URL not defined")
	}

	color := "good"

	links := ""
	if response != nil {
		links = fmt.Sprintf("<%s|View on Looker>, <%s|View on HubSpot>", hubSpotProps.LookerLink, fmt.Sprintf("https://app.hubspot.com/contacts/2762526/contact/%v", response.VID))
	} else {
		links = fmt.Sprintf("<%s|View on Looker>", hubSpotProps.LookerLink)
	}

	payload := &payload{
		Attachments: []*attachment{
			&attachment{
				Fallback: fmt.Sprintf("%s just signed up!", actor.Email),
				Title:    fmt.Sprintf("%s just signed up!", actor.Email),
				Color:    color,
				ThumbURL: actor.AvatarURL,
				Fields: []*field{
					&field{
						Title: "User Email",
						Value: actor.Email,
						Short: true,
					},
					&field{
						Title: "User profile links",
						Value: links,
						Short: false,
					},
				},
			},
		},
	}

	return post(payload, SignupsWebhookURL)
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
	if CommentsWebhookURL == "" {
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
						Short: false,
					},
				},
			},
		},
	}
	if strings.HasPrefix(repoRemoteURI, "github.com/sourcegraph") {
		payload.Attachments[0].Fields = append(payload.Attachments[0].Fields, &field{
			Title: "Deep link (SG only)",
			Value: commentURL,
			Short: false,
		})
	}

	return post(payload, CommentsWebhookURL)
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
