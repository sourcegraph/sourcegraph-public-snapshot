package internal_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	internaldb "github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// This error is passed to txstore.Done in order to always
// roll-back the transaction a test case executes in.
// This is meant to ensure each test case has a clean slate.
var errRollback = errors.New("tx: rollback")

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func transact(ctx context.Context, s *internal.Store, test func(testing.TB, *internal.Store)) func(*testing.T) {
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

func hasNoID(r *types.Repo) bool {
	return r.ID == 0
}

func hasID(ids ...api.RepoID) func(r *types.Repo) bool {
	return func(r *types.Repo) bool {
		for _, id := range ids {
			if r.ID == id {
				return true
			}
		}
		return false
	}
}

func isCloned(r *types.Repo) bool {
	return r.Cloned
}

func confGet() *conf.Unified {
	return &conf.Unified{}
}

func mustCreateServices(t testing.TB, store *internal.Store, svcs types.ExternalServices) {
	t.Helper()

	ctx := context.Background()

	for _, svc := range svcs {
		if err := store.ExternalServiceStore().Create(ctx, confGet, svc); err != nil {
			t.Fatalf("ExternalServiceStore.Create error: %v", err)
		}
	}
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db := dbtest.NewDB(t, *dsn)

	store := internal.NewStore(db, sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	lg := log15.New()
	lg.SetHandler(log15.DiscardHandler())
	store.Log = lg

	for _, tc := range []struct {
		name string
		test func(*testing.T, *internal.Store) func(*testing.T)
	}{
		{"Store/UpsertRepos", testStoreUpsertRepos},
		{"Store/UpsertSources", testStoreUpsertSources},
		{"Store/EnqueueSyncJobs", testStoreEnqueueSyncJobs(db, store)},
		{"Store/ListExternalRepoSpecs", testStoreListExternalRepoSpecs(db)},
		{"Store/SetClonedRepos", testStoreSetClonedRepos},
		{"Store/CountNotClonedRepos", testStoreCountNotClonedRepos},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				if t.Failed() {
					return
				}
				if _, err := db.Exec(`
DELETE FROM external_service_sync_jobs;
DELETE FROM external_service_repos;
DELETE FROM external_services;
DELETE FROM repo;
`); err != nil {
					t.Fatalf("cleaning up external services failed: %v", err)
				}
			})

			tc.test(t, store)(t)
		})
	}
}

func insertTestUser(t *testing.T, db *sql.DB) (userID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO users (username) VALUES ('bbs-admin') RETURNING id").Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func testStoreUpsertRepos(t *testing.T, store *internal.Store) func(*testing.T) {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()
	ctx := context.Background()

	return func(t *testing.T) {
		kinds := []string{
			extsvc.KindGitHub,
			extsvc.KindGitLab,
			extsvc.KindBitbucketServer,
			extsvc.KindAWSCodeCommit,
			extsvc.KindOther,
			extsvc.KindGitolite,
		}

		services := types.MakeExternalServices()

		if err := store.ExternalServiceStore().Upsert(ctx, services...); err != nil {
			t.Fatalf("Upsert error: %v", err)
		}

		servicesPerKind := types.ExternalServicesToMap(services)

		repositories := types.Repos{
			types.MakeGithubRepo(servicesPerKind[extsvc.KindGitHub]),
			types.MakeGitlabRepo(servicesPerKind[extsvc.KindGitLab]),
			types.MakeBitbucketServerRepo(servicesPerKind[extsvc.KindBitbucketServer]),
			types.MakeAWSCodeCommitRepo(servicesPerKind[extsvc.KindAWSCodeCommit]),
			types.MakeOtherRepo(servicesPerKind[extsvc.KindOther]),
			types.MakeGitoliteRepo(servicesPerKind[extsvc.KindGitolite]),
		}

		t.Run("no repos", func(t *testing.T) {
			if err := store.UpsertRepos(ctx); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
		})

		t.Run("many repos", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			want := types.GenerateRepos(7, repositories...)

			if err := tx.UpsertRepos(ctx, want...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
			if err := tx.UpsertSources(ctx, want.Sources(), nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}

			sort.Sort(want)

			if noID := want.Filter(hasNoID); len(noID) > 0 {
				t.Fatalf("UpsertRepos didn't assign an ID to all repos: %v", noID.Names())
			}

			have, err := tx.RepoStore().List(ctx, internaldb.ReposListOptions{
				ServiceTypes: kinds,
			})
			if err != nil {
				t.Fatalf("List error: %s", err)
			}

			if diff := cmp.Diff(have, []*types.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("List:\n%s", diff)
			}

			suffix := "-updated"
			now := clock.Now()
			for _, r := range want {
				r.Name += api.RepoName(suffix)
				r.URI += suffix
				r.Description += suffix
				r.UpdatedAt = now
				r.CreatedAt = now
				r.Archived = !r.Archived
				r.Fork = !r.Fork
			}

			if err = tx.UpsertRepos(ctx, want.Clone()...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if err = tx.UpsertSources(ctx, want.Clone().Sources(), nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			} else if have, err = tx.RepoStore().List(ctx, internaldb.ReposListOptions{}); err != nil {
				t.Errorf("List error: %s", err)
			} else if diff := cmp.Diff(have, []*types.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("List:\n%s", diff)
			}

			deleted := want.Clone().With(types.Opt.RepoDeletedAt(now))
			args := internaldb.ReposListOptions{}

			if err = tx.UpsertRepos(ctx, deleted...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			} else if have, err = tx.RepoStore().List(ctx, args); err != nil {
				t.Errorf("List error: %s", err)
			} else if diff := cmp.Diff(have, []*types.Repo(nil), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("List:\n%s", diff)
			}

			// Insert previously soft-deleted repos. Ensure we get back the same ID.
			if err = tx.UpsertRepos(ctx, want.Clone().With(types.Opt.RepoID(0))...); err != nil {
				t.Errorf("UpsertRepos error: %s", err)
			} else if err = tx.UpsertSources(ctx, want.Clone().Sources(), nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			} else if have, err = tx.RepoStore().List(ctx, internaldb.ReposListOptions{}); err != nil {
				t.Errorf("List error: %s", err)
			} else if diff := cmp.Diff(have, []*types.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("List:\n%s", diff)
			}

			// Delete all again, then try insert repos with different external
			// IDs but same name. Check we get new IDs.
			for _, r := range want {
				r.ID = 0
				r.ExternalRepo.ID += "-different"
			}
			if err = tx.UpsertRepos(ctx, deleted...); err != nil {
				t.Fatalf("UpsertRepos deleted error: %s", err)
			} else if err = tx.UpsertRepos(ctx, want...); err != nil {
				t.Fatalf("UpsertRepos want error: %s", err)
			} else if err = tx.UpsertSources(ctx, want.Sources(), nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			} else if have, err = tx.RepoStore().List(ctx, internaldb.ReposListOptions{}); err != nil {
				t.Errorf("List error: %s", err)
			} else if diff := cmp.Diff(have, []*types.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("List:\n%s", diff)
			} else if sameIDs := want.Filter(hasID(deleted.IDs()...)); len(sameIDs) > 0 {
				t.Errorf("List returned IDs of soft deleted repos: %v", sameIDs.Names())
			}
		}))

		t.Run("many repos soft-deleted and single repo reinserted", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			all := types.GenerateRepos(7, repositories...)

			if err := tx.UpsertRepos(ctx, all...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
			if err := tx.UpsertSources(ctx, all.Sources(), nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}

			sort.Sort(all)

			if noID := all.Filter(hasNoID); len(noID) > 0 {
				t.Fatalf("UpsertRepos didn't assign an ID to all repos: %v", noID.Names())
			}

			have, err := tx.RepoStore().List(ctx, internaldb.ReposListOptions{
				ServiceTypes: kinds,
			})
			if err != nil {
				t.Fatalf("List error: %s", err)
			}

			if diff := cmp.Diff(have, []*types.Repo(all), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("List:\n%s", diff)
			}

			allDeleted := all.Clone().With(types.Opt.RepoDeletedAt(now))
			args := internaldb.ReposListOptions{}

			if err = tx.UpsertRepos(ctx, allDeleted...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			} else if have, err = tx.RepoStore().List(ctx, args); err != nil {
				t.Errorf("List error: %s", err)
			} else if diff := cmp.Diff(have, []*types.Repo(nil), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("List:\n%s", diff)
			}

			// Insert one of the previously soft-deleted repos. Ensure ID on upserted repo is set and we get back the same ID.
			want := types.Repos{all[0]}
			upsert := want.Clone().With(types.Opt.RepoID(0))
			if err = tx.UpsertRepos(ctx, upsert...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}
			if upsert[0].ID == 0 {
				t.Fatalf("Repo ID is zero")
			}
			if err := tx.UpsertSources(ctx, upsert.Sources(), nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}

			if have, err = tx.RepoStore().List(ctx, internaldb.ReposListOptions{}); err != nil {
				t.Fatalf("List error: %s", err)
			}
			if diff := cmp.Diff(have, []*types.Repo(want), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("List:\n%s", diff)
			}
		}))

		t.Run("it shouldn't modify the cloned column", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			// UpsertRepos shouldn't set the cloned column to true
			r := types.GenerateRepos(1, repositories...)[0]
			r.Cloned = true
			if err := tx.UpsertRepos(ctx, r); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			count, err := tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if count != 1 {
				t.Fatalf("Wrong number of not cloned repos: %d", count)
			}

			// UpsertRepos shouldn't set the cloned column to false either
			if err := tx.SetClonedRepos(ctx, string(r.Name)); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
			r = r.Clone()
			r.Cloned = false
			if err := tx.UpsertRepos(ctx, r); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			count, err = tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if count != 0 {
				t.Fatalf("Wrong number of not cloned repos: %d", count)
			}
		}))
	}
}

func testStoreUpsertSources(t *testing.T, store *internal.Store) func(*testing.T) {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	services := types.MakeExternalServices()
	servicesPerKind := types.ExternalServicesToMap(services)
	mustCreateServices(t, store, services)

	return func(t *testing.T) {
		github := types.MakeGithubRepo(servicesPerKind[extsvc.KindGitHub])
		gitlab := types.MakeGitlabRepo(servicesPerKind[extsvc.KindGitLab])

		repositories := types.Repos{
			github,
			gitlab,
		}

		ctx := context.Background()

		t.Run("no sources", func(t *testing.T) {
			if err := store.UpsertSources(ctx, nil, nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}
		})

		t.Run("delete repo", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			want := types.GenerateRepos(7, repositories...)

			if err := tx.UpsertRepos(ctx, want...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sources := want.Sources()

			if err := tx.UpsertSources(ctx, sources, nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}

			// delete a repository
			want[0].DeletedAt = now
			if err := tx.RepoStore().Delete(ctx, want[0].ID); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			// un delete it
			want[0].DeletedAt = time.Time{}
			if err := tx.UpsertRepos(ctx, want...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			// it should not contain any source
			want[0].Sources = nil

			got, err := tx.RepoStore().List(ctx, internaldb.ReposListOptions{})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			if diff := cmp.Diff([]*types.Repo(want), got, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListRepos:\n%s", diff)
			}
		}))

		t.Run("delete external service", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			origRepos := types.GenerateRepos(7, repositories...)

			if err := tx.UpsertRepos(ctx, origRepos...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sources := origRepos.Sources()

			if err := tx.UpsertSources(ctx, sources, nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}

			// delete an external service
			svc := servicesPerKind[extsvc.KindGitHub]
			svc.DeletedAt = now
			if err := tx.ExternalServiceStore().Upsert(ctx, svc); err != nil {
				t.Fatalf("UpsertExternalServices error: %s", err)
			}

			// un delete it
			svc.DeletedAt = time.Time{}
			if err := tx.ExternalServiceStore().Upsert(ctx, svc); err != nil {
				t.Fatalf("UpsertExternalServices error: %s", err)
			}

			// All GitHub sources should be deleted and all orphan repositories should be excluded
			want := make([]*types.Repo, 0, len(origRepos))
			origRepos.Apply(func(r *types.Repo) {
				for urn := range r.Sources {
					if strings.Contains(urn, "github") {
						delete(r.Sources, urn)
					}
				}
				if len(r.Sources) > 0 {
					want = append(want, r)
				}
			})

			got, err := tx.RepoStore().List(ctx, internaldb.ReposListOptions{})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListRepos:\n%s", diff)
			}
		}))

		t.Run("inserts updates and deletes", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			want := types.GenerateRepos(7, repositories...)

			if err := tx.UpsertRepos(ctx, want...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			sources := want.Sources()

			if err := tx.UpsertSources(ctx, sources, nil, nil); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}

			have, err := tx.RepoStore().List(ctx, internaldb.ReposListOptions{})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			if diff := cmp.Diff([]*types.Repo(want), have, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListRepos:\n%s", diff)
			}

			updates := make(map[api.RepoID][]types.SourceInfo)
			deletes := make(map[api.RepoID][]types.SourceInfo)

			updates[want[0].ID] = sources[want[0].ID]
			updates[want[0].ID][0].CloneURL = "something-else"
			deletes[want[1].ID] = sources[want[1].ID]

			if err := tx.UpsertSources(ctx, nil, updates, deletes); err != nil {
				t.Fatalf("UpsertSources error: %s", err)
			}

			want[0].Sources[servicesPerKind[extsvc.KindGitHub].URN()] = &types.SourceInfo{
				CloneURL: "something-else",
				ID:       servicesPerKind[extsvc.KindGitHub].URN(),
			}

			// Remove the second element from want because it should be deleted automatically
			// by the time it become orphaned.
			want = append(want[:1], want[2:]...)

			have, err = tx.RepoStore().List(ctx, internaldb.ReposListOptions{})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			if diff := cmp.Diff([]*types.Repo(want), have, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("ListRepos:\n%s", diff)
			}
		}))
	}
}

func testStoreSetClonedRepos(t *testing.T, store *internal.Store) func(*testing.T) {
	services := types.MakeExternalServices()
	servicesPerKind := types.ExternalServicesToMap(services)
	mustCreateServices(t, store, services)

	return func(t *testing.T) {
		var repositories types.Repos
		for i := 0; i < 3; i++ {
			repositories = append(repositories, types.MakeGithubRepo(servicesPerKind[extsvc.KindGitHub]))
		}

		check := func(t testing.TB, ctx context.Context, tx *internal.Store, wantNames []string) {
			t.Helper()

			res, err := tx.RepoStore().List(ctx, internaldb.ReposListOptions{})
			if err != nil {
				t.Fatalf("ListRepos error: %s", err)
			}

			cloned := types.Repos(res).Filter(isCloned).Names()
			sort.Strings(cloned)

			if got, want := cloned, wantNames; !cmp.Equal(got, want) {
				t.Fatalf("got=%v, want=%v: %s", got, want, cmp.Diff(got, want))
			}
		}

		ctx := context.Background()

		t.Run("no repo name", func(t *testing.T) {
			if err := store.SetClonedRepos(ctx); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
		})

		t.Run("many repo names", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			stored := types.GenerateRepos(9, repositories...)

			if err := tx.RepoStore().Create(ctx, stored...); err != nil {
				t.Fatalf("RepoStore.Create error: %s", err)
			}

			sort.Sort(stored)

			names := stored[:3].Names()
			sort.Strings(names)

			if err := tx.SetClonedRepos(ctx, names...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
			check(t, ctx, tx, names)

			// setClonedRepositories should be idempotent and have the same behavior
			// when called with the same repos
			if err := tx.SetClonedRepos(ctx, names...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}
			check(t, ctx, tx, names)

			// when adding another repo to the list, the other repos must be set as well
			names = stored[:4].Names()
			sort.Strings(names)
			if err := tx.SetClonedRepos(ctx, names...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			check(t, ctx, tx, names)
		}))

		t.Run("repo names in mixed case", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			stored := types.GenerateRepos(9, repositories...)
			for i := range stored {
				if i%2 == 0 {
					stored[i].Name = api.RepoName(strings.ToUpper(string(stored[i].Name)))
				}
			}

			if err := tx.RepoStore().Create(ctx, stored...); err != nil {
				t.Fatalf("RepoStore.Create error: %s", err)
			}

			sort.Sort(stored)

			originalNames := stored.Names()
			sort.Strings(originalNames)

			lowerCaseNames := make([]string, 0, len(originalNames))
			for _, n := range originalNames {
				lowerCaseNames = append(lowerCaseNames, strings.ToLower(n))
			}

			if err := tx.SetClonedRepos(ctx, lowerCaseNames...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			check(t, ctx, tx, originalNames)
		}))
	}
}

func testStoreCountNotClonedRepos(t *testing.T, store *internal.Store) func(*testing.T) {
	return func(t *testing.T) {
		services := types.MakeExternalServices()
		servicesPerKind := types.ExternalServicesToMap(services)

		mustCreateServices(t, store, services)

		var repositories types.Repos
		for i := 0; i < 3; i++ {
			repositories = append(repositories, types.MakeGithubRepo(servicesPerKind[extsvc.KindGitHub]))
		}

		ctx := context.Background()

		t.Run("no cloned repos", func(t *testing.T) {
			count, err := store.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if diff := cmp.Diff(count, uint64(0)); diff != "" {
				t.Fatalf("CountNotClonedRepos:\n%s", diff)
			}
		})

		t.Run("multiple cloned repos", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			stored := types.GenerateRepos(10, repositories...)

			if err := tx.RepoStore().Create(ctx, stored...); err != nil {
				t.Fatalf("RepoStore.Create error: %s", err)
			}

			sort.Sort(stored)
			cloned := stored[:3].Names()

			if err := tx.SetClonedRepos(ctx, cloned...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			sort.Strings(cloned)

			count, err := tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if diff := cmp.Diff(count, uint64(7)); diff != "" {
				t.Fatalf("CountNotClonedRepos:\n%s", diff)
			}
		}))

		t.Run("deleted non cloned repos", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			stored := types.GenerateRepos(10, repositories...)

			if err := tx.RepoStore().Create(ctx, stored...); err != nil {
				t.Fatalf("RepoStore.Create error: %s", err)
			}

			sort.Sort(stored)
			cloned := stored[:3].Names()

			if err := tx.SetClonedRepos(ctx, cloned...); err != nil {
				t.Fatalf("SetClonedRepos error: %s", err)
			}

			sort.Strings(cloned)
			deletedCloned := stored[8:]

			if err := tx.RepoStore().Delete(ctx, deletedCloned.IDs()...); err != nil {
				t.Fatalf("UpsertRepos error: %s", err)
			}

			count, err := tx.CountNotClonedRepos(ctx)
			if err != nil {
				t.Fatalf("CountNotClonedRepos error: %s", err)
			}
			if diff := cmp.Diff(count, uint64(5)); diff != "" {
				t.Fatalf("CountNotClonedRepos:\n%s", diff)
			}
		}))
	}
}

func testStoreListRepos(t *testing.T, store *internal.Store) func(*testing.T) {
	clock := dbtesting.NewFakeClock(time.Now(), 0)
	now := clock.Now()

	services := types.MakeExternalServices()
	servicesPerKind := types.ExternalServicesToMap(services)
	mustCreateServices(t, store, services)

	unmanaged := types.MakeRepo("unmanaged", "https://example.com/", "non_existent_kind")
	github := types.MakeGithubRepo(servicesPerKind[extsvc.KindGitHub])
	gitlab := types.MakeGitlabRepo(servicesPerKind[extsvc.KindGitLab])
	bitbucketServer := types.MakeBitbucketServerRepo(servicesPerKind[extsvc.KindBitbucketServer])
	awsCodeCommit := types.MakeAWSCodeCommitRepo(servicesPerKind[extsvc.KindAWSCodeCommit])
	otherRepo := types.MakeOtherRepo(servicesPerKind[extsvc.KindOther])
	gitoliteRepo := types.MakeGitoliteRepo(servicesPerKind[extsvc.KindGitolite])

	repositories := types.Repos{
		github,
		gitlab,
		bitbucketServer,
		awsCodeCommit,
		otherRepo,
		gitoliteRepo,
	}

	kinds := []string{
		extsvc.KindGitHub,
		extsvc.KindGitLab,
		extsvc.KindBitbucketServer,
		extsvc.KindAWSCodeCommit,
		extsvc.KindOther,
		extsvc.KindGitolite,
	}

	type testCase struct {
		name   string
		args   func(stored types.Repos) internaldb.ReposListOptions
		stored types.Repos
		repos  types.ReposAssertion
		err    error
	}

	var testCases []testCase
	{
		stored := repositories.With(func(r *types.Repo) {
			r.ExternalRepo.ServiceType =
				strings.ToUpper(r.ExternalRepo.ServiceType)
		})

		testCases = append(testCases, testCase{
			name: "case-insensitive kinds",
			args: func(_ types.Repos) (args internaldb.ReposListOptions) {
				for _, kind := range kinds {
					args.ServiceTypes = append(args.ServiceTypes, strings.ToUpper(kind))
				}
				return args
			},
			stored: stored,
			repos:  types.Assert.ReposEqual(stored...),
		})
	}

	testCases = append(testCases, testCase{
		name: "ignores unmanaged",
		args: func(_ types.Repos) internaldb.ReposListOptions {
			return internaldb.ReposListOptions{ServiceTypes: kinds}
		},
		stored: types.Repos{github, gitlab, unmanaged}.Clone(),
		repos:  types.Assert.ReposEqual(github, gitlab),
	})

	{
		stored := repositories.With(types.Opt.RepoDeletedAt(now))
		testCases = append(testCases, testCase{
			name:   "excludes soft deleted repos by default",
			stored: stored,
			repos:  types.Assert.ReposEqual(),
		})
	}

	testCases = append(testCases, testCase{
		name:   "returns repos in ascending order by id",
		stored: types.GenerateRepos(7, repositories...),
		repos: types.Assert.ReposOrderedBy(func(a, b *types.Repo) bool {
			return a.ID < b.ID
		}),
	})

	testCases = append(testCases, testCase{
		name:   "returns repos by their names",
		stored: repositories,
		args: func(_ types.Repos) internaldb.ReposListOptions {
			return internaldb.ReposListOptions{
				Names: []string{string(github.Name), string(gitlab.Name)},
			}
		},
		repos: types.Assert.ReposEqual(github, gitlab),
	})

	testCases = append(testCases, testCase{
		name:   "returns repos by their ids",
		stored: repositories,
		args: func(stored types.Repos) internaldb.ReposListOptions {
			return internaldb.ReposListOptions{
				IDs: []api.RepoID{stored[0].ID, stored[1].ID},
			}
		},
		repos: types.Assert.ReposEqual(repositories[:2].Clone()...),
	})

	testCases = append(testCases, testCase{
		name:   "limits repos to the given kinds",
		stored: repositories,
		args: func(types.Repos) internaldb.ReposListOptions {
			return internaldb.ReposListOptions{
				ServiceTypes: []string{extsvc.KindGitHub, extsvc.KindGitLab},
			}
		},
		repos: types.Assert.ReposEqual(github, gitlab),
	})

	testCases = append(testCases, testCase{
		name: "only include private",
		args: func(types.Repos) internaldb.ReposListOptions {
			return internaldb.ReposListOptions{
				OnlyPrivate: true,
			}
		},
		stored: repositories,
		repos:  types.Assert.ReposEqual(gitlab),
	})

	testCases = append(testCases, testCase{
		name:   "use or",
		stored: repositories,
		args: func(types.Repos) internaldb.ReposListOptions {
			return internaldb.ReposListOptions{
				Names:        []string{"gitlab.com/bar/foo"},
				ServiceTypes: []string{"github"},
				UseOr:        true,
			}
		},
		repos: types.Assert.ReposEqual(github, gitlab),
	})

	testCases = append(testCases, testCase{
		name:   "use and",
		stored: repositories,
		args: func(types.Repos) internaldb.ReposListOptions {
			return internaldb.ReposListOptions{
				Names:        []string{"gitlab.com/bar/foo"},
				ServiceTypes: []string{"github"},
				UseOr:        false,
			}
		},
		repos: types.Assert.ReposEqual(),
	})

	{
		testCases = append(testCases, testCase{
			name:   "limit by external service",
			stored: repositories,
			args: func(types.Repos) internaldb.ReposListOptions {
				return internaldb.ReposListOptions{
					ExternalServiceID: servicesPerKind[extsvc.KindGitHub].ID,
				}
			},
			repos: types.Assert.ReposEqual(github),
		})
	}

	return func(t *testing.T) {
		ctx := context.Background()

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, transact(ctx, store, func(t testing.TB, tx *internal.Store) {
				stored := tc.stored.Clone()

				if err := tx.RepoStore().Create(ctx, stored...); err != nil {
					t.Fatalf("failed to setup store: %v", err)
				}

				var args internaldb.ReposListOptions
				if tc.args != nil {
					args = tc.args(stored)
				}

				rs, err := tx.RepoStore().List(ctx, args)
				if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
					t.Errorf("error:\nhave: %v\nwant: %v", have, want)
				}

				if tc.repos != nil {
					tc.repos(t, rs)
				}
			}))
		}

		t.Run("only include cloned", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			stored := types.GenerateRepos(5, repositories...).Clone()
			if err := tx.RepoStore().Create(ctx, stored...); err != nil {
				t.Fatalf("failed to setup store: %v", err)
			}

			sort.Sort(stored)

			cloned := stored[:3]
			if err := tx.SetClonedRepos(ctx, cloned.Names()...); err != nil {
				t.Fatalf("failed to set cloned repos: %v", err)
			}

			args := internaldb.ReposListOptions{
				OnlyCloned: true,
			}

			rs, err := tx.RepoStore().List(ctx, args)
			if err != nil {
				t.Errorf("failed to list repos: %v", err)
			}

			want := cloned.With(func(r *types.Repo) {
				r.Cloned = true
			})

			types.Assert.ReposEqual(want...)(t, rs)
		}))
	}
}

func testStoreListReposPagination(t *testing.T, store *internal.Store) func(*testing.T) {
	services := types.MakeExternalServices()
	servicesPerKind := types.ExternalServicesToMap(services)
	mustCreateServices(t, store, services)

	github := types.MakeGithubRepo(servicesPerKind[extsvc.KindGitHub])

	return func(t *testing.T) {
		ctx := context.Background()
		t.Run("", transact(ctx, store, func(t testing.TB, tx *internal.Store) {
			stored := types.GenerateRepos(7, github)
			if err := tx.RepoStore().Create(ctx, stored...); err != nil {
				t.Fatalf("RepoStore.Create error: %s", err)
			}

			sort.Sort(stored)

			lo, hi := -2, len(stored)+2
			for page := lo; page < hi; page++ {
				for limit := lo; limit < hi; limit++ {
					args := internaldb.ReposListOptions{
						LimitOffset: &internaldb.LimitOffset{
							Offset: (page + 1) * limit,
							Limit:  limit,
						},
					}

					listed, err := tx.RepoStore().List(ctx, args)
					if err != nil {
						t.Fatalf("unexpected error with page=%d, limit=%d: %v", page, limit, err)
					}

					var want types.Repos
					if limit <= 0 || limit >= len(stored) {
						want = stored
					} else {
						want = stored[:limit]
					}

					if have := types.Repos(listed); !reflect.DeepEqual(have, want) {
						t.Fatalf("page=%d, limit=%d: %s", page, limit, cmp.Diff(have, want))
					}
				}
			}
		}))
	}
}

func testStoreListExternalRepoSpecs(db *sql.DB) func(t *testing.T, repoStore *internal.Store) func(*testing.T) {
	return func(t *testing.T, store *internal.Store) func(*testing.T) {
		return func(t *testing.T) {
			ctx := context.Background()

			// Insert test repositories
			_, err := db.ExecContext(ctx, `
INSERT INTO repo (id, name, description, fork, external_id, external_service_type, external_service_id, deleted_at)
VALUES
	(1, 'github.com/user/repo1', '', FALSE, NULL, 'github', 'https://github.com/', NULL),
	(2, 'github.com/user/repo2', '', FALSE, 'MDEwOlJlcG9zaXRvcnky', NULL, 'https://github.com/', NULL),
	(3, 'github.com/user/repo3', '', FALSE, 'MDEwOlJlcG9zaXRvcnkz', 'github', NULL, NULL),
	(4, 'github.com/user/repo4', '', FALSE, 'MDEwOlJlcG9zaXRvcnk0', 'github', 'https://github.com/', NOW()),
	(5, 'github.com/user/repo5', '', FALSE, 'MDEwOlJlcG9zaXRvcnk1', 'github', 'https://github.com/', NULL)
`)
			if err != nil {
				t.Fatal(err)
			}

			ids, err := store.ListExternalRepoSpecs(ctx)
			if err != nil {
				t.Fatal(err)
			}
			want := map[api.ExternalRepoSpec]struct{}{
				{
					ID:          "MDEwOlJlcG9zaXRvcnk1",
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				}: {},
			}
			if diff := cmp.Diff(want, ids); diff != "" {
				t.Fatalf("IDs mismatch (-want +got):\n%s", diff)
			}
		}
	}
}

func testStoreEnqueueSyncJobs(db *sql.DB, store *internal.Store) func(t *testing.T, store *internal.Store) func(*testing.T) {
	return func(t *testing.T, _ *internal.Store) func(*testing.T) {
		t.Helper()

		clock := dbtesting.NewFakeClock(time.Now(), 0)
		now := clock.Now()

		services := types.GenerateExternalServices(10, types.MakeExternalServices()...)

		type testCase struct {
			name            string
			stored          types.ExternalServices
			queued          func(types.ExternalServices) []int64
			ignoreSiteAdmin bool
			err             error
		}

		var testCases []testCase

		testCases = append(testCases, testCase{
			name: "enqueue everything",
			stored: services.With(func(s *types.ExternalService) {
				s.NextSyncAt = now.Add(-10 * time.Second)
			}),
			queued: func(svcs types.ExternalServices) []int64 { return svcs.IDs() },
		})

		testCases = append(testCases, testCase{
			name: "nothing to enqueue",
			stored: services.With(func(s *types.ExternalService) {
				s.NextSyncAt = now.Add(10 * time.Second)
			}),
			queued: func(svcs types.ExternalServices) []int64 { return []int64{} },
		})

		testCases = append(testCases, testCase{
			name: "ignore siteadmin repos",
			stored: services.With(func(s *types.ExternalService) {
				s.NextSyncAt = now.Add(10 * time.Second)
			}),
			ignoreSiteAdmin: true,
			queued:          func(svcs types.ExternalServices) []int64 { return []int64{} },
		})

		{
			i := 0
			testCases = append(testCases, testCase{
				name: "some to enqueue",
				stored: services.With(func(s *types.ExternalService) {
					if i%2 == 0 {
						s.NextSyncAt = now.Add(10 * time.Second)
					} else {
						s.NextSyncAt = now.Add(-10 * time.Second)
					}
					i++
				}),
				queued: func(svcs types.ExternalServices) []int64 {
					var ids []int64
					for i := range svcs {
						if i%2 != 0 {
							ids = append(ids, svcs[i].ID)
						}
					}
					return ids
				},
			})
		}

		return func(t *testing.T) {
			ctx := context.Background()

			for _, tc := range testCases {
				tc := tc

				t.Run(tc.name, func(t *testing.T) {
					t.Cleanup(func() {
						if _, err := db.ExecContext(ctx, "DELETE FROM external_service_sync_jobs;DELETE FROM external_services"); err != nil {
							t.Fatal(err)
						}
					})
					stored := tc.stored.Clone()

					if err := store.ExternalServiceStore().Upsert(ctx, stored...); err != nil {
						t.Fatalf("failed to setup store: %v", err)
					}

					err := store.EnqueueSyncJobs(ctx, tc.ignoreSiteAdmin)
					if have, want := fmt.Sprint(err), fmt.Sprint(tc.err); have != want {
						t.Errorf("error:\nhave: %v\nwant: %v", have, want)
					}

					jobs, err := store.ListSyncJobs(ctx)
					if err != nil {
						t.Fatal(err)
					}

					gotIDs := make([]int64, 0, len(jobs))
					for _, job := range jobs {
						gotIDs = append(gotIDs, job.ExternalServiceID)
					}

					want := tc.queued(stored)
					sort.Slice(gotIDs, func(i, j int) bool {
						return gotIDs[i] < gotIDs[j]
					})
					sort.Slice(want, func(i, j int) bool {
						return want[i] < want[j]
					})

					if diff := cmp.Diff(want, gotIDs); diff != "" {
						t.Fatal(diff)
					}
				})
			}
		}
	}
}
