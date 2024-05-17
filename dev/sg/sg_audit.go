package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/slack-go/slack"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	sgslack "github.com/sourcegraph/sourcegraph/dev/sg/internal/slack"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var auditFormatFlag string
var auditPRGitHubToken string

var auditCommand = &cli.Command{
	Name:      "audit",
	Usage:     "Display audit trail for resources",
	ArgsUsage: "[target]",
	Hidden:    true,
	Category:  category.Company,
	Subcommands: []*cli.Command{{
		Name:  "pr",
		Usage: "Display audit trail for pull requests",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "format",
				Usage:       "Format to use for audit logs output",
				Value:       "terminal",
				DefaultText: "[markdown|terminal]",
				Destination: &auditFormatFlag,
			},
			&cli.StringFlag{
				Name:        "github.token",
				Usage:       "GitHub token to use when making API requests, defaults to $GITHUB_TOKEN.",
				Destination: &auditPRGitHubToken,
				Value:       os.Getenv("GITHUB_TOKEN"),
			},
		},
		Action: func(ctx *cli.Context) error {
			ghc := github.NewClient(oauth2.NewClient(ctx.Context, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: auditPRGitHubToken},
			)))

			logger := log.Scoped("auditPR")
			logger.Debug("fetching issues")
			issues, err := fetchIssues(ctx.Context, logger, ghc)
			if err != nil {
				return err
			}
			slack, err := sgslack.NewClient(ctx.Context, std.Out)
			if err != nil {
				return err
			}
			logger.Debug("formatting results")
			prAuditIssues, err := presentIssues(ctx.Context, ghc, slack, issues)
			if err != nil {
				return err
			}

			switch auditFormatFlag {
			case "terminal":
				var sb strings.Builder
				err = formatMarkdown(prAuditIssues, &sb)
				if err != nil {
					return err
				}
				std.Out.WriteMarkdown(sb.String())
			case "markdown":
				err = formatMarkdown(prAuditIssues, os.Stdout)
				if err != nil {
					return err
				}
			default:
				return flag.ErrHelp
			}

			return nil
		},
	}},
}

func fetchIssues(ctx context.Context, logger log.Logger, ghc *github.Client) ([]*github.Issue, error) {
	var issues []*github.Issue
	nextPage := 1
	for {
		logger.Debug("Listing issues", log.Int("nextPage", nextPage))
		is, r, err := ghc.Issues.ListByRepo(ctx, "sourcegraph", "sec-pr-audit-trail", &github.IssueListByRepoOptions{
			State:     "open",
			Direction: "asc",
			ListOptions: github.ListOptions{
				Page: nextPage,
			},
		})
		if err != nil {
			return nil, err
		}
		issues = append(issues, is...)
		if r.NextPage == 0 {
			break
		}
		nextPage = r.NextPage
	}
	return issues, nil
}

type prAuditIssue struct {
	Title     string
	Url       string
	Author    string
	CreatedAt string
}

func presentIssues(ctx context.Context, ghc *github.Client, slack *slack.Client, issues []*github.Issue) ([]prAuditIssue, error) {
	resolver := team.NewTeammateResolver(ghc, slack)

	var res []prAuditIssue
	for _, issue := range issues {
		assignee := issue.GetAssignee()
		if assignee == nil {
			return nil, errors.Newf("missing assignee in %s", issue.GetHTMLURL())
		}
		var title = issue.GetTitle()
		title = strings.ReplaceAll(title, "[", "")
		title = strings.ReplaceAll(title, "]", "")

		author, err := resolver.ResolveByGitHubHandle(ctx, assignee.GetLogin())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to format issue %s", issue.GetHTMLURL())
		}

		res = append(res, prAuditIssue{
			Title:     title,
			Url:       issue.GetHTMLURL(),
			CreatedAt: fmt.Sprintf("%d days ago", time.Since(issue.GetCreatedAt().Time)/(time.Hour*24)),
			Author:    author.SlackName, // Use author.SlackID in the next iteration, when automating the posting of this message
		})

	}
	return res, nil
}

func formatMarkdown(issues []prAuditIssue, w io.Writer) error {
	tmpl, err := template.New("pr-audit-report").Parse(auditMarkdownTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, issues)
}

var auditMarkdownTemplate = `*SOC2 Pull Request missing test plans :alert:*

> If you are mentioned in the following list, it means that one of your pull request has been merged without the mandatory test plan and review.

In order to be compliant with SOC2, you or someone from your team *must* document in the relevant issue why it was skipped and how you made sure that the changes aren't breaking anything.

1. Navigate to the issue mentioning you.
2. Explain why no test plan was provided and why the PR wasn't reviewed before being merged.
3. Close the issue.

Read more about [test plans](https://docs-legacy.sourcegraph.com/dev/background-information/testing_principles#test-plans) and [reviews](https://docs.sourcegraph.com/dev/background-information/pull_request_reviews).
{{""}}
{{- range . }}
- _{{ .CreatedAt }}_ @{{ .Author }}
    - [{{.Title}}]({{.Url}})
{{- end }}
`
