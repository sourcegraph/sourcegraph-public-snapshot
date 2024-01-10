package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SlackClient is a client for interfacing with the Slack Webhook API
type SlackClient struct {
	WebhookURL string

	HttpClient *http.Client
}

// NewSlackClient configures a SlackClient for a given webhook URL
func NewSlackClient(url string) *SlackClient {
	hc := http.Client{}

	c := SlackClient{
		WebhookURL: url,
		HttpClient: &hc,
	}

	return &c
}

// PostMessage posts a bytes.Buffer to the given Slack webhook URL with markdown enabled
func (s *SlackClient) PostMessage(b bytes.Buffer) error {

	type slackRequest struct {
		Text     string `json:"text"`
		Markdown bool   `json:"mrkdwn"`
	}

	payload, err := json.Marshal(slackRequest{Text: b.String(), Markdown: true})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Println(string(body))
		return errors.Newf("received non-200 status code %v: %s", resp.StatusCode, err.Error())
	}

	return nil
}

// TemplateData represents all the data required to correctly render the template
type TemplateData struct {
	VersionAge string

	Version     string
	Environment string

	CommitTooOld bool
	Threshold    string
	Drift        string

	InAllowedCommits bool
	NumCommits       int
}

// createMessage renders a template and returns teh result as a bytes.Buffer to either
// be printed or posted to Slack
func createMessage(td TemplateData) (bytes.Buffer, error) {
	var msg bytes.Buffer

	var slackTemplate = `:warning: *{{.Environment}}*'s version may be out of date.
Current version: ` + "`{{ .Version }}`" + ` was committed *{{ .VersionAge }} hours ago*.

*Alerts*:
{{- if not .InAllowedCommits}}
• ` + "`{{.Version}}`" + ` was not found in the last ` + "`{{.NumCommits}}`" + ` commits.
{{- end}}
{{- if .CommitTooOld}}
• ` + "`{{.Version}}`" + ` is ` + "`{{.Drift}}`" + ` older than the tip of ` + "`main`" + `which exceeds the threshold of ` + "`{{.Threshold}}`" + `
{{- end}}

Check <https://github.com/sourcegraph/deploy-sourcegraph-cloud/pulls|deploy-sourcegraph-cloud> to see if a release is blocked.

cc <!subteam^S02J9TTQLBU|dev-experience-support>`

	tpl, err := template.New("slack-message").Parse(slackTemplate)
	if err != nil {
		return msg, err
	}

	tw := tabwriter.NewWriter(&msg, 0, 8, 1, '\t', 0)

	err = tpl.Execute(tw, td)
	if err != nil {
		return msg, err
	}

	tw.Flush()

	return msg, nil
}
