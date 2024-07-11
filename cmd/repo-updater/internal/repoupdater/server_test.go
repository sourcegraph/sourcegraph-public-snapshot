package repoupdater

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestServer_EnqueueRepoUpdate(t *testing.T) {
	ctx := context.Background()

	svc := types.ExternalService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(`{
"url": "https://github.com",
"token": "secret-token",
"repos": ["owner/name"]
}`),
	}

	repo := types.Repo{
		ID:   1,
		Name: "github.com/foo/bar",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "bar",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com",
		},
		Metadata: new(github.Repository),
	}

	initStore := func(db database.DB) repos.Store {
		store := repos.NewStore(logtest.Scoped(t), db)
		if err := store.ExternalServiceStore().Upsert(ctx, &svc); err != nil {
			t.Fatal(err)
		}
		if err := store.RepoStore().Create(ctx, &repo); err != nil {
			t.Fatal(err)
		}
		return store
	}

	type testCase struct {
		name string
		repo api.RepoName
		res  *protocol.RepoUpdateResponse
		err  string
		init func(database.DB) repos.Store
	}

	testCases := []testCase{{
		name: "returns an error on store failure",
		init: func(realDB database.DB) repos.Store {
			mockRepos := dbmocks.NewMockRepoStore()
			mockRepos.ListFunc.SetDefaultReturn(nil, errors.New("boom"))
			realStore := initStore(realDB)
			mockStore := repos.NewMockStoreFrom(realStore)
			mockStore.RepoStoreFunc.SetDefaultReturn(mockRepos)
			return mockStore
		},
		err: `store.list-repos: boom`,
	}, {
		name: "missing repo",
		init: initStore,
		repo: "foo",
		err:  `repo foo not found with response: repo "foo" not found in store`,
	}, {
		name: "existing repo",
		repo: repo.Name,
		init: initStore,
		res: &protocol.RepoUpdateResponse{
			ID:   repo.ID,
			Name: string(repo.Name),
		},
	}}

	logger := logtest.Scoped(t)
	for _, tc := range testCases {
		tc := tc
		ctx := context.Background()

		t.Run(tc.name, func(t *testing.T) {
			sqlDB := dbtest.NewDB(t)
			store := tc.init(database.NewDB(logger, sqlDB))

			s := &Server{Logger: logger, Store: store, Scheduler: &fakeScheduler{}}
			gs := grpc.NewServer(defaults.ServerOptions(logger)...)
			proto.RegisterRepoUpdaterServiceServer(gs, s)

			srv := httptest.NewServer(internalgrpc.MultiplexHandlers(gs, http.NotFoundHandler()))
			defer srv.Close()

			cli := repoupdater.NewClient(srv.URL)
			if tc.err == "" {
				tc.err = "<nil>"
			}

			res, err := cli.EnqueueRepoUpdate(ctx, tc.repo)
			if have, want := fmt.Sprint(err), tc.err; !strings.Contains(have, want) {
				t.Errorf("have err: %q, want: %q", have, want)
			}

			if have, want := res, tc.res; !reflect.DeepEqual(have, want) {
				t.Errorf("response: %s", cmp.Diff(have, want))
			}
		})
	}
}

type fakeScheduler struct{}

func (s *fakeScheduler) UpdateOnce(_ api.RepoID, _ api.RepoName) {}
func (s *fakeScheduler) ScheduleInfo(_ api.RepoID) *protocol.RepoUpdateSchedulerInfoResult {
	return &protocol.RepoUpdateSchedulerInfoResult{}
}
