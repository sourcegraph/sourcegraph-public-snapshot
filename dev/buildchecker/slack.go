package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func slackMention(slackUserID string) string {
	return fmt.Sprintf("<@%s>", slackUserID)
}

func generateBranchEventSummary(locked bool, branch string, discussionChannel string, failedCommits []CommitInfo) string {
	branchStr := fmt.Sprintf("`%s`", branch)
	if !locked {
		return fmt.Sprintf(":white_check_mark: Pipeline healthy - %s unlocked!", branchStr)
	}
	message := fmt.Sprintf(`:alert: *Consecutive build failures detected - the %s branch has been locked.* :alert:
The authors of the following failed commits who are Sourcegraph teammates have been granted merge access to investigate and resolve the issue:
`, branchStr)

	// Reverse order of commits so that the oldest are listed first
	sort.Slice(failedCommits, func(i, j int) bool { return failedCommits[i].BuildCreated.After(failedCommits[j].BuildCreated) })

	for _, commit := range failedCommits {
		var mention string
		if commit.AuthorSlackID != "" {
			mention = slackMention(commit.AuthorSlackID)
		} else if commit.Author != "" {
			mention = commit.Author
		} else {
			mention = "unable to infer author"
		}

		message += fmt.Sprintf("\n- <https://github.com/sourcegraph/sourcegraph/commit/%s|%.7s> (<%s|build %d>): %s",
			commit.Commit, commit.Commit, commit.BuildURL, commit.BuildNumber, mention)
	}
	message += fmt.Sprintf(`

The branch will automatically be unlocked once a green build has run on %s.
Please head over to %s for relevant discussion about this branch lock.
:bulb: First time being mentioned by this bot? :point_right: <https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci/#build-has-failed-on-the-main-branch|Follow this step by step guide!>.

For more, refer to the <https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci|CI incident playbook> for help.

If unable to resolve the issue, please start an incident with the '/incident' Slack command.`, branchStr, discussionChannel)
	return message
}

// postSlackUpdate attempts to send the given summary to at each of the provided webhooks.
func postSlackUpdate(webhooks []string, summary string) (bool, error) {
	log.Printf("postSlackUpdate. len(webhooks)=%d\n", len(webhooks))
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
		return false, errors.Newf("MarshalIndent: %w", err)
	}
	log.Println("slackBody: ", string(body))

	// Attempt to send a message out to each
	var oneSucceeded bool
	var allErrs error
	for i, webhook := range webhooks {
		if len(webhook) == 0 {
			return false, nil
		}

		log.Println("posting to webhook ", i)

		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(body))
		if err != nil {
			allErrs = errors.CombineErrors(allErrs, errors.Newf("%s: NewRequest: %w", webhook, err))
			continue
		}
		req.Header.Add("Content-Type", "application/json")

		// Perform the HTTP Post on the webhook
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			allErrs = errors.CombineErrors(allErrs, errors.Newf("%s: client.Do: %w", webhook, err))
			continue
		}

		// Parse the response, to check if it succeeded
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			allErrs = errors.CombineErrors(allErrs, errors.Newf("%s: buf.ReadFrom(resp.Body): %w", webhook, err))
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			allErrs = errors.CombineErrors(allErrs, errors.Newf("%s: Status code %d response from Slack: %s", webhook, resp.StatusCode, buf.String()))
			continue
		}

		// Indicate at least one message succeeded
		oneSucceeded = true
	}

	return oneSucceeded, err
}
