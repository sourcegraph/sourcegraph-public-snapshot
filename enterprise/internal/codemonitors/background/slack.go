package background

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/slack-go/slack"

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

	// To see what this looks like:
	// https://app.slack.com/block-kit-builder/T02FSM7DL#%7B%22blocks%22:%5B%7B%22type%22:%22section%22,%22text%22:%7B%22text%22:%22*New%20results%20for%20code%20monitor*%22,%22type%22:%22mrkdwn%22%7D%7D,%7B%22type%22:%22section%22,%22text%22:%7B%22type%22:%22mrkdwn%22,%22text%22:%225%20new%20results%20for%20query%20%60%60%60type:diff%20repo:github.com/sourcegraph/sourcegraph$%20BEGIN%20PRIVATE%20KEY%60%60%60%22%7D%7D,%7B%22type%22:%22section%22,%22text%22:%7B%22type%22:%22mrkdwn%22,%22text%22:%22%3Chttps://sourcegraph.com/search?q=context:global+type:diff+repo:github.com/sourcegraph/sourcegraph%2524+BEGIN+PRIVATE+KEY&patternType=literal%7CView%20Search%20Results%3E%20%7C%20%3Chttps://sourcegraph.com/code-monitoring?visible=1%7CView%20Code%20Monitor%3E%22%7D%7D%5D%7D
	return &slack.WebhookMessage{
		Blocks: &slack.Blocks{BlockSet: []slack.Block{
			newMarkdownSection(fmt.Sprintf("*New results for Code Monitor \"%s\"*", args.MonitorDescription)),
			newMarkdownSection(fmt.Sprintf("%d new results for query: ```%s```", args.NumResults, args.Query)),
			newMarkdownSection(fmt.Sprintf(`<%s|View search on Sourcegraph> | <%s|View code monitor>`, args.QueryURL, args.MonitorURL)),
		}},
	}
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
