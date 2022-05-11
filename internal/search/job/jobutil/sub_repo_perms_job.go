package jobutil

import (
	"context"
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFilterJob creates a job that filters the streamed results
// of its child job using the default authz.DefaultSubRepoPermsChecker.
func NewFilterJob(child job.Job) job.Job {
	return &subRepoPermsFilterJob{child: child}
}

type subRepoPermsFilterJob struct {
	child job.Job
}

func (s *subRepoPermsFilterJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	checker := authz.DefaultSubRepoPermsChecker

	var (
		mu   sync.Mutex
		errs error
	)

	filteredStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		var err error
		event.Results, err = applySubRepoFiltering(ctx, checker, event.Results)
		if err != nil {
			mu.Lock()
			errs = errors.Append(errs, err)
			mu.Unlock()
		}
		stream.Send(event)
	})

	alert, err = s.child.Run(ctx, clients, filteredStream)
	if err != nil {
		errs = errors.Append(errs, err)
	}
	return alert, errs
}

func (s *subRepoPermsFilterJob) Name() string {
	return "SubRepoPermsFilterJob"
}

// applySubRepoFiltering filters a set of matches using the provided
// authz.SubRepoPermissionChecker
func applySubRepoFiltering(ctx context.Context, checker authz.SubRepoPermissionChecker, matches []result.Match) ([]result.Match, error) {
	if !authz.SubRepoEnabled(checker) {
		return matches, nil
	}

	a := actor.FromContext(ctx)
	var errs error

	// Filter matches in place
	filtered := matches[:0]

	for _, m := range matches {
		switch mm := m.(type) {
		case *result.FileMatch:
			repo := mm.Repo.Name
			matchedPath := mm.Path

			content := authz.RepoContent{
				Repo: repo,
				Path: matchedPath,
			}
			perms, err := authz.ActorPermissions(ctx, checker, a, content)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}

			if perms.Include(authz.Read) {
				filtered = append(filtered, m)
			}
		case *result.CommitMatch:
			allowed, err := authz.CanReadAllPaths(ctx, checker, mm.Repo.Name, mm.ModifiedFiles)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}
			if allowed {
				filtered = append(filtered, m)
			}
		case *result.RepoMatch:
			// Repo filtering is taking care of by our usual repo filtering logic
			filtered = append(filtered, m)
		}

	}

	if errs == nil {
		return filtered, nil
	}

	// We don't want to return sensitive authz information or excluded paths to the
	// user so we'll return generic error and log something more specific.
	log15.Warn("Applying sub-repo permissions to search results", "error", errs)
	return filtered, errors.New("subRepoFilterFunc")
}
