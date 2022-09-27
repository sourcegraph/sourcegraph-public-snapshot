package background

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"code.gitea.io/gitea/modules/hostmatcher"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func sendWebhookNotification(ctx context.Context, url string, args actionArgs) error {
	return postWebhook(ctx, url, args)
}

func postWebhook(ctx context.Context, url string, args actionArgs) error {

	// Create an allowList out of specified HostList
	allowList := hostmatcher.ParseHostMatchList("", args.HostList)

	webHookHttpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: hostmatcher.NewDialContext("code-monitor-webhook", allowList, nil),
		},
	}

	payload := generateWebhookPayload(args)
	raw, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshal failed")
	}

	resp, err := webHookHttpClient.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		return errors.Wrap(err, "failed new request")
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

func SendTestWebhook(ctx context.Context, description string, u string, hostList string) error {
	args := actionArgs{
		ExternalURL:        &url.URL{},
		MonitorDescription: description,
		Query:              "test query",
		HostList:           hostList,
	}
	return postWebhook(ctx, u, args)
}

type webhookPayload struct {
	MonitorDescription string          `json:"monitorDescription"`
	MonitorURL         string          `json:"monitorURL"`
	Query              string          `json:"query"`
	Results            []webhookResult `json:"results,omitempty"`
}

func generateWebhookPayload(args actionArgs) webhookPayload {
	p := webhookPayload{
		MonitorDescription: args.MonitorDescription,
		MonitorURL:         getCodeMonitorURL(args.ExternalURL, args.MonitorID, args.UTMSource),
		Query:              args.Query,
	}

	if args.IncludeResults {
		p.Results = generateResults(args.Results)
	}

	return p
}

type webhookResult struct {
	Repository           string   `json:"repository"`
	Commit               string   `json:"commit"`
	Message              string   `json:"message,omitempty"`
	MatchedMessageRanges [][2]int `json:"matchedMessageRanges,omitempty"`
	Diff                 string   `json:"diff,omitempty"`
	MatchedDiffRanges    [][2]int `json:"matchedDiffRanges,omitempty"`
}

func generateResults(in []*result.CommitMatch) []webhookResult {
	out := make([]webhookResult, len(in))
	for i, match := range in {
		res := webhookResult{
			Repository: string(match.Repo.Name),
			Commit:     string(match.Commit.ID),
		}
		if match.MessagePreview != nil {
			res.Message = match.MessagePreview.Content
			res.MatchedMessageRanges = rangesToInts(match.MessagePreview.MatchedRanges)
		}
		if match.DiffPreview != nil {
			res.Diff = match.DiffPreview.Content
			res.MatchedDiffRanges = rangesToInts(match.DiffPreview.MatchedRanges)
		}
		out[i] = res
	}
	return out
}

func rangesToInts(ranges result.Ranges) [][2]int {
	out := make([][2]int, len(ranges))
	for i, r := range ranges {
		out[i] = [2]int{r.Start.Offset, r.End.Offset}
	}
	return out
}
