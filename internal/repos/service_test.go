package repos_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func testServiceDeleteExternalService(t *testing.T, store *repos.Store) func(*testing.T) {
	ctx := context.Background()
	return transact(ctx, store, func(t testing.TB, tx *repos.Store) {
		t.Helper()

		svc := repos.NewService(tx)

		github1 := &types.ExternalService{Kind: extsvc.KindGitHub, DisplayName: "GitHub 1", Config: "{}"}
		github2 := &types.ExternalService{Kind: extsvc.KindGitHub, DisplayName: "GitHub 2", Config: "{}"}

		err := tx.ExternalServiceStore.Upsert(ctx, github1, github2)
		if err != nil {
			t.Fatal(err)
		}

		exclusive := &types.Repo{Name: "exclusive-repo-1", Sources: map[string]*types.SourceInfo{
			github1.URN(): {ID: github1.URN()},
		}}

		shared := &types.Repo{Name: "shared-repo-1", Sources: map[string]*types.SourceInfo{
			github1.URN(): {ID: github1.URN()},
			github2.URN(): {ID: github2.URN()},
		}}

		err = tx.RepoStore.Create(ctx, exclusive, shared)
		if err != nil {
			t.Fatal(err)
		}

		err = svc.DeleteExternalService(ctx, github1.ID)
		if err != nil {
			t.Fatal(err)
		}

		// Only the repos exclusively owned by this external service should have been deleted.
		_, err = tx.RepoStore.Get(ctx, exclusive.ID)
		if have, want := err, (&database.RepoNotFoundErr{ID: exclusive.ID}); !cmp.Equal(have, want) {
			t.Fatal("repo exclusively owned should have been deleted but wasn't")
		}

		have, err := tx.RepoStore.Get(ctx, shared.ID)
		if err != nil {
			t.Fatal(err)
		}

		delete(shared.Sources, github1.URN())

		if want := shared; !cmp.Equal(have, want) {
			t.Fatalf("shared repo diff: %v", cmp.Diff(have, want))
		}
	})
}
