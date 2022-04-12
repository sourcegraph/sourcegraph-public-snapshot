package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("Response body: ", string(body))
		return errors.Newf("received non-200 status code %v", resp.StatusCode)
	}

	return nil
}

// createMessage renders a template and returns teh result as a bytes.Buffer to either
// be printed or posted to Slack
func createMessage(version, environment string, current Commit, numCommits int) (bytes.Buffer, error) {
	var msg bytes.Buffer

	drift := time.Now().Sub(current.Date).Hours()

	var slackTemplate = `:warning: *{{.Environment}}*'s version was not found in the last {{.NumCommits}} commits.

Current version: ` + "`{{ .Version }}`" + ` was committed *{{.Drift}} hours ago*. 

Check <https://github.com/sourcegraph/deploy-sourcegraph-cloud/pulls|deploy-sourcegraph-cloud> to see if a release is blocked.

cc <!subteam^S02NFV6A536|devops-support>
`

	type templateData struct {
		Version     string
		Environment string
		Drift       string
		NumCommits  int
	}

	td := templateData{Version: version, Environment: environment, Drift: fmt.Sprintf("%.2f", drift), NumCommits: numCommits}

	tpl, err := template.New("slack-message").Parse(slackTemplate)
	if err != nil {
		return msg, err
	}

	// tw := tabwriter.NewWriter(&msg, 0, 8, 2, '\t', 0)
	tw := tabwriter.NewWriter(&msg, 0, 8, 1, '\t', 0)

	err = tpl.Execute(tw, td)
	if err != nil {
		return msg, err
	}

	tw.Flush()

	return msg, nil
}
