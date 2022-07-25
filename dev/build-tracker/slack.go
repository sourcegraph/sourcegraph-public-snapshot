package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/dev/team"
)

var unknownTeammate = team.Teammate{Name: "_N/A_"}

type NotificationClient struct {
	slack   slack.Client
	team    team.TeammateResolver
	logger  log.Logger
	Channel string
}

func NewNotificationClient(logger log.Logger, slackToken, githubToken, channel string) *NotificationClient {
	slack := slack.New(slackToken)
	githubClient := github.NewClient(http.DefaultClient)
	teamResolver := team.NewTeammateResolver(githubClient, slack)

	return &NotificationClient{
		logger:  logger.Scoped("notificationClient", "client which interacts with Slack and Github to send notifications"),
		slack:   *slack,
		team:    teamResolver,
		Channel: channel,
	}
}

func (c *NotificationClient) getTeammateForBuild(build *Build) (*team.Teammate, error) {
	if build.Author == nil {
		return nil, errors.New("nil Author")
	}
	name := build.Author.Name
	teammate, err := c.team.ResolveByName(context.Background(), name)
	if err != nil {
		// If we can't find the teammate then they probably aren't one, so lets set the teammate name to the author name
		return &team.Teammate{
			Name: name,
		}, nil
	}

	return teammate, nil
}

func (c *NotificationClient) sendNotification(build *Build) error {
	notifcationLogger := c.logger.With(log.Int("buildNumber", *build.Number), log.String("channel", c.Channel))
	notifcationLogger.Debug("creating slack json", log.Int("buildNumber", *build.Number))

	teammate, err := c.getTeammateForBuild(build)
	if err != nil {
		notifcationLogger.Error("failed to find teammate - using 'unknownTeammate'", log.Error(err))
		teammate = &unknownTeammate
	}

	blocks, err := createMessageBlocks(notifcationLogger, teammate, build)
	if err != nil {
		return err
	}

	notifcationLogger.Debug("sending notification")
	_, _, err = c.slack.PostMessage(c.Channel, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		notifcationLogger.Error("failed to post message", log.Error(err))
		return err
	}

	notifcationLogger.Info("notification posted")
	return nil
}

type GrafanaQuery struct {
	RefId string `json:"refId"`
	Expr  string `json:"expr"`
}

type GrafanaRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GrafanaPayload struct {
	DataSource string         `json:"datasource"`
	Queries    []GrafanaQuery `json:"queries"`
	Range      GrafanaRange   `json:"range"`
}

func grafanaURLFor(logger log.Logger, build *Build) string {
	logger = logger.Scoped("grafana", "generates grafana links")
	base, _ := url.Parse("https://sourcegraph.grafana.net/explore")
	queryData := struct {
		Build int
	}{
		Build: *build.Number,
	}
	tmplt := template.Must(template.New("Expression").Parse(`{app="buildkite", build="{{.Build}}", state="failed"} |~ "(?i)failed|panic|error|FAIL \\|"`))

	var buf bytes.Buffer
	if err := tmplt.Execute(&buf, queryData); err != nil {
		logger.Error("failed to execute template", log.String("template", "Expression"), log.Error(err))
	}
	var expression = buf.String()

	buf.Reset()
	begin := time.Now().Add(-(2 * time.Hour)).UnixMilli()
	end := time.Now().Add(15 * time.Minute).UnixMilli()

	data := GrafanaPayload{
		DataSource: "grafanacloud-sourcegraph-logs",
		Queries: []GrafanaQuery{
			{
				RefId: "A",
				Expr:  expression,
			},
		},
		Range: GrafanaRange{
			From: fmt.Sprintf("%d", begin),
			To:   fmt.Sprintf("%d", end),
		},
	}

	result, err := json.Marshal(data)
	if err != nil {
		logger.Error("failed to marshall GrafanaPayload", log.Error(err))
	}

	query := url.PathEscape(string(result))
	// default query escapes ":", which we  don't want since it is json. Query and Path escape
	// escape a few characters incorrectly so we fix it with this replacer.
	// Got the idea from https://sourcegraph.com/github.com/kubernetes/kubernetes/-/blob/vendor/github.com/PuerkitoBio/urlesc/urlesc.go?L115-121
	replacer := strings.NewReplacer(
		"+", "%20",
		"%28", "(",
		"%29", ")",
		"=", "%3D",
		"%2C", ",",
	)
	query = replacer.Replace(query)

	return base.String() + "?orgId=1&left=" + query
}

func commitLink(commit string) string {
	repo := "http://github.com/sourcegraph/sourcegraph"
	sgURL := fmt.Sprintf("%s/commit/%s", repo, commit)
	return fmt.Sprintf("<%s|%s>", sgURL, commit[:10])
}

func createMessageBlocks(logger log.Logger, teammate *team.Teammate, build *Build) ([]slack.Block, error) {
	failedJobs := "*Failed jobs:*\n"
	for _, j := range build.Jobs {
		if j.ExitStatus != nil && *j.ExitStatus != 0 && !j.SoftFailed {
			failedJobs += fmt.Sprintf("• %s", *j.Name)
			if j.WebURL != "" {
				failedJobs += fmt.Sprintf(" - <%s|logs>", j.WebURL)
			}
			failedJobs += "\n"
		}
	}

	if build.Number == nil {
		return nil, fmt.Errorf("cannot create message blocks for nil Build Number")
	}

	// should have the @ before here to tag people, but leaving that out for now
	author := teammate.SlackID
	if author == "" {
		author = teammate.Name
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject(slack.PlainTextType, fmt.Sprintf(":red_circle: Build %d failed", *build.Number), true, false),
		),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewSectionBlock(&slack.TextBlockObject{Type: slack.MarkdownType, Text: failedJobs}, nil, nil),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*:bust_in_silhouette: Author*\n%s", author)},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*:building_construction: Pipeline*\n%s", build.PipelineName())},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*:github: Commit*\n%s", commitLink(*build.Commit))},
			},
			nil,
		),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: `:brain: *More information on flakes* :point_down:
:one: *<https://docs.sourcegraph.com/dev/background-information/ci#flakes|CI flakes>*
:two: *<https://docs.sourcegraph.com/dev/how-to/testing#assessing-flaky-client-steps|Assessing flakey client steps>*

_"save your fellow dev some time and proactive disable flakes when you spot them"_`,
			},
			nil,
			nil,
		),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewActionBlock(
			"",
			[]slack.BlockElement{
				&slack.ButtonBlockElement{
					Type:  slack.METButton,
					Style: slack.StylePrimary,
					URL:   *build.WebURL,
					Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Go to build"},
				},
				&slack.ButtonBlockElement{
					Type: slack.METButton,
					URL:  grafanaURLFor(logger, build),
					Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "View logs on Grafana"},
				},
				&slack.ButtonBlockElement{
					Type:  slack.METButton,
					Style: slack.StyleDanger,
					URL:   "https://www.loom.com/share/58cedf44d44c45a292f650ddd3547337",
					Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Is this a flake ?"},
				},
			}...,
		),
	}

	return blocks, nil
}
