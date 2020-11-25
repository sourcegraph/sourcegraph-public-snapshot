package repos

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		results <- SourceResult{Source: s, Repo: r.With(Opt.RepoSources(s.svc.URN()))}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s FakeSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

//
// Assertions
//

// A ReposAssertion performs an assertion on the given Repos.
type ReposAssertion func(testing.TB, types.Repos)

// An ExternalServicesAssertion performs an assertion on the given
// ExternalServices.
type ExternalServicesAssertion func(testing.TB, types.ExternalServices)

// Assert contains assertion functions to be used in tests.
var Assert = struct {
	ReposEqual                func(...*types.Repo) ReposAssertion
	ReposOrderedBy            func(func(a, b *types.Repo) bool) ReposAssertion
	ExternalServicesOrderedBy func(func(a, b *types.ExternalService) bool) ExternalServicesAssertion
	ExternalServicesEqual     func(...*types.ExternalService) ExternalServicesAssertion
}{
	ReposEqual: func(rs ...*types.Repo) ReposAssertion {
		want := append(types.Repos{}, rs...).With(Opt.RepoID(0))
		return func(t testing.TB, have types.Repos) {
			t.Helper()
			// Exclude auto-generated IDs from equality tests
			have = append(types.Repos{}, have...).With(Opt.RepoID(0))
			if !reflect.DeepEqual(have, want) {
				t.Errorf("repos (-want +got): %s", cmp.Diff(want, have))
			}
		}
	},
	ReposOrderedBy: func(ord func(a, b *types.Repo) bool) ReposAssertion {
		return func(t testing.TB, have types.Repos) {
			t.Helper()
			want := have.Clone()
			sort.Slice(want, func(i, j int) bool {
				return ord(want[i], want[j])
			})
			if !reflect.DeepEqual(have, want) {
				t.Errorf("repos (-want +got): %s", cmp.Diff(want, have))
			}
		}
	},
	ExternalServicesEqual: func(es ...*types.ExternalService) ExternalServicesAssertion {
		want := append(types.ExternalServices{}, es...).With(Opt.ExternalServiceID(0))
		return func(t testing.TB, have types.ExternalServices) {
			t.Helper()
			// Exclude auto-generated IDs from equality tests
			have = append(types.ExternalServices{}, have...).With(Opt.ExternalServiceID(0))
			if !reflect.DeepEqual(have, want) {
				t.Errorf("external services (-want +got): %s", cmp.Diff(want, have))
			}
		}
	},
	ExternalServicesOrderedBy: func(ord func(a, b *types.ExternalService) bool) ExternalServicesAssertion {
		return func(t testing.TB, have types.ExternalServices) {
			t.Helper()
			want := have.Clone()
			sort.Slice(want, func(i, j int) bool {
				return ord(want[i], want[j])
			})
			if !reflect.DeepEqual(have, want) {
				t.Errorf("external services (-want +got): %s", cmp.Diff(want, have))
			}
		}
	},
}

//
// Functional options
//

// Opt contains functional options to be used in tests.
var Opt = struct {
	ExternalServiceID         func(int64) func(*types.ExternalService)
	ExternalServiceModifiedAt func(time.Time) func(*types.ExternalService)
	ExternalServiceDeletedAt  func(time.Time) func(*types.ExternalService)
	RepoID                    func(api.RepoID) func(*types.Repo)
	RepoName                  func(string) func(*types.Repo)
	RepoCreatedAt             func(time.Time) func(*types.Repo)
	RepoModifiedAt            func(time.Time) func(*types.Repo)
	RepoDeletedAt             func(time.Time) func(*types.Repo)
	RepoSources               func(...string) func(*types.Repo)
	RepoMetadata              func(interface{}) func(*types.Repo)
	RepoExternalID            func(string) func(*types.Repo)
}{
	ExternalServiceID: func(n int64) func(*types.ExternalService) {
		return func(e *types.ExternalService) {
			e.ID = n
		}
	},
	ExternalServiceModifiedAt: func(ts time.Time) func(*types.ExternalService) {
		return func(e *types.ExternalService) {
			e.UpdatedAt = ts
			e.DeletedAt = time.Time{}
		}
	},
	ExternalServiceDeletedAt: func(ts time.Time) func(*types.ExternalService) {
		return func(e *types.ExternalService) {
			e.UpdatedAt = ts
			e.DeletedAt = ts
		}
	},
	RepoID: func(n api.RepoID) func(*types.Repo) {
		return func(r *types.Repo) {
			r.ID = n
		}
	},
	RepoName: func(name string) func(*types.Repo) {
		return func(r *types.Repo) {
			r.Name = api.RepoName(name)
		}
	},
	RepoCreatedAt: func(ts time.Time) func(*types.Repo) {
		return func(r *types.Repo) {
			if r.RepoFields == nil {
				r.RepoFields = new(types.RepoFields)
			}

			r.CreatedAt = ts
			r.UpdatedAt = ts
			r.DeletedAt = time.Time{}
		}
	},
	RepoModifiedAt: func(ts time.Time) func(*types.Repo) {
		return func(r *types.Repo) {
			if r.RepoFields == nil {
				r.RepoFields = new(types.RepoFields)
			}

			r.UpdatedAt = ts
			r.DeletedAt = time.Time{}
		}
	},
	RepoDeletedAt: func(ts time.Time) func(*types.Repo) {
		return func(r *types.Repo) {
			if r.RepoFields == nil {
				r.RepoFields = new(types.RepoFields)
			}

			r.UpdatedAt = ts
			r.DeletedAt = ts
			r.Sources = map[string]*types.SourceInfo{}
		}
	},
	RepoSources: func(srcs ...string) func(*types.Repo) {
		return func(r *types.Repo) {
			if r.RepoFields == nil {
				r.RepoFields = new(types.RepoFields)
			}

			r.Sources = map[string]*types.SourceInfo{}
			for _, src := range srcs {
				r.Sources[src] = &types.SourceInfo{ID: src}
			}
		}
	},
	RepoMetadata: func(md interface{}) func(*types.Repo) {
		return func(r *types.Repo) {
			if r.RepoFields == nil {
				r.RepoFields = new(types.RepoFields)
			}

			r.Metadata = md
		}
	},
	RepoExternalID: func(id string) func(*types.Repo) {
		return func(r *types.Repo) {
			r.ExternalRepo.ID = id
		}
	},
}
