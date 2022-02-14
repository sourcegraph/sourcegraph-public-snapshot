package background

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

	var resultContentBlocks []slack.Block
	if args.IncludeResults {
		truncatedResults, truncatedCount := truncateResults(args.Results, 5)
		for _, result := range truncatedResults {
			resultContentBlocks = append(resultContentBlocks, newMarkdownSection(fmt.Sprintf(
				"%s@%s",
				result.Commit.Repository.Name,
				result.Commit.Oid[:8],
			)))
			var contentRaw string
			if result.DiffPreview != nil {
				contentRaw = result.DiffPreview.Value
			} else {
				contentRaw = result.MessagePreview.Value
			}
			resultContentBlocks = append(resultContentBlocks, newMarkdownSection(fmt.Sprintf("```%s```", contentRaw)))
		}
		if truncatedCount > 0 {
			resultContentBlocks = append(resultContentBlocks, newMarkdownSection(fmt.Sprintf("...and %d more matches.", truncatedCount)))
		}
	}

	// To see what this looks like:
	// https://app.slack.com/block-kit-builder/T02FSM7DL#%7B%22blocks%22:%5B%7B%22type%22:%22section%22,%22text%22:%7B%22text%22:%22*New%20results%20for%20code%20monitor*%22,%22type%22:%22mrkdwn%22%7D%7D,%7B%22type%22:%22section%22,%22text%22:%7B%22type%22:%22mrkdwn%22,%22text%22:%225%20new%20results%20for%20query%20%60%60%60type:diff%20repo:github.com/sourcegraph/sourcegraph$%20BEGIN%20PRIVATE%20KEY%60%60%60%22%7D%7D,%7B%22type%22:%22section%22,%22text%22:%7B%22type%22:%22mrkdwn%22,%22text%22:%22%3Chttps://sourcegraph.com/search?q=context:global+type:diff+repo:github.com/sourcegraph/sourcegraph%2524+BEGIN+PRIVATE+KEY&patternType=literal%7CView%20Search%20Results%3E%20%7C%20%3Chttps://sourcegraph.com/code-monitoring?visible=1%7CView%20Code%20Monitor%3E%22%7D%7D%5D%7D
	blocks := []slack.Block{
		newMarkdownSection(fmt.Sprintf("*New results for Code Monitor \"%s\"*", args.MonitorDescription)),
		newMarkdownSection(fmt.Sprintf("%d new results for query: `%s`", len(args.Results), args.Query)),
	}
	blocks = append(blocks, resultContentBlocks...)
	blocks = append(
		blocks,
		newMarkdownSection(fmt.Sprintf(`<%s|View search on Sourcegraph> | <%s|View code monitor>`, args.QueryURL, args.MonitorURL)),
	)
	return &slack.WebhookMessage{Blocks: &slack.Blocks{BlockSet: blocks}}
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
