package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"
)

type SlackClient struct {
	slack.Client
	logger  log.Logger
	Channel string
}

func NewSlackClient(logger log.Logger, token string) *SlackClient {
	return &SlackClient{
		logger:  logger.Scoped("slack", "client which interacts with Slack's webhook API"),
		Client:  *slack.New(token),
		Channel: "#william-buildchecker-webhook-test",
	}
}

func (c *SlackClient) sendNotification(build *Build) error {
	c.logger.Debug("creating slack json", log.Int("buildNumber", *build.Number))

	user, err := c.GetUserInfo(build.Author.Email)
	if err != nil {
		c.logger.Error("failed to get slack user", log.Error(err), log.String("Author.Email", build.Author.Email))
	}
	blocks, err := createMessageBlocks(c.logger, user, build)
	if err != nil {
		return err
	}

	c.logger.Debug("sending notification", log.Int("buildNumber", *build.Number), log.String("channel", c.Channel))
	_, _, err = c.PostMessage(c.Channel, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		c.logger.Error("failed to post message", log.Int("buildNumber", *build.Number), log.Error(err))
		return err
	}

	c.logger.Info("notification posted", log.Int("buildNumber", *build.Number))
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

func createMessageBlocks(logger log.Logger, user *slack.User, build *Build) ([]slack.Block, error) {
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

	if user != nil {
		logger.Info("slack user", log.String("ID", user.ID), log.String("Name", user.Name))
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
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*:bust_in_silhouette: Author*\n%s", build.Author.Name)},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*:building_construction: Pipeline*\n%s", build.PipelineName())},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*:github: Commit*\n%s", commitLink(*build.Commit))},
			},
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
					URL:   grafanaURLFor(logger, build),
					Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Is this a flake ?"},
				},
			}...,
		),
	}

	return blocks, nil
}
