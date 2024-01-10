package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var slackTemplate = `:arrow_left: *{{.Environment}} deployment*
<{{.BuildURL}}|:hammer: Build>{{if .TraceURL}} <{{.TraceURL}}|:footprints: Trace>{{end}}

*Updated services:*
{{- range .Services }}
	• ` + "`" + `{{ . }}` + "`" + `
{{- end }}

Pull Requests:
{{- range .PullRequests }}
	• <{{ .WebURL }}|{{ .Name }}> {{ .AuthorSlackID }}
{{- end }}`

type slackSummaryPresenter struct {
	Environment  string
	BuildURL     string
	Services     []string
	PullRequests []pullRequestPresenter
	TraceURL     string
}

func (presenter *slackSummaryPresenter) toString() string {
	tmpl, err := template.New("deployment-status-slack-summary").Parse(slackTemplate)
	if err != nil {
		logger.Fatal("failed to parse Slack summary", log.Error(err))
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, presenter)
	if err != nil {
		logger.Fatal("failed to execute Slack template", log.Error(err))
	}
	return sb.String()
}

type pullRequestPresenter struct {
	Name          string
	AuthorSlackID string
	WebURL        string
}

func slackSummary(ctx context.Context, teammates team.TeammateResolver, report *DeploymentReport, traceURL string) (*slackSummaryPresenter, error) {
	presenter := &slackSummaryPresenter{
		Environment: report.Environment,
		BuildURL:    report.BuildkiteBuildURL,
		Services:    report.Services,
	}

	for _, pr := range report.PullRequests {
		var (
			notifyOnDeploy   bool
			notifyOnServices = map[string]struct{}{}
		)
		for _, label := range pr.Labels {
			if *label.Name == "notify-on-deploy" {
				notifyOnDeploy = true
			}
			// Allow users to label 'service/$svc' to get notified only for deployments
			// when specific services are rolled out
			if strings.HasPrefix(*label.Name, "service/") {
				service := strings.Split(*label.Name, "/")[1]
				if service != "" {
					notifyOnServices[service] = struct{}{}
				}
			}
		}

		var authorSlackID string
		if notifyOnDeploy {
			// Check if we should notify for this particular deployment
			var shouldNotify bool
			if len(notifyOnServices) == 0 {
				shouldNotify = true
			} else {
				// If the desired service is included, then notify
				for _, svc := range report.ServicesPerPullRequest[pr.GetNumber()] {
					if _, ok := notifyOnServices[svc]; ok {
						shouldNotify = true
						break
					}
				}
			}

			if shouldNotify {
				user := pr.GetUser()
				if user == nil {
					return nil, errors.Newf("pull request %d has no user", pr.GetNumber())
				}
				teammate, err := teammates.ResolveByGitHubHandle(ctx, user.GetLogin())
				if err != nil {
					return nil, err
				}
				authorSlackID = fmt.Sprintf("<@%s>", teammate.SlackID)
			}
		}

		presenter.PullRequests = append(presenter.PullRequests, pullRequestPresenter{
			Name:          pr.GetTitle(),
			WebURL:        pr.GetHTMLURL(),
			AuthorSlackID: authorSlackID,
		})
	}

	if traceURL != "" {
		presenter.TraceURL = traceURL
	}

	return presenter, nil
}

// postSlackUpdate attempts to send the given summary to at each of the provided webhooks.
func postSlackUpdate(webhook string, presenter *slackSummaryPresenter) error {
	if webhook == "" {
		return nil
	}

	type slackText struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type slackBlock struct {
		Type slack.MessageBlockType `json:"type"`
		Text *slackText             `json:"text,omitempty"`
		// For type 'context'
		Elements []*slackText `json:"elements,omitempty"`
	}

	var blocks []slackBlock
	buildInfoContent := []*slackText{{
		Type: slack.MarkdownType,
		Text: fmt.Sprintf("<%s|:hammer: Build>", presenter.BuildURL),
	}}
	if presenter.TraceURL != "" {
		buildInfoContent = append(buildInfoContent, &slackText{
			Type: slack.MarkdownType,
			Text: fmt.Sprintf("<%s|:footprints: Trace>\n", presenter.TraceURL),
		})
	}

	servicesContent := &slackText{
		Type: slack.MarkdownType,
		Text: "*Updated services:*\n",
	}
	for _, service := range presenter.Services {
		servicesContent.Text += fmt.Sprintf("\t• `%s`\n", service)
	}

	pullRequestsBlocks := []slackBlock{{
		Type: slack.MBTSection,
		Text: &slackText{
			Type: slack.MarkdownType,
			Text: "*Pull Requests:*\n",
		},
	}}
	for _, pullRequest := range presenter.PullRequests {
		currentTextBlock := pullRequestsBlocks[len(pullRequestsBlocks)-1].Text
		pullRequestText := fmt.Sprintf("\t• <%s|%s> %s\n", pullRequest.WebURL, pullRequest.Name, pullRequest.AuthorSlackID)

		if len(currentTextBlock.Text)+len(pullRequestText) < 3000 {
			// this PR text still fits within the character limit of a text block
			currentTextBlock.Text += pullRequestText
		} else {
			// this PR text exceeds the limit so a new section block is required
			pullRequestsBlocks = append(pullRequestsBlocks, slackBlock{
				Type: slack.MBTSection,
				Text: &slackText{
					Type: slack.MarkdownType,
					// add empty character to fix dumb Slack autoformatting
					Text: "\u200e" + pullRequestText,
				},
			})
		}
	}

	blocks = append(blocks,
		slackBlock{
			Type: slack.MBTHeader,
			Text: &slackText{
				Type: slack.PlainTextType,
				Text: fmt.Sprintf(":arrow_left: %s deployment", presenter.Environment),
			},
		},
		slackBlock{
			Type:     slack.MBTContext,
			Elements: buildInfoContent,
		},
		slackBlock{
			Type: slack.MBTSection,
			Text: servicesContent,
		})
	blocks = append(blocks, pullRequestsBlocks...)

	// Generate request
	body, err := json.MarshalIndent(struct {
		Blocks []slackBlock `json:"blocks"`
	}{
		Blocks: blocks,
	}, "", "  ")
	if err != nil {
		return errors.Newf("MarshalIndent: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	// Perform the HTTP Post on the webhook
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Parse the response, to check if it succeeded
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if buf.String() != "ok" {
		return errors.Newf("failed to post on slack: %q", buf.String())
	}
	return err
}
