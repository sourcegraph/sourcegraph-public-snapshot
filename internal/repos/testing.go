package repos

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFakeSourcer returns a Sourcer which always returns the given error and source,
// ignoring the given external services.
func NewFakeSourcer(err error, src Source) Sourcer {
	return func(svc *types.ExternalService) (Source, error) {
		if err != nil {
			return nil, &SourceError{Err: err, ExtSvc: svc}
		}
		return src, nil
	}
}

// FakeSource is a fake implementation of Source to be used in tests.
type FakeSource struct {
	svc   *types.ExternalService
	repos []*types.Repo
	err   error
}

// NewFakeSource returns an instance of FakeSource with the given urn, error
// and repos.
func NewFakeSource(svc *types.ExternalService, err error, rs ...*types.Repo) *FakeSource {
	return &FakeSource{svc: svc, err: err, repos: rs}
}

// ListRepos returns the Repos that FakeSource was instantiated with
// as well as the error, if any.
func (s FakeSource) ListRepos(ctx context.Context, results chan SourceResult) {
	if s.err != nil {
		results <- SourceResult{Source: s, Err: s.err}
		return
	}

	for _, r := range s.repos {
		results <- SourceResult{Source: s, Repo: r.With(typestest.Opt.RepoSources(s.svc.URN()))}
	}
}

func (s FakeSource) GetRepo(ctx context.Context, name string) (*types.Repo, error) {
	for _, r := range s.repos {
		if strings.HasSuffix(string(r.Name), name) {
			return r, s.err
		}
	}

	if s.err == nil {
		return nil, errors.New("not found")
	}

	return nil, s.err
}

// ExternalServices returns a singleton slice containing the external service.
func (s FakeSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}
