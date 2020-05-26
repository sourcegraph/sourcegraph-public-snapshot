package authz

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

func TestPermsSyncer_ScheduleUsers(t *testing.T) {
	s := NewPermsSyncer(nil, nil, nil, nil)
	s.ScheduleUsers(context.Background(), 1)

	expHeap := []*syncRequest{
		{requestMeta: &requestMeta{
			Priority: priorityHigh,
			Type:     requestTypeUser,
			ID:       1,
		}, acquired: false, index: 0},
	}
	if diff := cmp.Diff(expHeap, s.queue.heap, cmpOpts); diff != "" {
		t.Fatalf("heap: %v", diff)
	}
}

func TestPermsSyncer_ScheduleRepos(t *testing.T) {
	s := NewPermsSyncer(nil, nil, nil, nil)
	s.ScheduleRepos(context.Background(), 1)

	expHeap := []*syncRequest{
		{requestMeta: &requestMeta{
			Priority: priorityHigh,
			Type:     requestTypeRepo,
			ID:       1,
		}, acquired: false, index: 0},
	}
	if diff := cmp.Diff(expHeap, s.queue.heap, cmpOpts); diff != "" {
		t.Fatalf("heap: %v", diff)
	}
}

type mockProvider struct {
	serviceType string
	serviceID   string

	fetchUserPerms func(context.Context, *extsvc.Account) ([]extsvc.RepoID, error)
	fetchRepoPerms func(ctx context.Context, repo *extsvc.Repository) ([]extsvc.AccountID, error)
}

func (*mockProvider) RepoPerms(context.Context, *extsvc.Account, []*types.Repo) ([]authz.RepoPerms, error) {
	return nil, nil
}

func (*mockProvider) FetchAccount(context.Context, *types.User, []*extsvc.Account) (*extsvc.Account, error) {
	return nil, nil
}

func (p *mockProvider) ServiceType() string {
	return p.serviceType
}

func (p *mockProvider) ServiceID() string {
	return p.serviceID
}

func (*mockProvider) Validate() []string {
	return nil
}

func (p *mockProvider) FetchUserPerms(ctx context.Context, acct *extsvc.Account) ([]extsvc.RepoID, error) {
	return p.fetchUserPerms(ctx, acct)
}

func (p *mockProvider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository) ([]extsvc.AccountID, error) {
	return p.fetchRepoPerms(ctx, repo)
}

type mockReposStore struct {
	listRepos func(context.Context, repos.StoreListReposArgs) ([]*repos.Repo, error)
}

func (s *mockReposStore) ListExternalServices(context.Context, repos.StoreListExternalServicesArgs) ([]*repos.ExternalService, error) {
	return nil, nil
}

func (s *mockReposStore) UpsertExternalServices(context.Context, ...*repos.ExternalService) error {
	return nil
}

func (s *mockReposStore) ListRepos(ctx context.Context, args repos.StoreListReposArgs) ([]*repos.Repo, error) {
	return s.listRepos(ctx, args)
}

func (s *mockReposStore) UpsertRepos(context.Context, ...*repos.Repo) error {
	return nil
}

func (s *mockReposStore) ListAllRepoNames(context.Context) ([]api.RepoName, error) {
	return nil, nil
}

func TestPermsSyncer_syncUserPerms(t *testing.T) {
	p := &mockProvider{
		serviceType: gitlab.ServiceType,
		serviceID:   "https://gitlab.com/",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	extAccount := extsvc.Account{
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
		},
	}

	edb.Mocks.Perms.ListExternalAccounts = func(context.Context, int32) ([]*extsvc.Account, error) {
		return []*extsvc.Account{&extAccount}, nil
	}
	edb.Mocks.Perms.SetUserPermissions = func(_ context.Context, p *authz.UserPermissions) error {
		if p.UserID != 1 {
			return fmt.Errorf("UserID: want 1 but got %d", p.UserID)
		}

		expIDs := []uint32{1}
		if diff := cmp.Diff(expIDs, p.IDs.ToArray()); diff != "" {
			return fmt.Errorf("IDs: %v", diff)
		}
		return nil
	}
	defer func() {
		edb.Mocks.Perms = edb.MockPerms{}
	}()

	reposStore := &mockReposStore{
		listRepos: func(_ context.Context, args repos.StoreListReposArgs) ([]*repos.Repo, error) {
			if !args.PrivateOnly {
				return nil, errors.New("PrivateOnly want true but got false")
			}
			return []*repos.Repo{{ID: 1}}, nil
		},
	}
	clock := func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}
	permsStore := edb.NewPermsStore(nil, clock)
	s := NewPermsSyncer(reposStore, permsStore, clock, nil)
	s.metrics.syncDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{"type", "success"})
	s.metrics.syncErrors = prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"type"})

	tests := []struct {
		name     string
		noPerms  bool
		fetchErr error
	}{
		{
			name:     "sync for the first time and encounter an error",
			noPerms:  true,
			fetchErr: errors.New("random error"),
		},
		{
			name:    "sync for the second time and succeed",
			noPerms: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p.fetchUserPerms = func(context.Context, *extsvc.Account) ([]extsvc.RepoID, error) {
				return []extsvc.RepoID{"1"}, test.fetchErr
			}

			err := s.syncUserPerms(context.Background(), 1, test.noPerms)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPermsSyncer_syncRepoPerms(t *testing.T) {
	p := &mockProvider{
		serviceType: gitlab.ServiceType,
		serviceID:   "https://gitlab.com/",
	}
	authz.SetProviders(false, []authz.Provider{p})
	defer authz.SetProviders(true, nil)

	edb.Mocks.Perms.Transact = func(context.Context) (*edb.PermsStore, error) {
		return &edb.PermsStore{}, nil
	}
	edb.Mocks.Perms.GetUserIDsByExternalAccounts = func(context.Context, *extsvc.Accounts) (map[string]int32, error) {
		return map[string]int32{"user": 1}, nil
	}
	edb.Mocks.Perms.SetRepoPermissions = func(_ context.Context, p *authz.RepoPermissions) error {
		if p.RepoID != 1 {
			return fmt.Errorf("RepoID: want 1 but got %d", p.RepoID)
		}

		expUserIDs := []uint32{1}
		if diff := cmp.Diff(expUserIDs, p.UserIDs.ToArray()); diff != "" {
			return fmt.Errorf("UserIDs: %v", diff)
		}
		return nil
	}
	edb.Mocks.Perms.SetRepoPendingPermissions = func(_ context.Context, accounts *extsvc.Accounts, _ *authz.RepoPermissions) error {
		expAccounts := &extsvc.Accounts{
			ServiceType: p.ServiceType(),
			ServiceID:   p.ServiceID(),
			AccountIDs:  []string{"pending_user"},
		}
		if diff := cmp.Diff(expAccounts, accounts); diff != "" {
			return fmt.Errorf("accounts: %v", diff)
		}
		return nil
	}
	defer func() {
		edb.Mocks.Perms = edb.MockPerms{}
	}()

	reposStore := &mockReposStore{
		listRepos: func(context.Context, repos.StoreListReposArgs) ([]*repos.Repo, error) {
			return []*repos.Repo{
				{
					ID:      1,
					Private: true,
					ExternalRepo: api.ExternalRepoSpec{
						ServiceID: p.ServiceID(),
					},
				},
			}, nil
		},
	}
	clock := func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}
	permsStore := edb.NewPermsStore(nil, clock)
	s := NewPermsSyncer(reposStore, permsStore, clock, nil)
	s.metrics.syncDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{}, []string{"type", "success"})
	s.metrics.syncErrors = prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"type"})

	tests := []struct {
		name     string
		noPerms  bool
		fetchErr error
	}{
		{
			name:     "sync for the first time and encounter an error",
			noPerms:  true,
			fetchErr: errors.New("random error"),
		},
		{
			name:    "sync for the second time and succeed",
			noPerms: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p.fetchRepoPerms = func(context.Context, *extsvc.Repository) ([]extsvc.AccountID, error) {
				return []extsvc.AccountID{"user", "pending_user"}, test.fetchErr
			}

			err := s.syncRepoPerms(context.Background(), 1, test.noPerms)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

type fakeExternalServiceLister struct{}

func (*fakeExternalServiceLister) ListExternalServices(context.Context, repos.StoreListExternalServicesArgs) ([]*repos.ExternalService, error) {
	return []*repos.ExternalService{
		{
			ID:          1,
			Kind:        "GITHUB",
			DisplayName: "GitHub.com",
			Config:      `{"url": "https://github.com"}`,
		},
	}, nil
}

func TestPermsSyncer_waitForRateLimit(t *testing.T) {
	ctx := context.Background()
	t.Run("no rate limit registry", func(t *testing.T) {
		s := NewPermsSyncer(nil, nil, nil, nil)

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err := s.waitForRateLimit(ctx, "https://github.com/", 100000)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("enough quota available", func(t *testing.T) {
		rateLimiterRegistry, err := repos.NewRateLimiterRegistry(ctx, &fakeExternalServiceLister{})
		if err != nil {
			t.Fatal(err)
		}
		s := NewPermsSyncer(nil, nil, nil, rateLimiterRegistry)

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err = s.waitForRateLimit(ctx, "https://github.com/", 1)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("not enough quota available", func(t *testing.T) {
		rateLimiterRegistry, err := repos.NewRateLimiterRegistry(ctx, &fakeExternalServiceLister{})
		if err != nil {
			t.Fatal(err)
		}
		s := NewPermsSyncer(nil, nil, nil, rateLimiterRegistry)

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		err = s.waitForRateLimit(ctx, "https://github.com/", 10)
		if err == nil {
			t.Fatalf("err: want %v but got nil", context.Canceled)
		}
	})
}

func TestPermsSyncer_syncPerms(t *testing.T) {
	request := &syncRequest{
		requestMeta: &requestMeta{
			Type: 3,
			ID:   1,
		},
		acquired: true,
	}

	// Request should be removed from the queue even if error occurred.
	s := NewPermsSyncer(nil, nil, nil, nil)
	s.queue.Push(request)

	expErr := "unexpected request type: 3"
	err := s.syncPerms(context.Background(), request)
	if err == nil || err.Error() != expErr {
		t.Fatalf("err: want %q but got %v", expErr, err)
	}

	if s.queue.Len() != 0 {
		t.Fatalf("queue length: want 0 but got %d", s.queue.Len())
	}
}
