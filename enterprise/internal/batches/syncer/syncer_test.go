package syncer

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newTestStore() *MockSyncStore {
	s := NewMockSyncStore()
	s.ClockFunc.SetDefaultReturn(timeutil.Now)
	return s
}

func TestSyncerRun(t *testing.T) {
	t.Parallel()

	t.Run("Sync due", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		now := time.Now()

		syncStore := newTestStore()
		syncStore.ListChangesetSyncDataFunc.SetDefaultReturn([]*btypes.ChangesetSyncData{
			{
				ChangesetID:       1,
				UpdatedAt:         now.Add(-2 * maxSyncDelay),
				LatestEvent:       now.Add(-2 * maxSyncDelay),
				ExternalUpdatedAt: now.Add(-2 * maxSyncDelay),
			},
		}, nil)

		syncFunc := func(ctx context.Context, ids int64) error {
			cancel()
			return nil
		}
		syncer := &changesetSyncer{
			syncStore:        syncStore,
			scheduleInterval: 10 * time.Minute,
			syncFunc:         syncFunc,
			metrics:          makeMetrics(&observation.TestContext),
		}
		go syncer.Run(ctx)
		select {
		case <-ctx.Done():
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Sync should have been triggered")
		}
	})

	t.Run("Sync due but reenqueued for reconciler", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		now := time.Now()
		updateCalled := false
		syncStore := newTestStore()
		// Return ErrNoResults, which is the result you get when the changeset preconditions aren't met anymore.
		// The sync data checks for the reconciler state and if it changed since the sync data was loaded,
		// we don't get back the changeset here and skip it.
		//
		// If we don't return ErrNoResults, the rest of the test will fail, because not all
		// methods of sync store are mocked.
		syncStore.GetChangesetFunc.SetDefaultReturn(nil, store.ErrNoResults)
		syncStore.UpdateChangesetCodeHostStateFunc.SetDefaultHook(func(context.Context, *btypes.Changeset) error {
			updateCalled = true
			return nil
		})
		syncStore.ListChangesetSyncDataFunc.SetDefaultReturn([]*btypes.ChangesetSyncData{
			{
				ChangesetID:       1,
				UpdatedAt:         now.Add(-2 * maxSyncDelay),
				LatestEvent:       now.Add(-2 * maxSyncDelay),
				ExternalUpdatedAt: now.Add(-2 * maxSyncDelay),
			},
		}, nil)

		syncer := &changesetSyncer{
			syncStore:        syncStore,
			scheduleInterval: 10 * time.Minute,
			metrics:          makeMetrics(&observation.TestContext),
		}
		syncer.Run(ctx)
		if updateCalled {
			t.Fatal("Called UpdateChangeset, but shouldn't have")
		}
	})

	t.Run("Sync not due", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		now := time.Now()
		syncStore := newTestStore()
		syncStore.ListChangesetSyncDataFunc.SetDefaultReturn([]*btypes.ChangesetSyncData{
			{
				ChangesetID:       1,
				UpdatedAt:         now,
				LatestEvent:       now,
				ExternalUpdatedAt: now,
			},
		}, nil)

		var syncCalled bool
		syncFunc := func(ctx context.Context, ids int64) error {
			syncCalled = true
			return nil
		}
		syncer := &changesetSyncer{
			syncStore:        syncStore,
			scheduleInterval: 10 * time.Minute,
			syncFunc:         syncFunc,
			metrics:          makeMetrics(&observation.TestContext),
		}
		syncer.Run(ctx)
		if syncCalled {
			t.Fatal("Sync should not have been triggered")
		}
	})

	t.Run("Priority added", func(t *testing.T) {
		// Empty schedule but then we add an item
		ctx, cancel := context.WithCancel(context.Background())

		syncFunc := func(ctx context.Context, ids int64) error {
			cancel()
			return nil
		}
		syncer := &changesetSyncer{
			syncStore:        newTestStore(),
			scheduleInterval: 10 * time.Minute,
			syncFunc:         syncFunc,
			priorityNotify:   make(chan []int64, 1),
			metrics:          makeMetrics(&observation.TestContext),
		}
		syncer.priorityNotify <- []int64{1}
		go syncer.Run(ctx)
		select {
		case <-ctx.Done():
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Sync not called")
		}
	})
}

func TestSyncRegistry_SyncCodeHosts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	externalServiceID := "https://example.com/"

	syncStore := newTestStore()
	syncStore.ListChangesetSyncDataFunc.SetDefaultReturn([]*btypes.ChangesetSyncData{
		{
			ChangesetID:           1,
			UpdatedAt:             time.Now(),
			RepoExternalServiceID: externalServiceID,
		},
	}, nil)

	codeHost := &btypes.CodeHost{ExternalServiceID: externalServiceID, ExternalServiceType: extsvc.TypeGitHub}
	codeHosts := []*btypes.CodeHost{codeHost}
	syncStore.ListCodeHostsFunc.SetDefaultHook(func(c context.Context, lcho store.ListCodeHostsOpts) ([]*btypes.CodeHost, error) {
		return codeHosts, nil
	})

	reg := NewSyncRegistry(ctx, syncStore, nil, &observation.TestContext)

	assertSyncerCount := func(t *testing.T, want int) {
		t.Helper()

		if len(reg.syncers) != want {
			t.Fatalf("Expected %d syncer, got %d", want, len(reg.syncers))
		}
	}

	reg.syncCodeHosts(ctx)
	assertSyncerCount(t, 1)

	// Adding it again should have no effect
	reg.addCodeHostSyncer(&btypes.CodeHost{ExternalServiceID: externalServiceID, ExternalServiceType: extsvc.TypeGitHub})
	assertSyncerCount(t, 1)

	// Simulate a service being removed
	codeHosts = []*btypes.CodeHost{}
	reg.syncCodeHosts(ctx)
	assertSyncerCount(t, 0)

	// And added again
	codeHosts = []*btypes.CodeHost{codeHost}
	reg.syncCodeHosts(ctx)
	assertSyncerCount(t, 1)
}

func TestSyncRegistry_EnqueueChangesetSyncs(t *testing.T) {
	t.Parallel()

	codeHostURL := "https://example.com/"

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	syncStore := newTestStore()
	syncStore.ListChangesetSyncDataFunc.SetDefaultReturn([]*btypes.ChangesetSyncData{
		{ChangesetID: 1, UpdatedAt: time.Now(), RepoExternalServiceID: codeHostURL},
		{ChangesetID: 3, UpdatedAt: time.Now(), RepoExternalServiceID: codeHostURL},
	}, nil)

	syncChan := make(chan int64)

	// In order to test that priority items are delivered we'll inject our own syncer
	// with a custom sync func
	syncerCtx, syncerCancel := context.WithCancel(ctx)
	t.Cleanup(syncerCancel)

	syncer := &changesetSyncer{
		syncStore:   syncStore,
		codeHostURL: codeHostURL,
		syncFunc: func(ctx context.Context, id int64) error {
			syncChan <- id
			return nil
		},
		priorityNotify: make(chan []int64, 1),
		metrics:        makeMetrics(&observation.TestContext),
		cancel:         syncerCancel,
	}
	go syncer.Run(syncerCtx)

	reg := NewSyncRegistry(ctx, syncStore, nil, &observation.TestContext)
	reg.syncers[codeHostURL] = syncer

	// Start handler in background, will be canceled when ctx is canceled
	go reg.handlePriorityItems()

	// Enqueue priority items, but only 1, 3 have valid ChangesetSyncData
	if err := reg.EnqueueChangesetSyncs(ctx, []int64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}

	// They should be delivered to the changesetSyncer
	for _, wantId := range []int64{1, 3} {
		select {
		case id := <-syncChan:
			if id != wantId {
				t.Fatalf("Expected %d, got %d", wantId, id)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for sync")
		}
	}
}

func TestLoadChangesetSource(t *testing.T) {
	ctx := context.Background()
	cf := httpcli.NewFactory(
		func(cli httpcli.Doer) httpcli.Doer {
			return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
				// Don't actually execute the request, just dump the authorization header
				// in the error, so we can assert on it further down.
				return nil, errors.New(req.Header.Get("Authorization"))
			})
		},
		httpcli.NewTimeoutOpt(1*time.Second),
	)

	externalService := types.ExternalService{
		ID:          1,
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub.com",
		Config:      `{"url": "https://github.com", "token": "123", "authorization": {}}`,
	}
	repo := &types.Repo{
		Name:    api.RepoName("test-repo"),
		URI:     "test-repo",
		Private: true,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "external-id-123",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: map[string]*types.SourceInfo{
			externalService.URN(): {
				ID:       externalService.URN(),
				CloneURL: "https://123@github.com/sourcegraph/sourcegraph",
			},
		},
	}

	// Store mocks.
	hasCredential := false
	syncStore := newTestStore()
	syncStore.GetSiteCredentialFunc.SetDefaultHook(func(ctx context.Context, opts store.GetSiteCredentialOpts) (*btypes.SiteCredential, error) {
		if hasCredential {
			cred := &btypes.SiteCredential{}
			cred.SetAuthenticator(ctx, &auth.OAuthBearerToken{Token: "456"})
			return cred, nil
		}
		return nil, store.ErrNoResults
	})
	ess := database.NewMockExternalServiceStore()
	ess.ListFunc.SetDefaultHook(func(ctx context.Context, options database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		return []*types.ExternalService{&externalService}, nil
	})
	syncStore.ExternalServicesFunc.SetDefaultReturn(ess)

	// If no site-credential exists, the token from the external service should be used.
	src, err := loadChangesetSource(ctx, cf, syncStore, repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := src.ValidateAuthenticator(ctx); err == nil {
		t.Fatal("unexpected nil error")
	} else if have, want := err.Error(), "Bearer 123"; have != want {
		t.Fatalf("invalid token used, want=%q have=%q", want, have)
	}

	// If one exists, prefer that one over the external service config ones.
	hasCredential = true
	src, err = loadChangesetSource(ctx, cf, syncStore, repo)
	if err != nil {
		t.Fatal(err)
	}
	if err := src.ValidateAuthenticator(ctx); err == nil {
		t.Fatal("unexpected nil error")
	} else if have, want := err.Error(), "Bearer 456"; have != want {
		t.Fatalf("invalid token used, want=%q have=%q", want, have)
	}
}
