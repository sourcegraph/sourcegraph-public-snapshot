package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
)

func slackMention(slackUserID string) string {
	return fmt.Sprintf("<@%s>", slackUserID)
}

func slackSummary(locked bool, failedCommits []CommitInfo) string {
	if !locked {
		return ":white_check_mark: Pipeline healthy - branch unlocked!"
	}
	message := `:alert: *Consecutive build failures detected - branch has been locked.* :alert:
The authors of the following failed commits who are Sourcegraph teammates have been granted merge access to investigate and resolve the issue:
`

	for _, commit := range failedCommits {
		var mention string
		if commit.AuthorSlackID != "" {
			mention = slackMention(commit.AuthorSlackID)
		} else {
			mention = commit.Author
		}
		message += fmt.Sprintf("\n- <https://github.com/sourcegraph/sourcegraph/commit/%s|%.7s>: %s",
			commit.Commit, commit.Commit, mention)
	}
	message += `

The branch will automatically be unlocked once a green build is run.
Refer to the <https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci|CI incident playbook> for help.
If unable to resolve the issue, please start an incident with the '/incident' Slack command.

cc: @dev-experience-support`
	return message
}

// postSlackUpdate attempts to send the given summary to at each of the provided webhooks.
func postSlackUpdate(webhooks []string, summary string) (bool, error) {
	if len(webhooks) == 0 {
		return false, nil
	}

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
		return false, fmt.Errorf("MarshalIndent: %w", err)
	}
	log.Println("slackBody: ", string(body))

	// Attempt to send a message out to each
	var errs error
	var oneSucceeded bool
	for _, webhook := range webhooks {
		if len(webhook) == 0 {
			return false, nil
		}

		log.Println("posting to ", webhook)

		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(body))
		if err != nil {
			errs = errors.CombineErrors(errs, fmt.Errorf("%s: NewRequest: %w", webhook, err))
			continue
		}
		req.Header.Add("Content-Type", "application/json")

		// Perform the HTTP Post on the webhook
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			errs = errors.CombineErrors(errs, fmt.Errorf("%s: client.Do: %w", webhook, err))
			continue
		}

		// Parse the response, to check if it succeeded
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			errs = errors.CombineErrors(errs, fmt.Errorf("%s: buf.ReadFrom(resp.Body): %w", webhook, err))
			continue
		}
		defer resp.Body.Close()
		if buf.String() != "ok" {
			errs = errors.CombineErrors(errs, fmt.Errorf("%s: non-ok response from Slack: %s", webhook, buf.String()))
			continue
		}

		// Indicate at least one message succeeded
		oneSucceeded = true
	}

	return oneSucceeded, err
}
