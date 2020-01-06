package repos

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// NewFakeSourcer returns a Sourcer which always returns the given error and sources,
// ignoring the given external services.
func NewFakeSourcer(err error, srcs ...Source) Sourcer {
	return func(svcs ...*ExternalService) (Sources, error) {
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
	svc   *ExternalService
	repos []*Repo
	err   error
}

// NewFakeSource returns an instance of FakeSource with the given urn, error
// and repos.
func NewFakeSource(svc *ExternalService, err error, rs ...*Repo) *FakeSource {
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
func (s FakeSource) ExternalServices() ExternalServices {
	return ExternalServices{s.svc}
}

// FakeStore is a fake implementation of Store to be used in tests.
type FakeStore struct {
	ListExternalServicesError   error // error to be returned in ListExternalServices
	UpsertExternalServicesError error // error to be returned in UpsertExternalServices
	GetRepoByNameError          error // error to be returned in GetRepoByName
	ListReposError              error // error to be returned in ListRepos
	UpsertReposError            error // error to be returned in UpsertRepos
	ListAllRepoNamesError       error // error to be returned in ListAllRepoNames

	svcIDSeq  int64
	repoIDSeq uint32
	svcByID   map[int64]*ExternalService
	repoByID  map[uint32]*Repo
	parent    *FakeStore
}

// Transact returns a TxStore whose methods operate within the context of a transaction.
func (s *FakeStore) Transact(ctx context.Context) (TxStore, error) {
	if s.parent != nil {
		return nil, errors.New("already in transaction")
	}

	svcByID := make(map[int64]*ExternalService, len(s.svcByID))
	for id, svc := range s.svcByID {
		svcByID[id] = svc.Clone()
	}

	repoByID := make(map[uint32]*Repo, len(s.repoByID))
	for _, r := range s.repoByID {
		clone := r.Clone()
		repoByID[r.ID] = clone
	}

	return &FakeStore{
		ListExternalServicesError:   s.ListExternalServicesError,
		UpsertExternalServicesError: s.UpsertExternalServicesError,
		GetRepoByNameError:          s.GetRepoByNameError,
		ListReposError:              s.ListReposError,
		UpsertReposError:            s.UpsertReposError,
		ListAllRepoNamesError:       s.ListAllRepoNamesError,

		svcIDSeq:  s.svcIDSeq,
		svcByID:   svcByID,
		repoIDSeq: s.repoIDSeq,
		repoByID:  repoByID,
		parent:    s,
	}, nil
}

// Done fakes the implementation of a TxStore's Done method by always discarding all state
// changes made during the transaction.
func (s *FakeStore) Done(e ...*error) {
	if len(e) > 0 && e[0] != nil && *e[0] != nil {
		return
	}

	// Transaction succeeded. Copy maps into parent
	p := s.parent
	s.parent = nil
	*p = *s
}

// ListExternalServices lists all stored external services that match the given args.
func (s FakeStore) ListExternalServices(ctx context.Context, args StoreListExternalServicesArgs) ([]*ExternalService, error) {
	if s.ListExternalServicesError != nil {
		return nil, s.ListExternalServicesError
	}

	if s.svcByID == nil {
		s.svcByID = make(map[int64]*ExternalService)
	}

	kinds := make(map[string]bool, len(args.Kinds))
	for _, kind := range args.Kinds {
		kinds[strings.ToLower(kind)] = true
	}

	ids := make(map[int64]bool, len(args.IDs))
	for _, id := range args.IDs {
		ids[id] = true
	}

	set := make(map[*ExternalService]bool, len(s.svcByID))
	svcs := make(ExternalServices, 0, len(s.svcByID))
	for _, svc := range s.svcByID {
		k := strings.ToLower(svc.Kind)

		if !set[svc] &&
			((len(kinds) == 0 && k != "phabricator") || kinds[k]) &&
			(len(ids) == 0 || ids[svc.ID]) &&
			!svc.IsDeleted() {

			svcs = append(svcs, svc)
			set[svc] = true
		}
	}

	sort.Sort(svcs)

	return svcs, nil
}

// UpsertExternalServices updates or inserts the given ExternalServices.
func (s *FakeStore) UpsertExternalServices(ctx context.Context, svcs ...*ExternalService) error {
	if s.UpsertExternalServicesError != nil {
		return s.UpsertExternalServicesError
	}

	if s.svcByID == nil {
		s.svcByID = make(map[int64]*ExternalService, len(svcs))
	}

	for _, svc := range svcs {
		if old := s.svcByID[svc.ID]; old != nil {
			old.Update(svc)
		} else {
			s.svcIDSeq++
			svc.ID = s.svcIDSeq
			svc.Kind = strings.ToUpper(svc.Kind)
			s.svcByID[svc.ID] = svc
		}
	}

	return nil
}

// GetRepoByName looks a repo by its name, returning it if found.
func (s FakeStore) GetRepoByName(ctx context.Context, name string) (*Repo, error) {
	if s.GetRepoByNameError != nil {
		return nil, s.GetRepoByNameError
	}

	for _, r := range s.repoByID {
		if r.Name == name {
			return r, nil
		}
	}

	return nil, ErrNoResults
}

// ListRepos lists all repos in the store that match the given arguments.
func (s FakeStore) ListRepos(ctx context.Context, args StoreListReposArgs) ([]*Repo, error) {
	if s.ListReposError != nil {
		return nil, s.ListReposError
	}

	kinds := make(map[string]bool, len(args.Kinds))
	for _, kind := range args.Kinds {
		kinds[strings.ToLower(kind)] = true
	}

	names := make(map[string]bool, len(args.Names))
	for _, name := range args.Names {
		names[strings.ToLower(name)] = true
	}

	ids := make(map[uint32]bool, len(args.IDs))
	for _, id := range args.IDs {
		ids[id] = true
	}

	externalRepos := make(map[api.ExternalRepoSpec]bool, len(args.ExternalRepos))
	for _, spec := range args.ExternalRepos {
		externalRepos[spec] = true
	}

	set := make(map[*Repo]bool, len(s.repoByID))
	repos := make(Repos, 0, len(s.repoByID))
	for _, r := range s.repoByID {
		if set[r] {
			continue
		}

		var preds []bool
		if len(kinds) > 0 {
			preds = append(preds, kinds[strings.ToLower(r.ExternalRepo.ServiceType)])
		}
		if len(names) > 0 {
			preds = append(preds, names[strings.ToLower(r.Name)])
		}
		if len(ids) > 0 {
			preds = append(preds, ids[r.ID])
		}
		if len(externalRepos) > 0 {
			preds = append(preds, externalRepos[r.ExternalRepo])
		}

		if (args.UseOr && evalOr(preds...)) || (!args.UseOr && evalAnd(preds...)) {
			repos = append(repos, r)
			set[r] = true
		}

	}

	sort.Sort(repos)

	if args.Limit > 0 && args.Limit <= int64(len(repos)) {
		repos = repos[:args.Limit]
	}

	return repos, nil
}

// ListAllRepoNames lists names of all repos in the store
func (s FakeStore) ListAllRepoNames(ctx context.Context) ([]api.RepoName, error) {
	if s.ListAllRepoNamesError != nil {
		return nil, s.ListAllRepoNamesError
	}

	names := make([]api.RepoName, 0, len(s.repoByID))
	for _, r := range s.repoByID {
		names = append(names, api.RepoName(r.Name))
	}

	return names, nil
}

func evalOr(bs ...bool) bool {
	if len(bs) == 0 {
		return true
	}
	for _, b := range bs {
		if b {
			return true
		}
	}
	return false
}

func evalAnd(bs ...bool) bool {
	for _, b := range bs {
		if !b {
			return false
		}
	}
	return true
}

// UpsertRepos upserts all the given repos in the store.
func (s *FakeStore) UpsertRepos(ctx context.Context, upserts ...*Repo) error {
	if s.UpsertReposError != nil {
		return s.UpsertReposError
	}

	if s.repoByID == nil {
		s.repoByID = make(map[uint32]*Repo, len(upserts))
	}

	var deletes, updates, inserts []*Repo
	for _, r := range upserts {
		switch {
		case r.IsDeleted():
			deletes = append(deletes, r)
		case r.ID != 0:
			updates = append(updates, r)
		default:
			inserts = append(inserts, r)
		}
	}

	for _, r := range deletes {
		delete(s.repoByID, r.ID)
	}

	for _, r := range updates {
		repo := s.repoByID[r.ID]
		if repo == nil {
			return errors.Errorf("upserting repo with non-existant ID: id=%v", r.ID)
		}
		repo.Update(r)
	}

	for _, r := range inserts {
		s.repoIDSeq++
		r.ID = s.repoIDSeq
		s.repoByID[r.ID] = r
	}

	return s.checkConstraints()
}

// checkConstraints ensures the FakeStore has not violated any constraints we
// maintain on our DB.
//
// Constraints:
// - name is unique case insensitively
// - external repo is unique if set
func (s *FakeStore) checkConstraints() error {
	seenName := map[string]bool{}
	seenExternalRepo := map[api.ExternalRepoSpec]bool{}
	for _, r := range s.repoByID {
		name := strings.ToLower(r.Name)
		if seenName[name] {
			return errors.Errorf("duplicate repo name: %s", name)
		}
		seenName[name] = true
		if r.ExternalRepo.IsSet() && seenExternalRepo[r.ExternalRepo] {
			return errors.Errorf("duplicate external repo spec: %v", r.ExternalRepo)
		}
		seenExternalRepo[r.ExternalRepo] = true
	}
	return nil
}

//
// Assertions
//

// A ReposAssertion performs an assertion on the given Repos.
type ReposAssertion func(testing.TB, Repos)

// An ExternalServicesAssertion performs an assertion on the given
// ExternalServices.
type ExternalServicesAssertion func(testing.TB, ExternalServices)

// Assert contains assertion functions to be used in tests.
var Assert = struct {
	ReposEqual                func(...*Repo) ReposAssertion
	ReposOrderedBy            func(func(a, b *Repo) bool) ReposAssertion
	ExternalServicesEqual     func(...*ExternalService) ExternalServicesAssertion
	ExternalServicesOrderedBy func(func(a, b *ExternalService) bool) ExternalServicesAssertion
}{
	ReposEqual: func(rs ...*Repo) ReposAssertion {
		want := append(Repos{}, rs...).With(Opt.RepoID(0))
		return func(t testing.TB, have Repos) {
			t.Helper()
			// Exclude auto-generated IDs from equality tests
			have = append(Repos{}, have...).With(Opt.RepoID(0))
			if !reflect.DeepEqual(have, want) {
				t.Errorf("repos (-want +got): %s", cmp.Diff(want, have))
			}
		}
	},
	ReposOrderedBy: func(ord func(a, b *Repo) bool) ReposAssertion {
		return func(t testing.TB, have Repos) {
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
	ExternalServicesEqual: func(es ...*ExternalService) ExternalServicesAssertion {
		want := append(ExternalServices{}, es...).With(Opt.ExternalServiceID(0))
		return func(t testing.TB, have ExternalServices) {
			t.Helper()
			// Exclude auto-generated IDs from equality tests
			have = append(ExternalServices{}, have...).With(Opt.ExternalServiceID(0))
			if !reflect.DeepEqual(have, want) {
				t.Errorf("external services (-want +got): %s", cmp.Diff(want, have))
			}
		}
	},
	ExternalServicesOrderedBy: func(ord func(a, b *ExternalService) bool) ExternalServicesAssertion {
		return func(t testing.TB, have ExternalServices) {
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
	ExternalServiceID         func(int64) func(*ExternalService)
	ExternalServiceModifiedAt func(time.Time) func(*ExternalService)
	ExternalServiceDeletedAt  func(time.Time) func(*ExternalService)
	RepoID                    func(uint32) func(*Repo)
	RepoName                  func(string) func(*Repo)
	RepoCreatedAt             func(time.Time) func(*Repo)
	RepoModifiedAt            func(time.Time) func(*Repo)
	RepoDeletedAt             func(time.Time) func(*Repo)
	RepoEnabled               func(bool) func(*Repo)
	RepoSources               func(...string) func(*Repo)
	RepoMetadata              func(interface{}) func(*Repo)
	RepoExternalID            func(string) func(*Repo)
}{
	ExternalServiceID: func(n int64) func(*ExternalService) {
		return func(e *ExternalService) {
			e.ID = n
		}
	},
	ExternalServiceModifiedAt: func(ts time.Time) func(*ExternalService) {
		return func(e *ExternalService) {
			e.UpdatedAt = ts
			e.DeletedAt = time.Time{}
		}
	},
	ExternalServiceDeletedAt: func(ts time.Time) func(*ExternalService) {
		return func(e *ExternalService) {
			e.UpdatedAt = ts
			e.DeletedAt = ts
		}
	},
	RepoID: func(n uint32) func(*Repo) {
		return func(r *Repo) {
			r.ID = n
		}
	},
	RepoName: func(name string) func(*Repo) {
		return func(r *Repo) {
			r.Name = name
		}
	},
	RepoCreatedAt: func(ts time.Time) func(*Repo) {
		return func(r *Repo) {
			r.CreatedAt = ts
			r.UpdatedAt = ts
			r.DeletedAt = time.Time{}
			r.Enabled = true
		}
	},
	RepoModifiedAt: func(ts time.Time) func(*Repo) {
		return func(r *Repo) {
			r.UpdatedAt = ts
			r.DeletedAt = time.Time{}
			r.Enabled = true
		}
	},
	RepoDeletedAt: func(ts time.Time) func(*Repo) {
		return func(r *Repo) {
			r.UpdatedAt = ts
			r.DeletedAt = ts
			r.Sources = map[string]*SourceInfo{}
		}
	},
	RepoEnabled: func(enabled bool) func(*Repo) {
		return func(r *Repo) {
			r.Enabled = enabled
		}
	},
	RepoSources: func(srcs ...string) func(*Repo) {
		return func(r *Repo) {
			r.Sources = map[string]*SourceInfo{}
			for _, src := range srcs {
				r.Sources[src] = &SourceInfo{ID: src}
			}
		}
	},
	RepoMetadata: func(md interface{}) func(*Repo) {
		return func(r *Repo) {
			r.Metadata = md
		}
	},
	RepoExternalID: func(id string) func(*Repo) {
		return func(r *Repo) {
			r.ExternalRepo.ID = id
		}
	},
}

// FakeClock provides a controllable clock for use in tests.
type FakeClock struct {
	epoch time.Time
	step  time.Duration
	steps int
}

// NewFakeClock returns a FakeClock instance that starts telling time at the given epoch.
// Every invocation of Now adds step amount of time to the clock.
func NewFakeClock(epoch time.Time, step time.Duration) FakeClock {
	return FakeClock{epoch: epoch, step: step}
}

// Now returns the current fake time and advances the clock "step" amount of time.
func (c *FakeClock) Now() time.Time {
	c.steps++
	return c.Time(c.steps)
}

// Time tells the time at the given step from the provided epoch.
func (c FakeClock) Time(step int) time.Time {
	// We truncate to microsecond precision because Postgres' timestamptz type
	// doesn't handle nanoseconds.
	return c.epoch.Add(time.Duration(step) * c.step).UTC().Truncate(time.Microsecond)
}
