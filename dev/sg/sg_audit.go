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

	"github.com/google/go-github/v41/github"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/slack-go/slack"
	sgslack "github.com/sourcegraph/sourcegraph/dev/sg/internal/slack"
	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	auditFlagSet    = flag.NewFlagSet("sg audit", flag.ExitOnError)
	auditPRFlagSet  = flag.NewFlagSet("sg audit pr", flag.ExitOnError)
	auditFormatFlag = auditPRFlagSet.String("format", "terminal", "Format to use for audit logs output")
)

var auditCommand = &ffcli.Command{
	Name:       "audit",
	ShortUsage: "sg audit [target]",
	ShortHelp:  "Display audit trail for resources",
	FlagSet:    auditFlagSet,
	Exec: func(ctx context.Context, args []string) error {
		return flag.ErrHelp
	},
	Subcommands: []*ffcli.Command{{
		Name:    "pr",
		FlagSet: auditPRFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			ghc := github.NewClient(nil)

			issues, err := fetchIssues(ctx, ghc)
			if err != nil {
				return err
			}
			slack, err := sgslack.NewClient(ctx)
			if err != nil {
				return err
			}
			prAuditIssues, err := presentIssues(ctx, ghc, slack, issues)
			if err != nil {
				return err
			}

			switch *auditFormatFlag {
			case "terminal":
				var sb strings.Builder
				err = formatMarkdown(prAuditIssues, &sb)
				if err != nil {
					return err
				}
				writePrettyMarkdown(sb.String())
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

func fetchIssues(ctx context.Context, ghc *github.Client) ([]*github.Issue, error) {
	var issues []*github.Issue
	for {
		is, r, err := ghc.Issues.ListByRepo(ctx, "sourcegraph", "sec-pr-audit-trail", &github.IssueListByRepoOptions{State: "open", Direction: "asc"})
		if err != nil {
			return nil, err
		}
		issues = append(issues, is...)
		if r.NextPage == 0 {
			break
		}
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
			return nil, err
		}

		res = append(res, prAuditIssue{
			Title:     title,
			Url:       issue.GetHTMLURL(),
			CreatedAt: fmt.Sprintf("%d days ago", time.Since(issue.GetCreatedAt())/(time.Hour*24)),
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

Read more about [test plans](https://docs.sourcegraph.com/dev/background-information/testing_principles#test-plans) and [reviews](https://docs.sourcegraph.com/dev/background-information/pull_request_reviews).
{{""}}
{{- range . }}
- _{{ .CreatedAt }}_ @{{ .Author }}
    - [{{.Title}}]({{.Url}})
{{- end }}
`
