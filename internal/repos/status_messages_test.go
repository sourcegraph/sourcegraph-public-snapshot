package repos

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestStatusMessages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	db := dbtest.NewDB(t, "")
	store := NewStore(db, sql.TxOptions{})

	githubService := &types.ExternalService{
		ID:          1,
		Config:      `{}`,
		Kind:        extsvc.KindGitHub,
		DisplayName: "github.com - test",
	}

	err := database.ExternalServices(db).Upsert(ctx, githubService)
	if err != nil {
		t.Fatal(err)
	}

	user1, err := database.Users(db).Create(ctx, database.NewUser{
		Email:                 "a1@example.com",
		Username:              "u1",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name            string
		stored          types.Repos
		gitserverCloned []string
		sourcerErr      error
		listRepoErr     error
		res             []StatusMessage
		err             string
	}{
		{
			name:            "all cloned",
			gitserverCloned: []string{"foobar"},
			stored:          []*types.Repo{{Name: "foobar", Cloned: true}},
			res:             nil,
		},
		{
			name:            "nothing cloned",
			stored:          []*types.Repo{{Name: "foobar"}},
			gitserverCloned: []string{},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "1 repositories enqueued for cloning...",
					},
				},
			},
		},
		{
			name:            "subset cloned",
			stored:          []*types.Repo{{Name: "foobar", Cloned: true}, {Name: "barfoo"}},
			gitserverCloned: []string{"foobar"},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "1 repositories enqueued for cloning...",
					},
				},
			},
		},
		{
			name:            "more cloned than stored",
			stored:          []*types.Repo{{Name: "foobar", Cloned: true}},
			gitserverCloned: []string{"foobar", "barfoo"},
			res:             nil,
		},
		{
			name:            "cloned different than stored",
			stored:          []*types.Repo{{Name: "foobar"}, {Name: "barfoo"}},
			gitserverCloned: []string{"one", "two", "three"},
			res: []StatusMessage{
				{
					Cloning: &CloningProgress{
						Message: "2 repositories enqueued for cloning...",
					},
				},
			},
		},
		{
			name:            "case insensitivity",
			gitserverCloned: []string{"foobar"},
			stored:          []*types.Repo{{Name: "FOOBar", Cloned: true}},
			res:             nil,
		},
		{
			name:            "case insensitivity to gitserver names",
			gitserverCloned: []string{"FOOBar"},
			stored:          []*types.Repo{{Name: "FOOBar", Cloned: true}},
			res:             nil,
		},
		{
			name:       "one external service syncer err",
			sourcerErr: errors.New("github is down"),
			res: []StatusMessage{
				{
					ExternalServiceSyncError: &ExternalServiceSyncError{
						Message:           "fetching from code host: 1 error occurred:\n\t* github is down\n\n",
						ExternalServiceId: githubService.ID,
					},
				},
			},
		},
		{
			name:        "one syncer err",
			listRepoErr: errors.New("could not connect to database"),
			res: []StatusMessage{
				{
					ExternalServiceSyncError: &ExternalServiceSyncError{
						Message:           "syncer.sync.store.list-repos: could not connect to database",
						ExternalServiceId: githubService.ID,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			stored := tc.stored.Clone()
			var cloned []string
			for _, r := range stored {
				r.ExternalRepo = api.ExternalRepoSpec{
					ID:          uuid.New().String(),
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				}
				if r.Cloned {
					cloned = append(cloned, string(r.Name))
				}
			}

			err := database.Repos(db).Create(ctx, stored...)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				ids := make([]api.RepoID, 0, len(stored))
				for _, r := range stored {
					ids = append(ids, r.ID)
				}
				err := database.Repos(db).Delete(ctx, ids...)
				if err != nil {
					t.Fatal(err)
				}
			})

			err = store.SetClonedRepos(ctx, cloned...)
			if err != nil {
				t.Fatal(err)
			}

			clock := timeutil.NewFakeClock(time.Now(), 0)
			syncer := &Syncer{
				Store: store,
				Now:   clock.Now,
			}

			if tc.sourcerErr != nil || tc.listRepoErr != nil {
				database.Mocks.Repos.List = func(v0 context.Context, v1 database.ReposListOptions) ([]*types.Repo, error) {
					return nil, tc.listRepoErr
				}
				defer func() {
					database.Mocks.Repos.List = nil
				}()
				sourcer := NewFakeSourcer(tc.sourcerErr, NewFakeSource(githubService, nil))
				// Run Sync so that possibly `LastSyncErrors` is set
				syncer.Sourcer = sourcer

				err = syncer.SyncExternalService(ctx, store, githubService.ID, time.Millisecond)

				// In prod, SyncExternalService is kicked off by a worker queue. Any error
				// returned will be stored in the external_service_sync_jobs table so we fake
				// that here.
				if err != nil {
					defer func() { database.Mocks.ExternalServices = database.MockExternalServices{} }()
					database.Mocks.ExternalServices.ListSyncErrors = func(ctx context.Context) (map[int64]string, error) {
						return map[int64]string{
							githubService.ID: err.Error(),
						}, nil
					}
				}

			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := FetchStatusMessages(ctx, db, user1)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("have err: %q, want: %q", have, want)
			}

			if have, want := res, tc.res; !reflect.DeepEqual(have, want) {
				t.Errorf("response: %s", cmp.Diff(have, want))
			}
		})
	}
}
