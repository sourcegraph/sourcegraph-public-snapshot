package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func slackSummary(locked bool, failedCommits []CommitInfo) string {
	if !locked {
		return ":white_check_mark: Pipeline healthy - branch unlocked!"
	}
	message := `:alert: *Consecutive build failures detected - branch has been locked.* :alert:
The authors of the following failed commits who are Sourcegraph teammates have been granted merge access to investigate and resolve the issue:
`
	for _, commit := range failedCommits {
		message += fmt.Sprintf("\n- <https://github.com/sourcegraph/sourcegraph/commit/%s|%s> - %s",
			commit.Commit, commit.Commit, commit.Author)
	}
	message += `

The branch will automatically be unlocked once a green build is run.
Refer to the <https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci|CI incident playbook> for help.
If unable to resolve the issue, please start an incident with the '/incident' Slack command.

cc: @dev-experience-support`
	return message
}

func postSlackUpdate(webhook string, summary string) error {
	type slackText struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type slackBlock struct {
		Type string     `json:"type"`
		Text *slackText `json:"text,omitempty"`
	}

	// Generate request
	body, err := json.MarshalIndent(struct {
		Blocks []slackBlock `json:"blocks"`
	}{
		Blocks: []slackBlock{{
			Type: "section",
			Text: &slackText{
				Type: "mrkdwn",
				Text: summary,
			},
		}},
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to post on slack: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to post on slack: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Perform the HTTP Post on the webhook
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post on slack: %w", err)
	}

	// Parse the response, to check if it succeeded
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if buf.String() != "ok" {
		return fmt.Errorf("failed to post on slack: %s", buf.String())
	}
	return nil
}
