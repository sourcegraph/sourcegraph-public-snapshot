// Package slack provides Slack integration.
package slack

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

var (
	// WebhookURL is the URL to the Slack webhook endpoint for posting
	// Slack notifications. If empty, Slack notifications are
	// disabled.
	WebhookURL = os.Getenv("SG_SLACK_WEBHOOK_URL")

	defaultChannel  = "#dev-bot"
	defaultUsername = "Sourcegraph"
)

// Enabled is a boolean indicating whether PostMessage should actually
// post a message. If false, it is a no-op.
func Enabled() bool { return WebhookURL != "" }

// PostOpts rerepsents options for posting a message to slack.
type PostOpts struct {
	Msg              string
	Username         string // If empty, defaultUsername is used instead.
	IconEmoji        string
	IconURL          string
	Channel          string // If empty, defaultChannel is used instead.
	DisableLinkNames bool   // If true, "@ mentions" won't notify users.
	Attachments      []string
}

// PostMessage posts a message to the Slack channel. If an error
// occurs, it is logged and PostMessage returns (without panicking).
func PostMessage(opt PostOpts) {
	if !Enabled() {
		log.Printf("Ignored Slack message: %q.", opt.Msg)
		return
	}

	if opt.IconEmoji == "" && opt.IconURL == "" {
		opt.IconURL = "https://sourcegraph.com/static/img/favicon.png"
	}

	if opt.Channel == "" {
		opt.Channel = defaultChannel
	}
	if opt.Username == "" {
		opt.Username = defaultUsername
	}

	type attachmentField struct {
		Title string `json:"title"`
		Value string `json:"value"`
		Short bool   `json:"short"`
	}
	type attachment struct {
		Fallback string            `json:"fallback"`
		Pretext  string            `json:"pretext"`
		Color    string            `json:"color"`
		Fields   []attachmentField `json:"fields"`
	}

	var o = struct {
		Channel     string       `json:"channel"`
		Username    string       `json:"username"`
		Text        string       `json:"text"`
		IconEmoji   string       `json:"icon_emoji,omitempty"`
		IconURL     string       `json:"icon_url,omitempty"`
		Attachments []attachment `json:"attachments,omitempty"`
		LinkNames   int          `json:"link_names,omitempty"`
	}{
		Channel:   opt.Channel,
		Username:  opt.Username,
		Text:      opt.Msg,
		IconEmoji: opt.IconEmoji,
		IconURL:   opt.IconURL,
		LinkNames: map[bool]int{true: 0, false: 1}[opt.DisableLinkNames],
	}

	// TODO(sqs): attachments

	postMessage := func() error {
		b, err := json.Marshal(o)
		if err != nil {
			return err
		}
		resp, err := http.Post(WebhookURL, "application/json", bytes.NewReader(b))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	}

	if err := postMessage(); err != nil {
		log.Printf("WARNING: failed to post Slack message %+v: %s.", o, err)
	}
}
