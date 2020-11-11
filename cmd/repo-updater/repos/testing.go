package repos

import (
	"context"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// NewFakeSourcer returns a Sourcer which always returns the given error and sources,
// ignoring the given external services.
func NewFakeSourcer(err error, srcs ...Source) Sourcer {
	return func(svcs ...*types.ExternalService) (Sources, error) {
		var errs *multierror.Error

		if err != nil {
			for _, svc := range svcs {
				errs = multierror.Append(errs, &SourceError{Err: err, ExtSvc: svc})
			}
			if len(svcs) == 0 {
				errs = multierror.Append(errs, &SourceError{Err: err, ExtSvc: nil})
			}
		}

		return srcs, errs.ErrorOrNil()
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
		results <- SourceResult{Source: s, Repo: r.With(types.Opt.RepoSources(s.svc.URN()))}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s FakeSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

// This error is passed to txstore.Done in order to always
// roll-back the transaction a test case executes in.
// This is meant to ensure each test case has a clean slate.
var errRollback = errors.New("tx: rollback")

func Transact(ctx context.Context, s *Store, test func(testing.TB, *Store)) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		var err error
		txStore := s

		if !s.InTransaction() {
			txStore, err = s.Transact(ctx)
			if err != nil {
				t.Fatalf("failed to start transaction: %v", err)
			}
			defer txStore.Done(errRollback)
		}

		test(t, txStore)
	}
}
