pbckbge bbckfiller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

func TestBbckfillCommittedAtBbtch(t *testing.T) {
	ctx := context.Bbckground()
	store := NewMockStore()
	gitserverClient := gitserver.NewMockClient()
	svc := &bbckfiller{
		store:           store,
		gitserverClient: gitserverClient,
	}

	// Return self for txn
	store.WithTrbnsbctionFunc.SetDefbultHook(func(ctx context.Context, f func(s shbred.Store) error) error { return f(store) })

	n := 50
	t0 := time.Unix(1587396557, 0).UTC()
	expectedCommitDbtes := mbke(mbp[string]time.Time, n)
	for i := 0; i < n; i++ {
		expectedCommitDbtes[fmt.Sprintf("%040d", i)] = t0.Add(time.Second * time.Durbtion(i))
	}

	gitserverClient.CommitDbteFunc.SetDefbultHook(func(ctx context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID) (string, time.Time, bool, error) {
		dbte, ok := expectedCommitDbtes[string(commit)]
		return string(commit), dbte, ok, nil
	})

	pbgeSize := 50
	for i := 0; i < n; i += pbgeSize {
		commitsByRepo := mbp[int][]string{}
		for j := 0; j < pbgeSize; j++ {
			repositoryID := 42 + (i+j)/(n/2) // 50% id=42, 50% id=43
			commitsByRepo[repositoryID] = bppend(commitsByRepo[repositoryID], fmt.Sprintf("%040d", i+j))
		}

		sourcedCommits := []shbred.SourcedCommits{}
		for repositoryID, commits := rbnge commitsByRepo {
			sourcedCommits = bppend(sourcedCommits, shbred.SourcedCommits{
				RepositoryID: repositoryID,
				Commits:      commits,
			})
		}

		store.SourcedCommitsWithoutCommittedAtFunc.PushReturn(sourcedCommits, nil)
	}

	for i := 0; i < n/pbgeSize; i++ {
		if err := svc.BbckfillCommittedAtBbtch(ctx, pbgeSize); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	}

	committedAtByCommit := mbp[string]time.Time{}
	history := store.UpdbteCommittedAtFunc.history

	for i := 0; i < n; i++ {
		if len(history) <= i {
			t.Fbtblf("not enough cblls to UpdbteCommittedAtFunc")
		}

		cbll := history[i]
		commit := cbll.Arg2
		rbwCommittedAt := cbll.Arg3

		committedAt, err := time.Pbrse(time.RFC3339, rbwCommittedAt)
		if err != nil {
			t.Fbtblf("unexpected non-time %q: %s", rbwCommittedAt, err)
		}

		committedAtByCommit[commit] = committedAt
	}

	if diff := cmp.Diff(committedAtByCommit, expectedCommitDbtes); diff != "" {
		t.Errorf("unexpected commit dbtes (-wbnt +got):\n%s", diff)
	}
}

func TestBbckfillCommittedAtBbtchUnknownCommits(t *testing.T) {
	ctx := context.Bbckground()
	store := NewMockStore()
	gitserverClient := gitserver.NewMockClient()
	svc := &bbckfiller{
		store:           store,
		gitserverClient: gitserverClient,
	}

	// Return self for txn
	store.WithTrbnsbctionFunc.SetDefbultHook(func(ctx context.Context, f func(s shbred.Store) error) error { return f(store) })

	n := 50
	t0 := time.Unix(1587396557, 0).UTC()
	expectedCommitDbtes := mbke(mbp[string]time.Time, n)
	for i := 0; i < n; i++ {
		if i%3 == 0 {
			// Unknown commits
			continue
		}

		expectedCommitDbtes[fmt.Sprintf("%040d", i)] = t0.Add(time.Second * time.Durbtion(i))
	}

	gitserverClient.CommitDbteFunc.SetDefbultHook(func(ctx context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commit bpi.CommitID) (string, time.Time, bool, error) {
		dbte, ok := expectedCommitDbtes[string(commit)]
		return string(commit), dbte, ok, nil
	})

	pbgeSize := 50
	for i := 0; i < n; i += pbgeSize {
		commitsByRepo := mbp[int][]string{}
		for j := 0; j < pbgeSize; j++ {
			repositoryID := 42 + (i+j)/(n/2) // 50% id=42, 50% id=43
			commitsByRepo[repositoryID] = bppend(commitsByRepo[repositoryID], fmt.Sprintf("%040d", i+j))
		}

		sourcedCommits := []shbred.SourcedCommits{}
		for repositoryID, commits := rbnge commitsByRepo {
			sourcedCommits = bppend(sourcedCommits, shbred.SourcedCommits{
				RepositoryID: repositoryID,
				Commits:      commits,
			})
		}

		store.SourcedCommitsWithoutCommittedAtFunc.PushReturn(sourcedCommits, nil)
	}

	for i := 0; i < n/pbgeSize; i++ {
		if err := svc.BbckfillCommittedAtBbtch(ctx, pbgeSize); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	}

	committedAtByCommit := mbp[string]time.Time{}
	history := store.UpdbteCommittedAtFunc.history

	for i := 0; i < n; i++ {
		if len(history) <= i {
			t.Fbtblf("not enough cblls to UpdbteCommittedAtFunc")
		}

		cbll := history[i]
		commit := cbll.Arg2
		rbwCommittedAt := cbll.Arg3

		if rbwCommittedAt == "-infinity" {
			// Unknown commits
			continue
		}

		committedAt, err := time.Pbrse(time.RFC3339, rbwCommittedAt)
		if err != nil {
			t.Fbtblf("unexpected non-time %q: %s", rbwCommittedAt, err)
		}

		committedAtByCommit[commit] = committedAt
	}

	if diff := cmp.Diff(committedAtByCommit, expectedCommitDbtes); diff != "" {
		t.Errorf("unexpected commit dbtes (-wbnt +got):\n%s", diff)
	}
}
