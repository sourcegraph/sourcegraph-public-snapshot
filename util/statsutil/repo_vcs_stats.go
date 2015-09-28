package statsutil

import (
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// This function silently ignores errors since we don't want to spam
// the local instance's logs about failing to compute committer stats.
// We will spot discrepancies upstream on PromDash.
func CountCommittersPerDomain(cl *sourcegraph.Client, ctx context.Context, reposList *sourcegraph.RepoList) (map[string]int32, error) {
	numCommitters := map[string]int32{}
	for _, r := range reposList.Repos {
		if r.VCS != "git" {
			// commit statistics not yet supported for other repo types.
			log15.Debug("CountCommittersPerDomain: committer stats not supported for repo type", "repo", r.URI, "type", r.VCS)
			continue
		}

		committersList, err := cl.Repos.ListCommitters(ctx, &sourcegraph.ReposListCommittersOp{
			Repo: sourcegraph.RepoSpec{URI: r.URI},
			Opt: &sourcegraph.RepoListCommittersOptions{
				Rev: r.DefaultBranch,
				ListOptions: sourcegraph.ListOptions{
					// get stats for a maximum of 1000 authors (ordered in decreasing
					// number of commits).
					PerPage: 1000,
				},
			},
		})
		if err != nil {
			log15.Debug("CountCommittersPerDomain: could not get committers list", "repo", r.URI, "error", err)
			continue
		}

		for _, committer := range committersList.Committers {
			var domain string
			if strings.Contains(committer.Email, "@") {
				domain = strings.SplitN(committer.Email, "@", 2)[1]
			} else {
				domain = committer.Email
			}

			if v, ok := numCommitters[domain]; ok {
				numCommitters[domain] = v + 1
			} else {
				numCommitters[domain] = 1
			}
		}
	}
	return numCommitters, nil
}
