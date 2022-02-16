package background

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/slack-go/slack"

	cmtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/types"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func sendSlackNotification(ctx context.Context, url string, args actionArgs) error {
	return postSlackWebhook(ctx, httpcli.ExternalDoer, url, slackPayload(args))
}

func slackPayload(args actionArgs) *slack.WebhookMessage {
	newMarkdownSection := func(s string) slack.Block {
		return slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", s, false, false), nil, nil)
	}

	blocks := []slack.Block{
		newMarkdownSection(fmt.Sprintf(
			"%s's Sourcegraph Code monitor, *%s*, detected *%d* new matches.",
			args.MonitorOwnerName,
			args.MonitorDescription,
			len(args.Results),
		)),
	}

	if args.IncludeResults {
		truncatedResults, truncatedCount := truncateResults(args.Results, 5)
		for _, result := range truncatedResults {
			resultType := "Message"
			if result.DiffPreview != nil {
				resultType = "Diff"
			}
			blocks = append(blocks, newMarkdownSection(fmt.Sprintf(
				"%s match: <%s|%s@%s>",
				resultType,
				getCommitURL(args.ExternalURL, result.Commit.Repository.Name, result.Commit.Oid, args.UTMSource),
				result.Commit.Repository.Name,
				result.Commit.Oid[:8],
			)))
			var contentRaw string
			if result.DiffPreview != nil {
				contentRaw = truncateString(result.DiffPreview.Value, 10)
			} else {
				contentRaw = truncateString(result.MessagePreview.Value, 10)
			}
			blocks = append(blocks, newMarkdownSection(fmt.Sprintf("```%s```", contentRaw)))
		}
		if truncatedCount > 0 {
			blocks = append(blocks, newMarkdownSection(fmt.Sprintf(
				"...and <%s|%d more matches>.",
				getSearchURL(args.ExternalURL, args.Query, args.UTMSource),
				truncatedCount,
			)))
		}
	} else {
		blocks = append(blocks, newMarkdownSection(fmt.Sprintf(
			"<%s|View results>",
			getSearchURL(args.ExternalURL, args.Query, args.UTMSource),
		)))
	}

	blocks = append(blocks,
		newMarkdownSection(fmt.Sprintf(
			`If you are %s, you can <%s|edit your code monitor>`,
			args.MonitorOwnerName,
			getCodeMonitorURL(args.ExternalURL, args.MonitorID, args.UTMSource),
		)),
	)
	return &slack.WebhookMessage{Blocks: &slack.Blocks{BlockSet: blocks}}
}

func truncateString(input string, lines int) string {
	splitLines := strings.SplitAfter(input, "\n")
	if len(splitLines) > lines {
		splitLines = splitLines[:lines]
		splitLines = append(splitLines, "...\n")
	}
	return strings.Join(splitLines, "")
}

func truncateResults(results cmtypes.CommitSearchResults, maxResults int) (cmtypes.CommitSearchResults, int) {
	remaining := maxResults
	var output cmtypes.CommitSearchResults
	for _, result := range results {
		var highlights []cmtypes.Highlight
		if result.DiffPreview != nil { // diff match
			highlights = result.DiffPreview.Highlights
		} else { // commit message match
			highlights = result.MessagePreview.Highlights
		}

		if len(highlights) < remaining {
			remaining -= len(highlights)
			output = append(output, result)
			continue
		}

		if len(highlights) == remaining {
			output = append(output, result)
			break
		}

		highlights = highlights[:remaining]
		if result.DiffPreview != nil {
			result.DiffPreview.Highlights = highlights
		} else {
			result.MessagePreview.Highlights = highlights
		}

		output = append(output, result)
		break
	}

	outputCount := 0
	for _, result := range output {
		if result.DiffPreview != nil {
			outputCount += len(result.DiffPreview.Highlights)
		} else {
			outputCount += len(result.MessagePreview.Highlights)
		}
	}

	totalCount := 0
	for _, result := range results {
		if result.DiffPreview != nil {
			totalCount += len(result.DiffPreview.Highlights)
		} else {
			totalCount += len(result.MessagePreview.Highlights)
		}
	}

	return output, totalCount - outputCount
}

// adapted from slack.PostWebhookCustomHTTPContext
func postSlackWebhook(ctx context.Context, doer httpcli.Doer, url string, msg *slack.WebhookMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "marshal failed")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return errors.Wrap(err, "failed new request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := doer.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to post webhook")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return StatusCodeError{
			Code:   resp.StatusCode,
			Status: resp.Status,
			Body:   string(body),
		}
	}

	return nil
}

func SendTestSlackWebhook(ctx context.Context, doer httpcli.Doer, description, url string) error {
	testMessage := &slack.WebhookMessage{Blocks: &slack.Blocks{BlockSet: []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf(
					"Test message for Code Monitor '%s'",
					description,
				),
				false,
				false,
			),
			nil,
			nil,
		),
	}}}

	return postSlackWebhook(ctx, doer, url, testMessage)
}
