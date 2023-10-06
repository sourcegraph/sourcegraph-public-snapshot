package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	repoOwner           = "sourcegraph"
	repoName            = "sourcegraph"
	commitsPerPage      = 30
	maxCommitsPageCount = 5
)

var ErrNoRelevantChanges = errors.New("no services changed, nothing to notify")

type DeploymentNotifier struct {
	dd               DeploymentDiffer
	ghc              *github.Client
	environment      string
	manifestRevision string
}

func NewDeploymentNotifier(ghc *github.Client, dd DeploymentDiffer, environment, manifestRevision string) *DeploymentNotifier {
	return &DeploymentNotifier{
		dd:               dd,
		ghc:              ghc,
		environment:      environment,
		manifestRevision: manifestRevision,
	}
}

type DeploymentReport struct {
	Environment       string
	DeployedAt        string
	BuildkiteBuildURL string
	ManifestRevision  string

	// Services, PullRequests are a summary of all services and pull requests included in
	// this deployment. For more accurate association of PRs to which services got deployed,
	// use ServicesPerPullRequest instead.
	Services     []string
	PullRequests []*github.PullRequest

	// ServicesPerPullRequest is an accurate representation of exactly which pull requests
	// are associated with each service, because each service might be deployed with a
	// different source diff. This is important for notifications, tracing, etc.
	ServicesPerPullRequest map[int][]string
}

func (dn *DeploymentNotifier) Report(ctx context.Context) (*DeploymentReport, error) {
	services, err := dn.dd.Services()
	if err != nil {
		return nil, errors.Wrap(err, "failed to infer changes")
	}

	// Use a map so we avoid duplicate PRs.
	prSet := map[int64]*github.PullRequest{}
	prServicesMap := map[int]map[string]struct{}{}

	groups := groupByDiff(services)
	for diff, services := range groups {
		if diff.Old == diff.New {
			// If nothing changed, just skip.
			continue
		}
		groupPrs, err := dn.getNewPullRequests(ctx, diff.Old, diff.New)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get pull requests")
		}
		for _, pr := range groupPrs {
			prSet[pr.GetID()] = pr

			// Track exactly which services are associated with which PRs
			for _, service := range services {
				if _, ok := prServicesMap[pr.GetNumber()]; !ok {
					prServicesMap[pr.GetNumber()] = map[string]struct{}{}
				}
				prServicesMap[pr.GetNumber()][service] = struct{}{}
			}
		}
	}

	var prs []*github.PullRequest
	for _, pr := range prSet {
		prs = append(prs, pr)
	}

	// Sort the PRs so the tests are stable.
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].GetMergedAt().After(prs[j].GetMergedAt().Time)
	})

	var deployedServices []string
	for app := range services {
		deployedServices = append(deployedServices, app)
	}

	// Sort the Services so the tests are stable.
	sort.Strings(deployedServices)

	if len(prs) == 0 {
		return nil, ErrNoRelevantChanges
	}

	return &DeploymentReport{
		Environment:       dn.environment,
		PullRequests:      prs,
		DeployedAt:        time.Now().In(time.UTC).Format(time.RFC822Z),
		Services:          deployedServices,
		BuildkiteBuildURL: os.Getenv("BUILDKITE_BUILD_URL"),
		ManifestRevision:  dn.manifestRevision,

		ServicesPerPullRequest: makeServicesPerPullRequest(prServicesMap),
	}, nil
}

func makeServicesPerPullRequest(prServicesSet map[int]map[string]struct{}) map[int][]string {
	servicesPerPullRequest := map[int][]string{}
	for pr, servicesMap := range prServicesSet {
		services := []string{}
		for service := range servicesMap {
			services = append(services, service)
		}
		sort.Strings(services)
		servicesPerPullRequest[pr] = services
	}
	return servicesPerPullRequest
}

// getNewCommits returns a slice of commits starting from the target commit up to the currently deployed commit.
func (dn *DeploymentNotifier) getNewCommits(ctx context.Context, oldCommit string, newCommit string) ([]*github.RepositoryCommit, error) {
	page := 1
	var commits []*github.RepositoryCommit
	for page != 0 && page != maxCommitsPageCount {
		cs, resp, err := dn.ghc.Repositories.ListCommits(ctx, repoOwner, repoName, &github.CommitsListOptions{
			SHA: "main",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: commitsPerPage,
			},
		})
		if err != nil {
			return nil, err
		}
		commits = append(commits, cs...)
		var currentCommitIdx int
		for i, commit := range commits {
			if strings.HasPrefix(commit.GetSHA(), newCommit) {
				currentCommitIdx = i
			}
			if strings.HasPrefix(commit.GetSHA(), oldCommit) {
				return commits[currentCommitIdx:i], nil
			}
		}
		page = resp.NextPage
	}
	return nil, errors.Newf("commit %s not found in the last %d commits", oldCommit, maxCommitsPageCount*commitsPerPage)
}

// parsePRNumberInMergeCommit extracts the pull request number from a merge commit.
// Merge commits can either be a single line in which case they'll end with the pull request number,
// or multiple lines if they include a description (mostly because of squashed commits being automatically
// added by GitHub if the author does not remove them). In that case, the pull request number is the
// the last one on the first line.
func parsePRNumberInMergeCommit(message string) int {
	mergeCommitMessageRegexp := regexp.MustCompile(`\(#(\d+)\)$`) // $ ensures we're always getting the last (#XXXXX), in case of reverts.
	firstLine := strings.Split(message, "\n")[0]
	matches := mergeCommitMessageRegexp.FindStringSubmatch(firstLine)
	if len(matches) > 1 {
		num, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0
		}
		return num
	}
	return 0
}

func (dn *DeploymentNotifier) getNewPullRequests(ctx context.Context, oldCommit string, newCommit string) ([]*github.PullRequest, error) {
	repoCommits, err := dn.getNewCommits(ctx, oldCommit, newCommit)
	if err != nil {
		return nil, err
	}
	prNums := map[int]struct{}{}
	for _, rc := range repoCommits {
		message := rc.GetCommit().GetMessage()
		if prNum := parsePRNumberInMergeCommit(message); prNum > 0 {
			prNums[prNum] = struct{}{}
		}
	}
	var pullsSinceLastCommit []*github.PullRequest
	for prNum := range prNums {
		pull, _, err := dn.ghc.PullRequests.Get(
			ctx,
			repoOwner,
			repoName,
			prNum,
		)
		if err != nil {
			return nil, err
		}
		pullsSinceLastCommit = append(pullsSinceLastCommit, pull)
	}
	return pullsSinceLastCommit, nil
}

type deploymentGroups map[ServiceVersionDiff][]string

func groupByDiff(diffs map[string]*ServiceVersionDiff) deploymentGroups {
	groups := deploymentGroups{}
	for appName, diff := range diffs {
		groups[*diff] = append(groups[*diff], appName)
	}
	return groups
}

var commentTemplate = `### Deployment status

[Deployed at {{ .DeployedAt }}]({{ .BuildkiteBuildURL }}):

{{- range .Services }}
- ` + "`" + `{{ . }}` + "`" + `
{{- end }}
`

func renderComment(report *DeploymentReport, traceURL string) (string, error) {
	tmpl, err := template.New("deployment-status-comment").Parse(commentTemplate)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	err = tmpl.Execute(&sb, report)
	if err != nil {
		return "", err
	}
	if traceURL != "" {
		_, err = sb.WriteString(fmt.Sprintf("\n[Deployment trace](%s)", traceURL))
		if err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}
