package repos

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// FakeSourcer is a fake implementation of Sourcer to be used in tests.
type FakeSourcer struct {
	err  error
	srcs []Source
}

// NewFakeSourcer returns an instance of FakeSourcer with the given error
// and sources
func NewFakeSourcer(err error, srcs ...Source) *FakeSourcer {
	return &FakeSourcer{err: err, srcs: srcs}
}

// ListSources returns the Sources that FakeSourcer was instantiated with that have one
// of the given kinds as well the error, if any.
func (s FakeSourcer) ListSources(_ context.Context, kinds ...string) (srcs []Source, err error) {
	if s.err != nil {
		return nil, s.err
	}

	kindset := make(map[string]bool, len(kinds))
	for _, k := range kinds {
		kindset[k] = true
	}

	for _, src := range s.srcs {
		switch s := src.(type) {
		case *FakeSource:
			if kindset[s.kind] {
				srcs = append(srcs, s)
			}
		default:
			panic(fmt.Errorf("FakeSourcer not compatible with %#v yet", s))
		}
	}

	return srcs, s.err
}

// FakeSource is a fake implementation of Source to be used in tests.
type FakeSource struct {
	urn   string
	kind  string
	repos []*Repo
	err   error
}

// NewFakeSource returns an instance of FakeSource with the given urn, error
// and repos.
func NewFakeSource(urn, kind string, err error, rs ...*Repo) *FakeSource {
	return &FakeSource{urn: urn, kind: kind, err: err, repos: rs}
}

// ListRepos returns the Repos that FakeSource was instantiated with
// as well as the error, if any.
func (s FakeSource) ListRepos(context.Context) ([]*Repo, error) {
	repos := make([]*Repo, len(s.repos))
	for i, r := range s.repos {
		repos[i] = r.With(Opt.Sources(s.urn))
	}
	return repos, s.err
}

// FakeStore is a fake implementation of Store to be used in tests.
type FakeStore struct {
	byName map[string]*Repo
	byID   map[api.ExternalRepoSpec]*Repo
	get    error // error to be returned in GetRepoByName
	list   error // error to be returned in ListRepos
	upsert error // error to be returned in UpsertRepos
}

// NewFakeStore returns an instance of FakeStore with the given urn, error
// and repos.
func NewFakeStore(get, list, upsert error, rs ...*Repo) *FakeStore {
	s := FakeStore{
		byName: map[string]*Repo{},
		byID:   map[api.ExternalRepoSpec]*Repo{},
		get:    get,
		list:   list,
		upsert: upsert,
	}
	_ = s.UpsertRepos(context.Background(), rs...)
	return &s
}

// GetRepoByName looks a repo by its name, returning it if found.
func (s FakeStore) GetRepoByName(ctx context.Context, name string) (*Repo, error) {
	if s.get != nil {
		return nil, s.get
	}

	r := s.byName[name]
	if r == nil || !r.DeletedAt.IsZero() {
		return nil, ErrNoResults
	}

	return r, nil
}

// ListRepos lists all repos in the store that have one of the specified external service kinds.
func (s FakeStore) ListRepos(ctx context.Context, kinds ...string) ([]*Repo, error) {
	if s.list != nil {
		return nil, s.list
	}

	kindset := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		kindset[kind] = true
	}

	set := make(map[*Repo]bool, len(s.byName))
	repos := make(Repos, 0, len(s.byName))
	for _, r := range s.byName {
		if !set[r] && len(kinds) == 0 || kindset[r.ExternalRepo.ServiceType] {
			repos = append(repos, r)
			set[r] = true
		}
	}

	sort.Sort(repos)

	return repos, nil
}

// UpsertRepos upserts all the given repos in the store.
func (s *FakeStore) UpsertRepos(ctx context.Context, upserts ...*Repo) error {
	if s.upsert != nil {
		return s.upsert
	}

	if s.byName == nil {
		s.byName = make(map[string]*Repo, len(upserts))
	}

	if s.byID == nil {
		s.byID = make(map[api.ExternalRepoSpec]*Repo, len(upserts))
	}

	for _, upsert := range upserts {
		if repo := s.byID[upsert.ExternalRepo]; repo != nil {
			repo.Update(upsert)
		} else if repo = s.byName[upsert.Name]; repo != nil {
			repo.Update(upsert)
		} else {
			s.byName[upsert.Name] = upsert
			if upsert.ExternalRepo != (api.ExternalRepoSpec{}) {
				s.byID[upsert.ExternalRepo] = upsert
			}
		}
	}

	return nil
}

//
// Repo functional options
//

// Opt contains functional options for Repos to be used in tests.
var Opt = struct {
	ID         func(uint32) func(*Repo)
	CreatedAt  func(time.Time) func(*Repo)
	ModifiedAt func(time.Time) func(*Repo)
	DeletedAt  func(time.Time) func(*Repo)
	Sources    func(...string) func(*Repo)
	Metadata   func(interface{}) func(*Repo)
	ExternalID func(string) func(*Repo)
}{
	ID: func(n uint32) func(*Repo) {
		return func(r *Repo) {
			r.ID = n
		}
	},
	CreatedAt: func(ts time.Time) func(*Repo) {
		return func(r *Repo) {
			r.CreatedAt = ts
			r.DeletedAt = time.Time{}
		}
	},
	ModifiedAt: func(ts time.Time) func(*Repo) {
		return func(r *Repo) {
			r.UpdatedAt = ts
			r.DeletedAt = time.Time{}
		}
	},
	DeletedAt: func(ts time.Time) func(*Repo) {
		return func(r *Repo) {
			r.UpdatedAt = ts
			r.DeletedAt = ts
			r.Sources = map[string]*SourceInfo{}
		}
	},
	Sources: func(srcs ...string) func(*Repo) {
		return func(r *Repo) {
			r.Sources = map[string]*SourceInfo{}
			for _, src := range srcs {
				r.Sources[src] = &SourceInfo{ID: src}
			}
		}
	},
	Metadata: func(md interface{}) func(*Repo) {
		return func(r *Repo) {
			r.Metadata = md
		}
	},
	ExternalID: func(id string) func(*Repo) {
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
	return c.epoch.Add(time.Duration(step) * c.step).Truncate(time.Microsecond).UTC()
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

// ReposUpdateMetdata updates the metadata of repo with the given name.
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
