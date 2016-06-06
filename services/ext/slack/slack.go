// Package slack provides Slack integration.
package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"
)

// PostOpts rerepsents options for posting a message to slack.
type PostOpts struct {
	Msg              string
	Username         string // If empty, defaultUsername is used instead.
	IconEmoji        string
	IconURL          string
	Channel          string // If empty, defaultChannel is used instead.
	DisableLinkNames bool   // If true, "@ mentions" won't notify users.
	Attachments      []string
	WebhookURL       string
}

// PostMessage posts a message to the Slack channel. If an error
// occurs, it is logged and PostMessage returns (without panicking).
func PostMessage(opt PostOpts) {
	var webhookURL string
	if opt.WebhookURL != "" {
		webhookURL = opt.WebhookURL
	} else {
		webhookURL = Config.GetWebhookURLIfConfigured()
	}
	if webhookURL == "" {
		log15.Debug("Ignored Slack message", "msg", opt.Msg)
		return
	}

	if opt.IconEmoji == "" && opt.IconURL == "" {
		opt.IconURL = Config.DefaultIcon
	}
	if opt.Channel == "" {
		opt.Channel = Config.DefaultChannel
	}
	if opt.Channel != "" && !strings.HasPrefix(opt.Channel, "#") {
		opt.Channel = "#" + opt.Channel
	}
	if opt.Username == "" {
		opt.Username = Config.DefaultUsername
	}
	opt.DisableLinkNames = opt.DisableLinkNames || Config.DisableLinkNames

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
		Channel     string       `json:"channel,omitempty"`
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
		resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(b))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	}

	go func() {
		if err := postMessage(); err != nil {
			log15.Warn("Failed to post Slack message", "payload", o, "error", err)
		}
	}()
}
