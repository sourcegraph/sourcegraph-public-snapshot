package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
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
	blocks := createMessageBlocks(c.logger, build)

	c.logger.Debug("sending notification", log.Int("buildNumber", *build.Number), log.String("channel", c.Channel))
	_, _, err := c.PostMessage(c.Channel, slack.MsgOptionBlocks(blocks...))
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
	From int64 `json:"From"`
	To   int64 `json:"To"`
}

type GrafanaPayload struct {
	DataSource string         `json:"datasource"`
	Queries    []GrafanaQuery `json:"queries"`
	Range      GrafanaRange   `json:"range"`
}

func grafanaURLFor(logger log.Logger, build *Build) string {
	logger = logger.Scoped("grafana", "generates grafana links")
	base, _ := url.Parse("http://sourcegraph.grafana.net/explore")
	queryData := struct {
		Build int
	}{
		Build: *build.Number,
	}
	tmplt := template.Must(template.New("Expression").Parse(`{app="buildkite", build="{{.Build}}", branch="main", state="failed"} # to search the whole build remove job here!
    |~ "(?i)failed|panic" # this is a case insensitive regular expression, feel free to unleash your regex-fu!
    `))

	var buf bytes.Buffer
	if err := tmplt.Execute(&buf, queryData); err != nil {
		logger.Error("failed to execute template", log.String("template", "Expression"), log.Error(err))
	}
	var expression = buf.String()

	logger.Debug("---->", log.String("expression", expression))

	buf.Reset()
	begin := time.Now().Add(-(1 * time.Hour)).UnixNano()
	end := time.Now().Add(5 * time.Minute).UnixNano()

	data := GrafanaPayload{
		DataSource: "grafanacloud-sourcegraph-logs",
		Queries: []GrafanaQuery{
			{
				RefId: "A",
				Expr:  expression,
			},
		},
		Range: GrafanaRange{
			From: begin,
			To:   end,
		},
	}

	result, err := json.Marshal(data)
	if err != nil {
		logger.Error("failed to marshall GrafanaPayload", log.Error(err))
	}

	logger.Debug("---->", log.String("GrafanaPayload", (string(result))))
	var values = make(url.Values)
	values.Add("orgId", "1")
	values.Add("left", string(result))

	base.RawQuery = values.Encode()
	logger.Debug("---->", log.String("url params", base.RawQuery))

	return base.String()
}

func createMessageBlocks(logger log.Logger, build *Build) []slack.Block {
	failed := make([]*slack.TextBlockObject, 0)
	for _, j := range build.Jobs {
		if j.ExitStatus != nil && *j.ExitStatus != 0 && !j.SoftFailed {
			failed = append(failed, &slack.TextBlockObject{
				Type:  slack.PlainTextType,
				Text:  *j.Name,
				Emoji: true,
			})
		}
	}

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject(slack.PlainTextType, fmt.Sprintf("Build %d failed", *build.Number), true, false),
		),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewSectionBlock(
			&slack.TextBlockObject{Type: slack.MarkdownType, Text: "The following steps have failed"},
			failed,
			nil,
		),
		&slack.DividerBlock{
			Type: slack.MBTDivider,
		},
		slack.NewSectionBlock(
			&slack.TextBlockObject{Type: slack.MarkdownType, Text: fmt.Sprintf("<%s|%s>", *build.WebURL, *build.Message)},
			[]*slack.TextBlockObject{
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*Author*\n%s", build.Author.Name)},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*Pipeline*\n%s", *build.Pipeline.ID)},
				{Type: slack.MarkdownType, Text: fmt.Sprintf("*Commit*\n`%s`", *build.Commit)},
			}, nil,
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
					URL:   *build.URL,
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
					Text:  &slack.TextBlockObject{Type: slack.PlainTextType, Text: "Is this a Flake ?"},
				},
			}...,
		),
	}

	return blocks
}
