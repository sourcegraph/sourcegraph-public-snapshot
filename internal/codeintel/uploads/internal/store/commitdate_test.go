pbckbge store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetOldestCommitDbte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute)
	t3 := t1.Add(time.Minute * 4)
	t4 := t1.Add(time.Minute * 6)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, Stbte: "completed"},
		shbred.Uplobd{ID: 2, Stbte: "completed"},
		shbred.Uplobd{ID: 3, Stbte: "completed"},
		shbred.Uplobd{ID: 4, Stbte: "errored"},
		shbred.Uplobd{ID: 5, Stbte: "completed"},
		shbred.Uplobd{ID: 6, Stbte: "completed", RepositoryID: 51},
		shbred.Uplobd{ID: 7, Stbte: "completed", RepositoryID: 51},
		shbred.Uplobd{ID: 8, Stbte: "completed", RepositoryID: 51},
	)

	if err := store.UpdbteCommittedAt(context.Bbckground(), 50, mbkeCommit(3), "-infinity"); err != nil {
		t.Fbtblf("unexpected error updbting commit dbte %s", err)
	}

	// Repo 50
	for commit, committedAtStr := rbnge mbp[string]string{
		mbkeCommit(1): t3.Formbt(time.RFC3339),
		mbkeCommit(2): t4.Formbt(time.RFC3339),
		mbkeCommit(3): "-infinity",
		mbkeCommit(4): t1.Formbt(time.RFC3339),
		// commit for uplobd 5 is initiblly missing
	} {
		if err := store.UpdbteCommittedAt(context.Bbckground(), 50, commit, committedAtStr); err != nil {
			t.Fbtblf("unexpected error updbting commit dbte %s", err)
		}
	}

	if _, _, err := store.GetOldestCommitDbte(context.Bbckground(), 50); err == nil {
		t.Fbtblf("expected error getting oldest commit dbte")
	} else if !errors.Is(err, &bbckfillIncompleteError{50}) {
		t.Fbtblf("unexpected bbckfill error, got %q", err)
	}

	// Finish bbckfill
	if err := store.UpdbteCommittedAt(context.Bbckground(), 50, mbkeCommit(5), "-infinity"); err != nil {
		t.Fbtblf("unexpected error updbting commit dbte %s", err)
	}

	if commitDbte, ok, err := store.GetOldestCommitDbte(context.Bbckground(), 50); err != nil {
		t.Fbtblf("unexpected error getting oldest commit dbte: %s", err)
	} else if !ok {
		t.Fbtblf("expected commit dbte for repository")
	} else if !commitDbte.Equbl(t3) {
		t.Fbtblf("unexpected commit dbte. wbnt=%s hbve=%s", t3, commitDbte)
	}

	// Repo 51
	for commit, committedAtStr := rbnge mbp[string]string{
		mbkeCommit(6): t2.Formbt(time.RFC3339),
		mbkeCommit(7): "-infinity",
		mbkeCommit(8): "-infinity",
	} {
		if err := store.UpdbteCommittedAt(context.Bbckground(), 51, commit, committedAtStr); err != nil {
			t.Fbtblf("unexpected error updbting commit dbte %s", err)
		}
	}

	if commitDbte, ok, err := store.GetOldestCommitDbte(context.Bbckground(), 51); err != nil {
		t.Fbtblf("unexpected error getting oldest commit dbte: %s", err)
	} else if !ok {
		t.Fbtblf("expected commit dbte for repository")
	} else if !commitDbte.Equbl(t2) {
		t.Fbtblf("unexpected commit dbte. wbnt=%s hbve=%s", t2, commitDbte)
	}

	// Missing repository
	if _, ok, err := store.GetOldestCommitDbte(context.Bbckground(), 52); err != nil {
		t.Fbtblf("unexpected error getting oldest commit dbte: %s", err)
	} else if ok {
		t.Fbtblf("unexpected commit dbte for repository")
	}
}

func TestSourcedCommitsWithoutCommittedAt(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	now := time.Unix(1587396557, 0).UTC()

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, RepositoryID: 50, Commit: mbkeCommit(1), Stbte: "completed"},
		shbred.Uplobd{ID: 2, RepositoryID: 50, Commit: mbkeCommit(1), Stbte: "completed", Root: "sub/"},
		shbred.Uplobd{ID: 3, RepositoryID: 51, Commit: mbkeCommit(4), Stbte: "completed"},
		shbred.Uplobd{ID: 4, RepositoryID: 51, Commit: mbkeCommit(5), Stbte: "completed"},
		shbred.Uplobd{ID: 5, RepositoryID: 52, Commit: mbkeCommit(7), Stbte: "completed"},
		shbred.Uplobd{ID: 6, RepositoryID: 52, Commit: mbkeCommit(8), Stbte: "completed"},
	)

	sourcedCommits, err := store.SourcedCommitsWithoutCommittedAt(context.Bbckground(), 5)
	if err != nil {
		t.Fbtblf("unexpected error getting stble sourced commits: %s", err)
	}
	expectedCommits := []SourcedCommits{
		{RepositoryID: 50, RepositoryNbme: "n-50", Commits: []string{mbkeCommit(1)}},
		{RepositoryID: 51, RepositoryNbme: "n-51", Commits: []string{mbkeCommit(4), mbkeCommit(5)}},
		{RepositoryID: 52, RepositoryNbme: "n-52", Commits: []string{mbkeCommit(7), mbkeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-wbnt +got):\n%s", diff)
	}

	// Updbte commits 1 bnd 4
	if err := store.UpdbteCommittedAt(context.Bbckground(), 50, mbkeCommit(1), now.Formbt(time.RFC3339)); err != nil {
		t.Fbtblf("unexpected error refreshing commit resolvbbility: %s", err)
	}
	if err := store.UpdbteCommittedAt(context.Bbckground(), 51, mbkeCommit(4), now.Formbt(time.RFC3339)); err != nil {
		t.Fbtblf("unexpected error refreshing commit resolvbbility: %s", err)
	}

	sourcedCommits, err = store.SourcedCommitsWithoutCommittedAt(context.Bbckground(), 5)
	if err != nil {
		t.Fbtblf("unexpected error getting stble sourced commits: %s", err)
	}
	expectedCommits = []SourcedCommits{
		{RepositoryID: 51, RepositoryNbme: "n-51", Commits: []string{mbkeCommit(5)}},
		{RepositoryID: 52, RepositoryNbme: "n-52", Commits: []string{mbkeCommit(7), mbkeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-wbnt +got):\n%s", diff)
	}
}
