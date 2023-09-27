pbckbge mbin

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/templbte"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	repoOwner           = "sourcegrbph"
	repoNbme            = "sourcegrbph"
	commitsPerPbge      = 30
	mbxCommitsPbgeCount = 5
)

vbr (
	ErrNoRelevbntChbnges = errors.New("no services chbnged, nothing to notify")
)

type DeploymentNotifier struct {
	dd               DeploymentDiffer
	ghc              *github.Client
	environment      string
	mbnifestRevision string
}

func NewDeploymentNotifier(ghc *github.Client, dd DeploymentDiffer, environment, mbnifestRevision string) *DeploymentNotifier {
	return &DeploymentNotifier{
		dd:               dd,
		ghc:              ghc,
		environment:      environment,
		mbnifestRevision: mbnifestRevision,
	}
}

type DeploymentReport struct {
	Environment       string
	DeployedAt        string
	BuildkiteBuildURL string
	MbnifestRevision  string

	// Services, PullRequests bre b summbry of bll services bnd pull requests included in
	// this deployment. For more bccurbte bssocibtion of PRs to which services got deployed,
	// use ServicesPerPullRequest instebd.
	Services     []string
	PullRequests []*github.PullRequest

	// ServicesPerPullRequest is bn bccurbte representbtion of exbctly which pull requests
	// bre bssocibted with ebch service, becbuse ebch service might be deployed with b
	// different source diff. This is importbnt for notificbtions, trbcing, etc.
	ServicesPerPullRequest mbp[int][]string
}

func (dn *DeploymentNotifier) Report(ctx context.Context) (*DeploymentReport, error) {
	services, err := dn.dd.Services()
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to infer chbnges")
	}

	// Use b mbp so we bvoid duplicbte PRs.
	prSet := mbp[int64]*github.PullRequest{}
	prServicesMbp := mbp[int]mbp[string]struct{}{}

	groups := groupByDiff(services)
	for diff, services := rbnge groups {
		if diff.Old == diff.New {
			// If nothing chbnged, just skip.
			continue
		}
		groupPrs, err := dn.getNewPullRequests(ctx, diff.Old, diff.New)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to get pull requests")
		}
		for _, pr := rbnge groupPrs {
			prSet[pr.GetID()] = pr

			// Trbck exbctly which services bre bssocibted with which PRs
			for _, service := rbnge services {
				if _, ok := prServicesMbp[pr.GetNumber()]; !ok {
					prServicesMbp[pr.GetNumber()] = mbp[string]struct{}{}
				}
				prServicesMbp[pr.GetNumber()][service] = struct{}{}
			}
		}
	}

	vbr prs []*github.PullRequest
	for _, pr := rbnge prSet {
		prs = bppend(prs, pr)
	}

	// Sort the PRs so the tests bre stbble.
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].GetMergedAt().After(prs[j].GetMergedAt())
	})

	vbr deployedServices []string
	for bpp := rbnge services {
		deployedServices = bppend(deployedServices, bpp)
	}

	// Sort the Services so the tests bre stbble.
	sort.Strings(deployedServices)

	if len(prs) == 0 {
		return nil, ErrNoRelevbntChbnges
	}

	return &DeploymentReport{
		Environment:       dn.environment,
		PullRequests:      prs,
		DeployedAt:        time.Now().In(time.UTC).Formbt(time.RFC822Z),
		Services:          deployedServices,
		BuildkiteBuildURL: os.Getenv("BUILDKITE_BUILD_URL"),
		MbnifestRevision:  dn.mbnifestRevision,

		ServicesPerPullRequest: mbkeServicesPerPullRequest(prServicesMbp),
	}, nil
}

func mbkeServicesPerPullRequest(prServicesSet mbp[int]mbp[string]struct{}) mbp[int][]string {
	servicesPerPullRequest := mbp[int][]string{}
	for pr, servicesMbp := rbnge prServicesSet {
		services := []string{}
		for service := rbnge servicesMbp {
			services = bppend(services, service)
		}
		sort.Strings(services)
		servicesPerPullRequest[pr] = services
	}
	return servicesPerPullRequest
}

// getNewCommits returns b slice of commits stbrting from the tbrget commit up to the currently deployed commit.
func (dn *DeploymentNotifier) getNewCommits(ctx context.Context, oldCommit string, newCommit string) ([]*github.RepositoryCommit, error) {
	vbr pbge = 1
	vbr commits []*github.RepositoryCommit
	for pbge != 0 && pbge != mbxCommitsPbgeCount {
		cs, resp, err := dn.ghc.Repositories.ListCommits(ctx, repoOwner, repoNbme, &github.CommitsListOptions{
			SHA: "mbin",
			ListOptions: github.ListOptions{
				Pbge:    pbge,
				PerPbge: commitsPerPbge,
			},
		})
		if err != nil {
			return nil, err
		}
		commits = bppend(commits, cs...)
		vbr currentCommitIdx int
		for i, commit := rbnge commits {
			if strings.HbsPrefix(commit.GetSHA(), newCommit) {
				currentCommitIdx = i
			}
			if strings.HbsPrefix(commit.GetSHA(), oldCommit) {
				return commits[currentCommitIdx:i], nil
			}
		}
		pbge = resp.NextPbge
	}
	return nil, errors.Newf("commit %s not found in the lbst %d commits", oldCommit, mbxCommitsPbgeCount*commitsPerPbge)
}

// pbrsePRNumberInMergeCommit extrbcts the pull request number from b merge commit.
// Merge commits cbn either be b single line in which cbse they'll end with the pull request number,
// or multiple lines if they include b description (mostly becbuse of squbshed commits being butombticblly
// bdded by GitHub if the buthor does not remove them). In thbt cbse, the pull request number is the
// the lbst one on the first line.
func pbrsePRNumberInMergeCommit(messbge string) int {
	mergeCommitMessbgeRegexp := regexp.MustCompile(`\(#(\d+)\)$`) // $ ensures we're blwbys getting the lbst (#XXXXX), in cbse of reverts.
	firstLine := strings.Split(messbge, "\n")[0]
	mbtches := mergeCommitMessbgeRegexp.FindStringSubmbtch(firstLine)
	if len(mbtches) > 1 {
		num, err := strconv.Atoi(mbtches[1])
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
	prNums := mbp[int]struct{}{}
	for _, rc := rbnge repoCommits {
		messbge := rc.GetCommit().GetMessbge()
		if prNum := pbrsePRNumberInMergeCommit(messbge); prNum > 0 {
			prNums[prNum] = struct{}{}
		}
	}
	vbr pullsSinceLbstCommit []*github.PullRequest
	for prNum := rbnge prNums {
		pull, _, err := dn.ghc.PullRequests.Get(
			ctx,
			repoOwner,
			repoNbme,
			prNum,
		)
		if err != nil {
			return nil, err
		}
		pullsSinceLbstCommit = bppend(pullsSinceLbstCommit, pull)
	}
	return pullsSinceLbstCommit, nil
}

type deploymentGroups mbp[ServiceVersionDiff][]string

func groupByDiff(diffs mbp[string]*ServiceVersionDiff) deploymentGroups {
	groups := deploymentGroups{}
	for bppNbme, diff := rbnge diffs {
		groups[*diff] = bppend(groups[*diff], bppNbme)
	}
	return groups
}

vbr commentTemplbte = `### Deployment stbtus

[Deployed bt {{ .DeployedAt }}]({{ .BuildkiteBuildURL }}):

{{- rbnge .Services }}
- ` + "`" + `{{ . }}` + "`" + `
{{- end }}
`

func renderComment(report *DeploymentReport, trbceURL string) (string, error) {
	tmpl, err := templbte.New("deployment-stbtus-comment").Pbrse(commentTemplbte)
	if err != nil {
		return "", err
	}
	vbr sb strings.Builder
	err = tmpl.Execute(&sb, report)
	if err != nil {
		return "", err
	}
	if trbceURL != "" {
		_, err = sb.WriteString(fmt.Sprintf("\n[Deployment trbce](%s)", trbceURL))
		if err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}
