package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var token string
var date string
var pipeline string
var slack string
var shortDateFormat = "2006-01-02"
var longDateFormat = "2006-01-02 15:04 (MST)"

func init() {
	flag.StringVar(&token, "buildkite.token", "", "mandatory buildkite token")
	flag.StringVar(&date, "date", "", "date for builds")
	flag.StringVar(&pipeline, "buildkite.pipeline", "sourcegraph", "name of the pipeline to inspect")
	flag.StringVar(&slack, "slack.webhook", "", "Slack Webhook URL to post the results on")
}

type event struct {
	at          time.Time
	state       string
	buildURL    string
	buildNumber int
}

type report struct {
	details []string
	summary string
}

type slackBody struct {
	Blocks []slackBlock `json:"blocks"`
}

type slackBlock struct {
	Type     string      `json:"type"`
	Text     *slackText  `json:"text,omitempty"`
	Elements []slackText `json:"elements,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	flag.Parse()

	var t time.Time
	var err error
	if date != "" {
		t, err = time.Parse(shortDateFormat, date)
		if err != nil {
			panic(err)
		}
	} else {
		t = time.Now()
		t = t.Add(-1 * 24 * time.Hour)
	}

	config, err := buildkite.NewTokenConfig(token, false)
	if err != nil {
		panic(err)
	}
	client := buildkite.NewClient(config.Client())

	var builds []buildkite.Build
	nextPage := 0
	for {
		bs, resp, err := client.Builds.ListByPipeline("sourcegraph", pipeline, &buildkite.BuildsListOptions{
			Branch: "main",
			// Select all builds that finished on or after the beginning of the day ...
			FinishedFrom: BoD(t),
			// To those who were created before or on the end of the day.
			CreatedTo:   EoD(t),
			ListOptions: buildkite.ListOptions{Page: nextPage},
		})
		if err != nil {
			panic(err)
		}
		nextPage = resp.NextPage
		builds = append(builds, bs...)

		if nextPage == 0 {
			break
		}
	}

	if len(builds) == 0 {
		panic("no builds")
	}

	ends := []*event{}
	for _, b := range builds {
		if b.FinishedAt != nil {
			if b.FinishedAt.Time.Day() != t.Day() {
				// Because we select builds that can be created on a given day but may not have finished yet
				// we need to discard those.
				continue
			}
			ends = append(ends, &event{
				at:          b.FinishedAt.Time,
				state:       *b.State,
				buildURL:    *b.WebURL,
				buildNumber: *b.Number,
			})
		}
	}
	sort.Slice(ends, func(i, j int) bool { return ends[i].at.Before(ends[j].at) })

	var lastRed *event
	red := time.Duration(0)
	var report report
	for _, event := range ends {
		buildLink := slackLink(fmt.Sprintf("build %d", event.buildNumber), event.buildURL)
		if event.state == "failed" {
			// if a build failed, compute how much time until the next green
			lastRed = event
			report.details = append(report.details, fmt.Sprintf("Failure on %s: %s",
				event.at.Format(longDateFormat), buildLink))
		}
		if event.state == "passed" && lastRed != nil {
			// if a build passed and we previously were red, stop recording the duration.
			red += event.at.Sub(lastRed.at)
			lastRed = nil
			report.details = append(report.details, fmt.Sprintf("Fixed on %s: %s",
				event.at.Format(longDateFormat), buildLink))
		}
	}
	report.summary = fmt.Sprintf("On %s, the pipeline was red for *%s* - see the %s for more details.",
		t.Format(shortDateFormat), red.Round(time.Second).String(), slackLink("CI dashboard", ciDashboardURL(BoD(t), EoD(t))))

	if slack == "" {
		// If we're meant to print the results on stdout.
		for _, detail := range report.details {
			fmt.Println(detail)
		}
		fmt.Println(report.summary)
	} else if err := postOnSlack(&report); err != nil {
		panic(err)
	}
}

func postOnSlack(report *report) error {
	var text string
	for _, detail := range report.details {
		text += "• " + detail + " \n"
	}

	slackBody := slackBody{
		Blocks: []slackBlock{
			{
				Type: "section",
				Text: &slackText{
					Type: "mrkdwn",
					Text: report.summary,
				},
			},
		},
	}

	if len(report.details) > 0 {
		// Add the details block only if there are details, otherwise Slack API will
		// consider the block to be invalid and will reject it.
		var text string
		for _, detail := range report.details {
			text += "• " + detail + " \n"
		}
		slackBody.Blocks = append(slackBody.Blocks,
			slackBlock{
				Type: "context",
				Elements: []slackText{
					{
						Type: "mrkdwn",
						Text: text,
					},
				},
			},
		)
	}

	body, err := json.MarshalIndent(slackBody, "", "  ")
	if err != nil {
		return errors.Newf("failed to post on slack: %w", err)
	}
	// Perform the HTTP Post on the webhook
	req, err := http.NewRequest(http.MethodPost, slack, bytes.NewBuffer(body))
	if err != nil {
		return errors.Newf("failed to post on slack: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Newf("failed to post on slack: %w", err)
	}

	// Parse the response, to check if it succeeded
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if buf.String() != "ok" {
		return errors.Newf("failed to post on slack: %s", buf.String())
	}
	return nil
}

func BoD(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func EoD(t time.Time) time.Time {
	return BoD(t).Add(time.Hour * 24).Add(-1 * time.Nanosecond)
}

// slackLink returns Slack's weird markdown link format thing.
// https://api.slack.com/reference/surfaces/formatting#linking-urls
func slackLink(title, url string) string {
	return fmt.Sprintf("<%s|%s>", url, title)
}

// ciDashboardURL returns a link to our CI overview dashboard.
func ciDashboardURL(start, end time.Time) string {
	const dashboard = "https://sourcegraph.grafana.net/d/iBBWbxFnk/ci"
	return fmt.Sprintf("%s?from=%d&to=%d",
		dashboard, start.UnixMilli(), end.UnixMilli())
}
