package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var slackTemplate = `:arrow_left: *{{.Environment}}* deployment (<{{.BuildURL}}|build>)

- Services:
{{- range .Services }}
    - ` + "`" + `{{ . }}` + "`" + `
{{- end }}

- Pull Requests:
{{- range .PullRequests }}
    - <{{ .WebURL }}|{{ .Name }}> {{ .AuthorSlackID }}
{{- end }}`

type slackSummaryPresenter struct {
	Environment  string
	BuildURL     string
	Services     []string
	PullRequests []pullRequestPresenter
}

type pullRequestPresenter struct {
	Name          string
	AuthorSlackID string
	WebURL        string
}

func slackSummary(ctx context.Context, teammates team.TeammateResolver, report *DeploymentReport, traceURL string) (string, error) {
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
					return "", errors.Newf("pull request %d has no user", pr.GetNumber())
				}
				teammate, err := teammates.ResolveByGitHubHandle(ctx, user.GetLogin())
				if err != nil {
					return "", err
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

	tmpl, err := template.New("deployment-status-slack-summary").Parse(slackTemplate)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, presenter)
	if err != nil {
		return "", err
	}

	if traceURL != "" {
		_, err = sb.WriteString(fmt.Sprintf("\n<%s|Deployment trace>", traceURL))
		if err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

// postSlackUpdate attempts to send the given summary to at each of the provided webhooks.
func postSlackUpdate(webhook string, summary string) error {
	if webhook == "" {
		return nil
	}

	type slackText struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type slackBlock struct {
		Type string     `json:"type"`
		Text *slackText `json:"text,omitempty"`
	}

	var blocks []slackBlock
	for _, s := range strings.Split(summary, "\n\n") {
		blocks = append(blocks, slackBlock{
			Type: "section",
			Text: &slackText{
				Type: "mrkdwn",
				Text: s,
			},
		})
	}

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
