package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type NotificationClient struct {
	slack   slack.Client
	team    team.TeammateResolver
	logger  log.Logger
	channel string
}

func NewNotificationClient(logger log.Logger, slackToken, githubToken, channel string) *NotificationClient {
	slack := slack.New(slackToken)

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	githubClient := github.NewClient(&httpClient)
	teamResolver := team.NewTeammateResolver(githubClient, slack)

	return &NotificationClient{
		logger:  logger.Scoped("notificationClient", "client which interacts with Slack and Github to send notifications"),
		slack:   *slack,
		team:    teamResolver,
		channel: channel,
	}
}

func (c *NotificationClient) getTeammateForBuild(build *Build) (*team.Teammate, error) {
	if build.Author == nil {
		return nil, errors.New("nil Author")
	}
	return c.team.ResolveByCommitAuthor(context.Background(), "sourcegraph", "sourcegraph", build.commit())
}

func (c *NotificationClient) send(build *Build) error {
	logger := c.logger.With(log.Int("buildNumber", build.number()), log.String("channel", c.channel))
	logger.Debug("creating slack json", log.Int("buildNumber", build.number()))

	teammate, err := c.getTeammateForBuild(build)
	if err != nil {
		logger.Error("failed to find teammate", log.Error(err))
	}

	blocks, err := createMessageBlocks(logger, teammate, build)
	if err != nil {
		return err
	}

	logger.Debug("sending notification")
	_, _, err = c.slack.PostMessage(c.channel, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		logger.Error("failed to post message", log.Error(err))
		return err
	}

	logger.Info("notification posted")
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

func grafanaURLFor(build *Build) (string, error) {
	queryData := struct {
		Build int
	}{
		Build: intp(build.Number),
	}
	tmpl := template.Must(template.New("Expression").Parse(`{app="buildkite", build="{{.Build}}", state="failed"} |~ "(?i)failed|panic|error|FAIL \\|"`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, queryData); err != nil {
		return "", err
	}
	var expression = buf.String()

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
		return "", errors.Wrap(err, "failed to marshall GrafanaPayload")
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

	return "https://sourcegraph.grafana.net/explore?orgId=1&left=" + query, nil
}

func commitLink(msg, commit string) string {
	repo := "http://github.com/sourcegraph/sourcegraph"
	sgURL := fmt.Sprintf("%s/commit/%s", repo, commit)
	return fmt.Sprintf("<%s|%s>", sgURL, msg)
}

func slackMention(teammate *team.Teammate, build *Build) string {
	if teammate == nil {
		authorName := build.authorName()
		if authorName == "" {
			authorName = "N/A"
		}
		return fmt.Sprintf("Teammate *%s* not found. If this is you, ensure the github field is set in your profile <https://github.com/sourcegraph/handbook/blob/main/data/team.yml|here>", authorName)
	}

	return fmt.Sprintf("<@%s>", teammate.SlackID)
}

func createMessageBlocks(logger log.Logger, teammate *team.Teammate, build *Build) ([]slack.Block, error) {
	msg, _, _ := strings.Cut(build.message(), "\n")
	msg += fmt.Sprintf(" (%s)", build.commit()[:7])
	failedSection := fmt.Sprintf("%s\n\n", commitLink(msg, build.commit()))
	failedSection += "*Failed jobs*\n\n"
	for _, j := range build.Jobs {
		if j.ExitStatus != nil && *j.ExitStatus != 0 && !j.SoftFailed {
			failedSection += fmt.Sprintf("• %s", *j.Name)
			if j.WebURL != "" {
				failedSection += fmt.Sprintf(" - <%s|logs>", j.WebURL)
			}
			failedSection += "\n"
		}
	}

	author := slackMention(teammate, build)
	grafanaURL, err := grafanaURLFor(build)
	if err != nil {
		return nil, err
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject(slack.PlainTextType, fmt.Sprintf(":red_circle: Build %d failed", build.number()), true, false),
		),
		slack.NewSectionBlock(
			nil,
			[]*slack.TextBlockObject{
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*Author:* %s", author)},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*Pipeline:* %s", build.Pipeline.name())},
			},
			nil,
		),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewSectionBlock(&slack.TextBlockObject{Type: slack.MarkdownType, Text: failedSection}, nil, nil),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: `:books: *More information on flakes*
• *<https://docs.sourcegraph.com/dev/background-information/ci#flakes|How to disable flakey tests>*
• *<https://docs.sourcegraph.com/dev/how-to/testing#assessing-flaky-client-steps|Recognizing flakey client steps and how to fix them>*

_:sourcegraph: disable flakes on sight and save your fellow teammate some time!_`,
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
					URL:  grafanaURL,
					Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "View logs on Grafana"},
				},
				&slack.ButtonBlockElement{
					Type: slack.METButton,
					URL:  "https://www.loom.com/share/58cedf44d44c45a292f650ddd3547337",
					Text: &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Is this a flake ?"},
				},
			}...,
		),
	}

	return blocks, nil
}
