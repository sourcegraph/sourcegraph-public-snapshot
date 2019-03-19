package repos

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// NewFakeSourcer returns a Sourcer which always returns the given error and sources,
// ignoring the given external services.
func NewFakeSourcer(err error, srcs ...Source) Sourcer {
	return func(...*ExternalService) (Sources, error) {
		return srcs, err
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
func (s FakeSource) ListRepos(context.Context) ([]*Repo, error) {
	repos := make([]*Repo, len(s.repos))
	for i, r := range s.repos {
		repos[i] = r.With(Opt.RepoSources(s.svc.URN()))
	}
	return repos, s.err
}

// FakeStore is a fake implementation of Store to be used in tests.
type FakeStore struct {
	ListExternalServicesError   error // error to be returned in ListExternalServices
	UpsertExternalServicesError error // error to be returned in UpsertExternalServices
	GetRepoByNameError          error // error to be returned in GetRepoByName
	ListReposError              error // error to be returned in ListRepos
	UpsertReposError            error // error to be returned in UpsertRepos

	svcIDSeq   int64
	repoIDSeq  uint32
	svcByID    map[int64]*ExternalService
	repoByName map[string]*Repo
	repoByID   map[api.ExternalRepoSpec]*Repo
}

// Transact returns a TxStore whose methods operate within the context of a transaction.
func (s *FakeStore) Transact(ctx context.Context) (TxStore, error) {
	svcByID := make(map[int64]*ExternalService, len(s.svcByID))
	for id, svc := range s.svcByID {
		svcByID[id] = svc.Clone()
	}

	repoByName := make(map[string]*Repo, len(s.repoByName))
	repoByID := make(map[api.ExternalRepoSpec]*Repo, len(s.repoByID))
	for name, r := range s.repoByName {
		clone := r.Clone()
		repoByName[name] = clone
		repoByID[r.ExternalRepo] = clone
	}

	return &FakeStore{
		ListExternalServicesError:   s.ListExternalServicesError,
		UpsertExternalServicesError: s.UpsertExternalServicesError,
		GetRepoByNameError:          s.GetRepoByNameError,
		ListReposError:              s.ListReposError,
		UpsertReposError:            s.UpsertReposError,

		svcIDSeq:   s.svcIDSeq,
		svcByID:    svcByID,
		repoIDSeq:  s.repoIDSeq,
		repoByName: repoByName,
		repoByID:   repoByID,
	}, nil
}

// Done fakes the implementation of a TxStore's Done method by always discarding all state
// changes made during the transaction.
func (s *FakeStore) Done(...*error) {}

// ListExternalServices lists all stored external services that are not deleted and have one of the
// specified kinds.
func (s FakeStore) ListExternalServices(ctx context.Context, kinds ...string) ([]*ExternalService, error) {
	if s.ListExternalServicesError != nil {
		return nil, s.ListExternalServicesError
	}

	if s.svcByID == nil {
		s.svcByID = make(map[int64]*ExternalService)
	}

	kindset := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		kindset[strings.ToLower(kind)] = true
	}

	svcs := make(ExternalServices, 0, len(s.svcByID))
	for _, svc := range s.svcByID {
		if len(kinds) == 0 || kindset[strings.ToLower(svc.Kind)] {
			svcs = append(svcs, svc)
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

	if s.repoByName == nil {
		s.repoByName = make(map[string]*Repo)
	}

	r := s.repoByName[name]
	if r == nil || !r.DeletedAt.IsZero() {
		return nil, ErrNoResults
	}

	return r, nil
}

// ListRepos lists all repos in the store that have one of the specified external service kinds.
func (s FakeStore) ListRepos(ctx context.Context, kinds ...string) ([]*Repo, error) {
	if s.ListReposError != nil {
		return nil, s.ListReposError
	}

	if s.repoByName == nil {
		s.repoByName = make(map[string]*Repo)
	}

	kindset := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		kindset[strings.ToLower(kind)] = true
	}

	set := make(map[*Repo]bool, len(s.repoByName))
	repos := make(Repos, 0, len(s.repoByName))
	for _, r := range s.repoByName {
		if !set[r] && len(kinds) == 0 || kindset[strings.ToLower(r.ExternalRepo.ServiceType)] {
			repos = append(repos, r)
			set[r] = true
		}
	}

	sort.Sort(repos)

	return repos, nil
}

// UpsertRepos upserts all the given repos in the store.
func (s *FakeStore) UpsertRepos(ctx context.Context, upserts ...*Repo) error {
	if s.UpsertReposError != nil {
		return s.UpsertReposError
	}

	if s.repoByName == nil {
		s.repoByName = make(map[string]*Repo, len(upserts))
	}

	if s.repoByID == nil {
		s.repoByID = make(map[api.ExternalRepoSpec]*Repo, len(upserts))
	}

	for _, upsert := range upserts {
		if repo := s.repoByID[upsert.ExternalRepo]; repo != nil {
			repo.Update(upsert)
		} else if repo = s.repoByName[upsert.Name]; repo != nil {
			repo.Update(upsert)
		} else {
			s.repoIDSeq++
			upsert.ID = s.repoIDSeq
			s.repoByName[upsert.Name] = upsert
			if upsert.ExternalRepo != (api.ExternalRepoSpec{}) {
				s.repoByID[upsert.ExternalRepo] = upsert
			}
		}
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
		want := Repos(rs)
		return func(t testing.TB, have Repos) {
			have.Apply(Opt.RepoID(0)) // Exclude auto-generated IDs from equality tests
			if !reflect.DeepEqual(have, want) {
				t.Errorf("repos: %s", cmp.Diff(have, want))
			}
		}
	},
	ReposOrderedBy: func(ord func(a, b *Repo) bool) ReposAssertion {
		return func(t testing.TB, have Repos) {
			want := have.Clone()
			sort.Slice(want, func(i, j int) bool {
				return ord(want[i], want[j])
			})
			if !reflect.DeepEqual(have, want) {
				t.Errorf("repos: %s", cmp.Diff(have, want))
			}
		}
	},
	ExternalServicesEqual: func(es ...*ExternalService) ExternalServicesAssertion {
		want := ExternalServices(es)
		return func(t testing.TB, have ExternalServices) {
			have.Apply(Opt.ExternalServiceID(0)) // Exclude auto-generated IDs from equality tests
			if !reflect.DeepEqual(have, want) {
				t.Errorf("external services: %s", cmp.Diff(have, want))
			}
		}
	},
	ExternalServicesOrderedBy: func(ord func(a, b *ExternalService) bool) ExternalServicesAssertion {
		return func(t testing.TB, have ExternalServices) {
			want := have.Clone()
			sort.Slice(want, func(i, j int) bool {
				return ord(want[i], want[j])
			})
			if !reflect.DeepEqual(have, want) {
				t.Errorf("external services: %s", cmp.Diff(have, want))
			}
		}
	},
}

//
// Functional options
//

// Opt contains functional options to be used in tests.
var Opt = struct {
	ExternalServiceID        func(int64) func(*ExternalService)
	ExternalServiceDeletedAt func(time.Time) func(*ExternalService)
	RepoID                   func(uint32) func(*Repo)
	RepoName                 func(string) func(*Repo)
	RepoCreatedAt            func(time.Time) func(*Repo)
	RepoModifiedAt           func(time.Time) func(*Repo)
	RepoDeletedAt            func(time.Time) func(*Repo)
	RepoSources              func(...string) func(*Repo)
	RepoMetadata             func(interface{}) func(*Repo)
	RepoExternalID           func(string) func(*Repo)
}{
	ExternalServiceID: func(n int64) func(*ExternalService) {
		return func(e *ExternalService) {
			e.ID = n
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
			r.Enabled = false
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

// FakeInternalAPI implements the InternalAPI interface with the given in memory data.
// It's safe for concurrent use.
type FakeInternalAPI struct {
	mu     sync.RWMutex
	svcs   map[string][]*api.ExternalService
	repos  map[api.RepoName]*api.Repo
	repoID api.RepoID
}

// NewFakeInternalAPI returns a new FakeInternalAPI initialised with the given data.
func NewFakeInternalAPI(svcs []*api.ExternalService, repos []*api.Repo) *FakeInternalAPI {
	fa := FakeInternalAPI{
		svcs:  map[string][]*api.ExternalService{},
		repos: map[api.RepoName]*api.Repo{},
	}

	for _, svc := range svcs {
		fa.svcs[svc.Kind] = append(fa.svcs[svc.Kind], svc)
	}

	for _, repo := range repos {
		fa.repos[repo.Name] = repo
	}

	return &fa
}

// ExternalServicesList lists external services of the given Kind. A non-existent kind
// will result in an error being returned.
func (a *FakeInternalAPI) ExternalServicesList(
	_ context.Context,
	req api.ExternalServicesListRequest,
) ([]*api.ExternalService, error) {

	a.mu.RLock()
	defer a.mu.RUnlock()

	var svcs []*api.ExternalService
	for _, kind := range req.Kinds {
		svcs = append(svcs, a.svcs[kind]...)
	}

	sort.Slice(svcs, func(i, j int) bool { return svcs[i].ID < svcs[j].ID })

	return svcs, nil
}

// ReposList returns the list of all repos in the API.
func (a *FakeInternalAPI) ReposList() []*api.Repo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	repos := make([]*api.Repo, 0, len(a.repos))
	for _, repo := range a.repos {
		repos = append(repos, repo)
	}

	return repos
}

// ReposCreateIfNotExists creates the given repo if it doesn't exist. Repos with
// with an empty name are invalid and will result in an error to be returned.
func (a *FakeInternalAPI) ReposCreateIfNotExists(
	ctx context.Context,
	req api.RepoCreateOrUpdateRequest,
) (*api.Repo, error) {

	if req.RepoName == "" {
		return nil, errors.New("invalid empty repo name")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	repo, ok := a.repos[req.RepoName]
	if !ok {
		a.repoID++
		repo = &api.Repo{
			ID:           a.repoID,
			Name:         req.RepoName,
			Enabled:      req.Enabled,
			ExternalRepo: req.ExternalRepo,
		}
		a.repos[req.RepoName] = repo
	}

	return repo, nil
}

// ReposUpdateMetadata updates the metadata of repo with the given name.
// Non-existent repos return an error.
func (a *FakeInternalAPI) ReposUpdateMetadata(
	ctx context.Context,
	repoName api.RepoName,
	description string,
	fork, archived bool,
) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.repos[repoName]
	if !ok {
		return fmt.Errorf("repo %q not found", repoName)
	}

	// Fake success, no updates needed because the returned types (api.Repo)
	// don't include these fields.

	return nil
}
