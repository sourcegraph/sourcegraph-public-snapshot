package github

import (
	"fmt"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/githubutil"
)

// RepoStatuses is a GitHub-backed implementation of the RepoStatuses
// store.
type RepoStatuses struct{}

var _ store.RepoStatuses = (*RepoStatuses)(nil)

func (s *RepoStatuses) GetCombined(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	repo, commitID := repoRev.URI, repoRev.CommitID
	if commitID == "" {
		return &sourcegraph.CombinedStatus{Rev: repoRev.Rev}, nil
	}

	owner, name, err := githubutil.SplitGitHubRepoURI(repo)
	if err != nil {
		return nil, err
	}
	status, _, err := client(ctx).repos.GetCombinedStatus(owner, name, commitID, nil)
	if err != nil {
		return nil, err
	}
	return combinedStatusFromGitHub(status), nil
}

func combinedStatusFromGitHub(ghstat *github.CombinedStatus) *sourcegraph.CombinedStatus {
	var s sourcegraph.CombinedStatus
	s.CommitID = *ghstat.SHA
	s.State = *ghstat.State
	for _, st := range ghstat.Statuses {
		s.Statuses = append(s.Statuses, &sourcegraph.RepoStatus{
			State:     *st.State,
			Context:   *st.Context,
			TargetURL: *st.TargetURL,
		})
	}
	return &s
}

func githubStatusFromStatus(stat *sourcegraph.RepoStatus) *github.RepoStatus {
	var ghstat github.RepoStatus
	ghstat.State = github.String(stat.State)
	ghstat.Context = github.String(stat.Context)
	ghstat.TargetURL = github.String(stat.TargetURL)
	return &ghstat
}

func (s *RepoStatuses) Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) error {
	repo, commitID := repoRev.URI, repoRev.CommitID
	if commitID == "" {
		return fmt.Errorf("creating named revision status unsupported for GitHub repositories")
	}

	owner, name, err := githubutil.SplitGitHubRepoURI(repo)
	if err != nil {
		return err
	}

	_, _, err = client(ctx).repos.CreateStatus(owner, name, commitID, githubStatusFromStatus(status))
	return err
}
