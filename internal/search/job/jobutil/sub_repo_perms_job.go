package jobutil

import (
	"context"
	"sync"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
		event.Results, err = applySubRepoFiltering(ctx, checker, clients.Logger, event.Results)
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

func (s *subRepoPermsFilterJob) Attributes(job.Verbosity) []attribute.KeyValue { return nil }

func (s *subRepoPermsFilterJob) Children() []job.Describer {
	return []job.Describer{s.child}
}

func (s *subRepoPermsFilterJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.child = job.Map(s.child, fn)
	return &cp
}

// applySubRepoFiltering filters a set of matches using the provided
// authz.SubRepoPermissionChecker
func applySubRepoFiltering(ctx context.Context, checker authz.SubRepoPermissionChecker, logger log.Logger, matches []result.Match) ([]result.Match, error) {
	if !authz.SubRepoEnabled(checker) {
		return matches, nil
	}

	a := actor.FromContext(ctx)
	var errs error

	// Filter matches in place
	filtered := matches[:0]

	errCache := map[api.RepoName]struct{}{} // cache repos that errored

	for _, m := range matches {
		// If the check errored before, skip the repo
		if _, ok := errCache[m.RepoName().Name]; ok {
			continue
		}
		enabled, err := authz.SubRepoEnabledForRepoID(ctx, checker, m.RepoName().ID)
		if err != nil {
			// If an error occurs while checking sub-repo perms, we omit it from the results
			if err != ctx.Err() {
				logger.Error("Could not determine if sub-repo permissions are enabled for repo, skipping", log.Error(err), log.String("repoName", string(m.RepoName().Name)))
			}
			errCache[m.RepoName().Name] = struct{}{}
			continue
		}
		if !enabled {
			filtered = append(filtered, m)
			continue
		}
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
			allowed, err := authz.CanReadAnyPath(ctx, checker, mm.Repo.Name, mm.ModifiedFiles)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}
			if allowed {
				if !diffIsEmpty(mm.DiffPreview) {
					filtered = append(filtered, m)
				}
			}
		case *result.RepoMatch:
			// Repo filtering is taken care of by our usual repo filtering logic
			filtered = append(filtered, m)
			// Owner matches are found after the sub-repo permissions filtering, hence why we don't have
			// an OwnerMatch case here.
		}
	}

	if errs == nil {
		return filtered, nil
	}

	// We don't want to return sensitive authz information or excluded paths to the
	// user so we'll return generic error and log something more specific.
	logger.Warn("Applying sub-repo permissions to search results", log.Error(errs))
	return filtered, errors.New("subRepoFilterFunc")
}

func diffIsEmpty(diffPreview *result.MatchedString) bool {
	if diffPreview != nil {
		if diffPreview.Content == "" || len(diffPreview.MatchedRanges) == 0 {
			return true
		}
	}
	return false
}
