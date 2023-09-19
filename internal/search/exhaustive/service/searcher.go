package service

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	sgtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func FromSearchClient(client client.SearchClient) NewSearcher {
	return newSearcherFunc(func(ctx context.Context, q string) (SearchQuery, error) {
		// TODO adjust NewSearch API to enforce the user passing in a user id.
		// IE do not rely on ctx actor since that could easily lead to a bug.

		// TODO this hack is an ugly workaround to get the plan and jobs to
		// get into a shape we like. it will break in bad ways but works for
		// EAP.
		q = "type:file index:no " + q

		inputs, err := client.Plan(
			ctx,
			"V3",
			nil,
			q,
			search.Precise,
			search.Streaming,
		)
		if err != nil {
			return nil, err
		}

		// Hacky for now, but hard to adjust client API just yet.
		inputs.Exhaustive = true

		exhaustive, err := jobutil.NewExhaustive(inputs)
		if err != nil {
			return nil, err
		}

		return searchQuery{
			exhaustive: exhaustive,
			clients:    client.JobClients(),
		}, nil
	})
}

// TODO maybe reuse for the fake
type newSearcherFunc func(context.Context, string) (SearchQuery, error)

func (f newSearcherFunc) NewSearch(ctx context.Context, q string) (SearchQuery, error) {
	return f(ctx, q)
}

type searchQuery struct {
	exhaustive jobutil.Exhaustive
	clients    job.RuntimeClients
}

// TODO make this an iterator return since the result could be large and the
// underlying infra already relies on iterators
func (s searchQuery) RepositoryRevSpecs(ctx context.Context) ([]types.RepositoryRevSpecs, error) {
	var repoRevSpecs []types.RepositoryRevSpecs
	it := s.exhaustive.RepositoryRevSpecs(ctx, s.clients)
	for it.Next() {
		repoRev := it.Current()
		var revspecs []string
		for _, rev := range repoRev.Revs {
			revStr := rev.String()
			// avoid storing empty string since our DB expects non-empty
			// string + this is easier to read in the DB.
			if revStr == "" {
				revStr = "HEAD"
			}
			revspecs = append(revspecs, revStr)
		}
		repoRevSpecs = append(repoRevSpecs, types.RepositoryRevSpecs{
			Repository:         repoRev.Repo.ID,
			RevisionSpecifiers: types.RevisionSpecifierJoin(revspecs),
		})
	}

	err := it.Err()
	if isReposMissingError(err) {
		// This isn't an error for us, we just don't search anything. We don't
		// have the concept of alerts yet in search jobs.
		err = nil
	}

	return repoRevSpecs, err
}

func (s searchQuery) ResolveRepositoryRevSpec(ctx context.Context, repoRevSpec types.RepositoryRevSpecs) ([]types.RepositoryRevision, error) {
	repoPagerRepoRevSpec, err := s.toRepoRevSpecs(ctx, repoRevSpec)
	if err != nil {
		return nil, err
	}

	page, err := s.exhaustive.ResolveRepositoryRevSpec(ctx, s.clients, []repos.RepoRevSpecs{repoPagerRepoRevSpec})
	if isReposMissingError(err) {
		// This isn't an error for us, we just don't search anything. We don't
		// have the concept of alerts yet in search jobs.
		err = nil
	}
	if err != nil {
		return nil, err
	}
	if page.BackendsMissing > 0 {
		return nil, errors.New("job needs to be retried, some backends are down")
	}
	var repoRevs []types.RepositoryRevision
	for _, repoRev := range page.RepoRevs {
		if repoRev.Repo.ID != repoRevSpec.Repository {
			return nil, errors.Errorf("ResolveRepositoryRevSpec returned a different repo (%d) to the input %v", repoRev.Repo.ID, repoRevSpec)
		}
		for _, rev := range repoRev.Revs {
			repoRevs = append(repoRevs, types.RepositoryRevision{
				RepositoryRevSpecs: repoRevSpec,
				Revision:           rev,
			})
		}
	}
	return repoRevs, nil
}

func (s searchQuery) toRepoRevSpecs(ctx context.Context, repoRevSpec types.RepositoryRevSpecs) (repos.RepoRevSpecs, error) {
	repo, err := s.minimalRepo(ctx, repoRevSpec.Repository)
	if err != nil {
		return repos.RepoRevSpecs{}, err
	}

	var revs []query.RevisionSpecifier
	for _, revspec := range repoRevSpec.RevisionSpecifiers.Get() {
		revs = append(revs, query.ParseRevisionSpecifier(revspec))
	}

	return repos.RepoRevSpecs{
		Repo: repo,
		Revs: revs,
	}, nil
}

func (s searchQuery) Search(ctx context.Context, repoRev types.RepositoryRevision, w CSVWriter) error {
	repo, err := s.minimalRepo(ctx, repoRev.Repository)
	if err != nil {
		return err
	}

	job := s.exhaustive.Job(&search.RepositoryRevisions{
		Repo: repo,
		Revs: []string{repoRev.Revision},
	})

	if err := w.WriteHeader("repo_id", "repo_name", "revision", "commit", "path"); err != nil {
		return err
	}

	repoID := strconv.Itoa(int(repoRev.Repository))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu          sync.Mutex // serialize writes to w
		writeRowErr error      // capture if w.WriteRow fails
	)

	// TODO currently ignoring returned Alert
	_, err = job.Run(ctx, s.clients, streaming.StreamFunc(func(se streaming.SearchEvent) {
		mu.Lock()
		defer mu.Unlock()

		for _, match := range se.Results {
			// TODO actually write useful CSV
			key := match.Key()
			err := w.WriteRow(repoID, string(key.Repo), repoRev.Revision, string(key.Commit), key.Path)
			if err != nil {
				cancel()
				writeRowErr = err
			}
		}
	}))

	if writeRowErr != nil {
		return writeRowErr
	}

	return err
}

func (s searchQuery) minimalRepo(ctx context.Context, repoID api.RepoID) (sgtypes.MinimalRepo, error) {
	minimalRepos, err := s.clients.DB.Repos().ListMinimalRepos(ctx, database.ReposListOptions{
		IDs: []api.RepoID{repoID},
	})
	if err != nil {
		return sgtypes.MinimalRepo{}, err
	}
	if len(minimalRepos) != 1 {
		return sgtypes.MinimalRepo{}, errors.Errorf("looking up repo %d found %d entries", repoID, len(minimalRepos))
	}
	return minimalRepos[0], nil
}

func isReposMissingError(err error) bool {
	var m repos.MissingRepoRevsError
	return errors.Is(err, repos.ErrNoResolvedRepos) || errors.HasType(err, &m)
}
