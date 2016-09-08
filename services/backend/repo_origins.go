package backend

import (
	"strconv"
	"time"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	"sourcegraph.com/sqs/pbtypes"
)

// newRepoFromOrigin creates a new repo from origin o. o must not be nil.
func (s *repos) newRepoFromOrigin(ctx context.Context, o *sourcegraph.Origin) (*sourcegraph.Repo, error) {
	if err := checkValidOriginAndSetDefaultURL(o); err != nil {
		return nil, err
	}
	switch o.Service {
	case sourcegraph.Origin_GitHub:
		return s.newRepoFromGitHubOrigin(ctx, o)

	default:
		return nil, errInvalidOriginService
	}
}

// newRepoFromGitHubOrigin creates a new repo from a GitHub origin o. o must not be nil.
// TODO: This helper should be inlined into newRepoFromOrigin, the only place that uses it.
//       It's not a clean, meaningful abstraction, so having it be a separate func hurts readability
//       instead of improving it.
func (s *repos) newRepoFromGitHubOrigin(ctx context.Context, o *sourcegraph.Origin) (*sourcegraph.Repo, error) {
	githubID, err := strconv.Atoi(o.ID)
	if err != nil {
		return nil, err
	}

	ghRepo, err := github.ReposFromContext(ctx).GetByID(ctx, githubID)
	if err != nil {
		return nil, err
	}

	// Purposefully set very few fields. We don't want to cache
	// metadata, because it'll get stale, and fetching online from
	// GitHub is quite easy and (with HTTP caching) performant.
	ts := pbtypes.NewTimestamp(time.Now())
	return &sourcegraph.Repo{
		Owner:        ghRepo.Owner,
		Name:         ghRepo.Name,
		URI:          githubutil.RepoURI(ghRepo.Owner, ghRepo.Name),
		HTTPCloneURL: ghRepo.HTTPCloneURL,
		Description:  ghRepo.Description,
		Mirror:       true,
		Fork:         ghRepo.Fork,
		CreatedAt:    &ts,

		// KLUDGE: set this to be true to avoid accidentally treating
		// a private GitHub repo as public (the real value should be
		// populated from GitHub on the fly).
		Private: true,

		Origin: o,
	}, nil
}
