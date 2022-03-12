package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	repoOwner = "sourcegraph"
	repoName  = "sourcegraph"
)

type DeploymentNotifier struct {
	vr           VersionRequester
	ghc          *github.Client
	changedFiles []string
	targetCommit string
}

func NewDeploymentNotifier(ghc *github.Client, vr VersionRequester, targetCommit string, changedFiles []string) *DeploymentNotifier {
	return &DeploymentNotifier{
		vr:           vr,
		ghc:          ghc,
		changedFiles: changedFiles,
		targetCommit: targetCommit,
	}
}

// getNewCommits returns a slice of commits starting from the target commit up to the currently deployed commit.
func (dn *DeploymentNotifier) getNewCommits(ctx context.Context, lastCommit string) ([]*github.RepositoryCommit, error) {
	var page = 1
	var commits []*github.RepositoryCommit
	for page != 0 && page != 4 {
		cs, resp, err := dn.ghc.Repositories.ListCommits(ctx, repoOwner, repoName, &github.CommitsListOptions{
			SHA: "main",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 30,
			},
		})
		if err != nil {
			return nil, err
		}
		commits = append(commits, cs...)
		var currentCommitIdx int
		for i, commit := range commits {
			if commit.GetSHA() == dn.targetCommit {
				currentCommitIdx = i
			}
			if commit.GetSHA() == lastCommit {
				for _, c := range commits[currentCommitIdx:i] {
					fmt.Println(c.GetSHA())
				}
				return commits[currentCommitIdx:i], nil
			}
		}
		page = resp.NextPage
	}
	return nil, errors.Newf("commit %s not found in the last 90 commits", lastCommit)
}

func (dn *DeploymentNotifier) getNewPullRequests(ctx context.Context) ([]*github.PullRequest, error) {
	liveCommit, err := dn.vr.LastCommit()
	if err != nil {
		return nil, err
	}
	repoCommits, err := dn.getNewCommits(ctx, liveCommit)
	if err != nil {
		return nil, err
	}
	var pullsSinceLastCommit []*github.PullRequest
	for _, rc := range repoCommits {
		pulls, _, err := dn.ghc.PullRequests.ListPullRequestsWithCommit(
			ctx,
			repoOwner,
			repoName,
			rc.GetSHA(),
			&github.PullRequestListOptions{
				State: "merged",
			},
		)
		if err != nil {
			return nil, err
		}
		pullsSinceLastCommit = append(pullsSinceLastCommit, pulls...)
	}
	return pullsSinceLastCommit, nil
}

func (dn *DeploymentNotifier) Summary(ctx context.Context) (string, error) {
	prs, err := dn.getNewPullRequests(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println("found prs:")
	for _, pr := range prs {
		fmt.Println("-", pr.GetNumber())
	}

	renderComment(dn.deployedApps())
	return "", nil
}

func (dn *DeploymentNotifier) deployedApps() []string {
	var deployedApps []string
	for _, file := range dn.changedFiles {
		if filepath.Ext(file) != ".yaml" {
			continue
		}
		base := filepath.Base(file)
		components := strings.Split(base, ".")
		if len(components) < 3 {
			continue
		}
		// gitserver.[Deployment|StatefulSet|DaemonSet].yaml
		kind := components[1]

		if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" {
			deployedApps = append(deployedApps, components[0])
		}
	}
	return deployedApps
}

func renderComment(deployedApps []string) (string, error) {
	tmpl, err := template.New("deployment-status-comment").Parse(commentTemplate)
	if err != nil {
		return "", err
	}
	presenter := commentPresenter{
		DeployedAt: time.Now().In(time.UTC).Format(time.RFC822Z),
		Apps:       deployedApps,
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, presenter)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

type commentPresenter struct {
	DeployedAt        string
	Apps              []string
	BuildkiteBuildURL string
}

var commentTemplate = `### Deployment status

[Deployed at {{ .DeployedAt }}]({{ .BuildkiteBuildURL }}):

{{- range .Apps }}
- ` + "`" + `{{ . }}` + "`" + `
{{- end }}
`
