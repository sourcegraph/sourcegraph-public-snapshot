package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
)

var token string
var date string
var pipeline string
var slack string
var days int
var shortDateFormat = "2006-01-02"
var longDateFormat = "2006-01-02 15:04 (MST)"
var googleOAuthToken string
var googleOAuthConfig string
var spreadsheetID string
var authenticate bool

func init() {
	flag.StringVar(&token, "token", "", "mandatory buildkite token")
	flag.StringVar(&date, "date", "", "date for builds, ex: 2021-12-10")
	flag.IntVar(&days, "days", 1, "how many days (and thus reports) to include")
	flag.StringVar(&pipeline, "pipeline", "sourcegraph", "name of the pipeline to inspect")
	flag.StringVar(&slack, "slack", "", "Slack Webhook URL to post the results on")
	flag.StringVar(&googleOAuthToken, "google-oauth.token", "", "Google OAuth Token")
	flag.StringVar(&googleOAuthConfig, "google-oauth.config", "", "Google OAuth Config")
	flag.StringVar(&spreadsheetID, "spreadsheet.id", "", "Google Spreadsheet ID, (https://docs.google.com/spreadsheets/d/[ID]/edit)")
	flag.BoolVar(&authenticate, "google-oauth.authenticate", false, "Prompt the user to sign-in to generate a token")
}

type event struct {
	at          time.Time
	state       string
	buildURL    string
	buildNumber int
}

type report struct {
	downtime time.Duration
	t        time.Time
	details  []string
	summary  string
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

func getDownTime(bkc *buildkite.Client, t time.Time) (*report, error) {
	var builds []buildkite.Build
	nextPage := 0
	for {
		bs, resp, err := bkc.Builds.ListByPipeline("sourcegraph", pipeline, &buildkite.BuildsListOptions{
			Branch: "main",
			// Select all builds that finished on or after the beginning of the day ...
			FinishedFrom: BoD(t),
			// To those who were created before or on the end of the day.
			CreatedTo:   EoD(t),
			ListOptions: buildkite.ListOptions{Page: nextPage},
		})
		if err != nil {
			return nil, err
		}
		nextPage = resp.NextPage
		builds = append(builds, bs...)

		if nextPage == 0 {
			break
		}
	}

	if len(builds) == 0 {
		return nil, nil
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
	report := report{
		t: t,
	}

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

	report.downtime = red

	return &report, nil
}

func main() {
	flag.Parse()

	// If the authenticate flag was passed, start the google OAuth authentication process
	if authenticate {
		cfg, err := getOAuthConfig(googleOAuthConfig)
		if err != nil {
			log.Fatalf("Invalid OAuth config: %s", err)
		}
		token := getTokenFromWeb(cfg)
		jsonToken, _ := json.Marshal(token)
		fmt.Printf("Please find your token below:\n%s\n", jsonToken)
		fmt.Printf("You can now use the flag -google-oauth.token=YOUR_TOKEN to automate metrics reporting in %s\n",
			spreadsheetID,
		)
		return
	}

	// Parse the date flag
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

	reports := []*report{}

	for i := 0; i < days; i++ {
		report, err := getDownTime(client, t)
		if err != nil {
			panic(err)
		}
		reports = append(reports, report)
		t = t.Add(-24 * time.Hour)
	}

	if googleOAuthToken != "" {
		for _, report := range reports {
			if report == nil {
				continue
			}
			// If a google token was passed, update the spreadsheet as well.
			err := pushReport(context.Background(), googleOAuthConfig, googleOAuthToken, report.t, report.downtime)
			if err != nil {
				panic(err)
			}
		}
	}

	if slack == "" {
		// If we're meant to print the results on stdout, loop through the results.
		for _, report := range reports {
			if report == nil {
				fmt.Println("no builds, skipping")
				continue
			}
			for _, detail := range report.details {
				fmt.Println(detail)
			}
			fmt.Println(report.summary)
		}
	} else {
		report := reports[len(reports)-1]
		if report == nil {
			fmt.Println("no builds, skipping")
		}
		if err := postOnSlack(report); err != nil {
			panic(err)
		}
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
			{
				Type: "context",
				Elements: []slackText{
					{
						Type: "mrkdwn",
						Text: text,
					},
				},
			},
		},
	}

	body, err := json.MarshalIndent(slackBody, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to post on slack: %w", err)
	}
	// Perform the HTTP Post on the webhook
	req, err := http.NewRequest(http.MethodPost, slack, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to post on slack: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
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
