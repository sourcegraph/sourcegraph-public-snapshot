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

func (s *repos) newRepoFromGitHubOrigin(ctx context.Context, o *sourcegraph.Origin) (*sourcegraph.Repo, error) {
	gitHubID, err := strconv.Atoi(o.ID)
	if err != nil {
		return nil, err
	}

	ghrepo, err := github.ReposFromContext(ctx).GetByID(ctx, gitHubID)
	if err != nil {
		return nil, err
	}

	// Purposefully set very few fields. We don't want to cache
	// metadata, because it'll get stale, and fetching online from
	// GitHub is quite easy and (with HTTP caching) performant.
	ts := pbtypes.NewTimestamp(time.Now())
	return &sourcegraph.Repo{
		Owner:        ghrepo.Owner,
		Name:         ghrepo.Name,
		URI:          githubutil.RepoURI(ghrepo.Owner, ghrepo.Name),
		HTTPCloneURL: ghrepo.HTTPCloneURL,
		Description:  ghrepo.Description,
		Mirror:       true,
		Fork:         ghrepo.Fork,
		CreatedAt:    &ts,

		// KLUDGE: set this to be true to avoid accidentally treating
		// a private GitHub repo as public (the real value should be
		// populated from GitHub on the fly).
		Private: true,

		Origin: o,
	}, nil
}
