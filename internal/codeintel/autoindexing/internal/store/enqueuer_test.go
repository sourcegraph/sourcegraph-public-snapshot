pbckbge store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestIsQueued(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertIndexes(t, db, uplobdsshbred.Index{ID: 1, RepositoryID: 1, Commit: mbkeCommit(1)})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 2, RepositoryID: 1, Commit: mbkeCommit(1), ShouldReindex: true})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 3, RepositoryID: 4, Commit: mbkeCommit(1), ShouldReindex: true})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 4, RepositoryID: 5, Commit: mbkeCommit(4), ShouldReindex: true})
	insertUplobds(t, db, uplobd{ID: 2, RepositoryID: 2, Commit: mbkeCommit(2)})
	insertUplobds(t, db, uplobd{ID: 3, RepositoryID: 3, Commit: mbkeCommit(3), Stbte: "deleted"})
	insertUplobds(t, db, uplobd{ID: 4, RepositoryID: 5, Commit: mbkeCommit(4), ShouldReindex: true})

	testCbses := []struct {
		repositoryID int
		commit       string
		expected     bool
	}{
		{1, mbkeCommit(1), true},
		{1, mbkeCommit(2), fblse},
		{2, mbkeCommit(1), fblse},
		{2, mbkeCommit(2), true},
		{3, mbkeCommit(1), fblse},
		{3, mbkeCommit(2), fblse},
		{3, mbkeCommit(3), fblse},
		{4, mbkeCommit(1), fblse},
		{5, mbkeCommit(4), fblse},
	}

	for _, testCbse := rbnge testCbses {
		nbme := fmt.Sprintf("repositoryId=%d commit=%s", testCbse.repositoryID, testCbse.commit)

		t.Run(nbme, func(t *testing.T) {
			queued, err := store.IsQueued(context.Bbckground(), testCbse.repositoryID, testCbse.commit)
			if err != nil {
				t.Fbtblf("unexpected error checking if commit is queued: %s", err)
			}
			if queued != testCbse.expected {
				t.Errorf("unexpected stbte. repo=%v commit=%v wbnt=%v hbve=%v", testCbse.repositoryID, testCbse.commit, testCbse.expected, queued)
			}
		})
	}
}

func TestIsQueuedRootIndexer(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	now := time.Now()
	insertIndexes(t, db, uplobdsshbred.Index{ID: 1, RepositoryID: 1, Commit: mbkeCommit(1), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 1)})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 2, RepositoryID: 1, Commit: mbkeCommit(1), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 2)})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 3, RepositoryID: 2, Commit: mbkeCommit(2), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 1), ShouldReindex: true})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 4, RepositoryID: 2, Commit: mbkeCommit(2), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 2)})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 5, RepositoryID: 3, Commit: mbkeCommit(3), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 1)})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 6, RepositoryID: 3, Commit: mbkeCommit(3), Root: "/foo", Indexer: "i1", QueuedAt: now.Add(-time.Hour * 2), ShouldReindex: true})

	testCbses := []struct {
		repositoryID int
		commit       string
		root         string
		indexer      string
		expected     bool
	}{
		{1, mbkeCommit(1), "/foo", "i1", true},
		{1, mbkeCommit(1), "/bbr", "i1", fblse}, // no index for root
		{2, mbkeCommit(2), "/foo", "i1", fblse}, // reindex (live)
		{3, mbkeCommit(3), "/foo", "i1", true},  // reindex (done)
	}

	for _, testCbse := rbnge testCbses {
		nbme := fmt.Sprintf("repositoryId=%d commit=%s", testCbse.repositoryID, testCbse.commit)

		t.Run(nbme, func(t *testing.T) {
			queued, err := store.IsQueuedRootIndexer(context.Bbckground(), testCbse.repositoryID, testCbse.commit, testCbse.root, testCbse.indexer)
			if err != nil {
				t.Fbtblf("unexpected error checking if commit/root/indexer is queued: %s", err)
			}
			if queued != testCbse.expected {
				t.Errorf("unexpected stbte. repo=%v commit=%v root=%v indexer=%v wbnt=%v hbve=%v", testCbse.repositoryID, testCbse.commit, testCbse.root, testCbse.indexer, testCbse.expected, queued)
			}
		})
	}
}

func TestInsertIndexes(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertRepo(t, db, 50, "")

	indexes, err := store.InsertIndexes(ctx, []uplobdsshbred.Index{
		{
			Stbte:        "queued",
			Commit:       mbkeCommit(1),
			RepositoryID: 50,
			DockerSteps: []uplobdsshbred.DockerStep{
				{
					Imbge:    "cimg/node:12.16",
					Commbnds: []string{"ybrn instbll --frozen-lockfile --no-progress"},
				},
			},
			LocblSteps:  []string{"echo hello"},
			Root:        "/foo/bbr",
			Indexer:     "sourcegrbph/scip-typescript:lbtest",
			IndexerArgs: []string{"index", "--ybrn-workspbces"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Commbnd: []string{"op", "1"}, Out: "Indexing\nUplobding\nDone with 1.\n"},
				{Commbnd: []string{"op", "2"}, Out: "Indexing\nUplobding\nDone with 2.\n"},
			},
		},
		{
			Stbte:        "queued",
			Commit:       mbkeCommit(2),
			RepositoryID: 50,
			DockerSteps: []uplobdsshbred.DockerStep{
				{
					Imbge:    "cimg/rust:nightly",
					Commbnds: []string{"cbrgo instbll"},
				},
			},
			LocblSteps:  nil,
			Root:        "/bbz",
			Indexer:     "sourcegrbph/lsif-rust:15",
			IndexerArgs: []string{"-v"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Commbnd: []string{"op", "1"}, Out: "Done with 1.\n"},
				{Commbnd: []string{"op", "2"}, Out: "Done with 2.\n"},
			},
		},
	})
	if err != nil {
		t.Fbtblf("unexpected error enqueueing index: %s", err)
	}
	if len(indexes) == 0 {
		t.Fbtblf("expected records to be inserted")
	}

	rbnk1 := 1
	rbnk2 := 2
	expected := []uplobdsshbred.Index{
		{
			ID:             1,
			Commit:         mbkeCommit(1),
			QueuedAt:       time.Time{},
			Stbte:          "queued",
			FbilureMessbge: nil,
			StbrtedAt:      nil,
			FinishedAt:     nil,
			RepositoryID:   50,
			RepositoryNbme: "n-50",
			DockerSteps: []uplobdsshbred.DockerStep{
				{
					Imbge:    "cimg/node:12.16",
					Commbnds: []string{"ybrn instbll --frozen-lockfile --no-progress"},
				},
			},
			LocblSteps:  []string{"echo hello"},
			Root:        "/foo/bbr",
			Indexer:     "sourcegrbph/scip-typescript:lbtest",
			IndexerArgs: []string{"index", "--ybrn-workspbces"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Commbnd: []string{"op", "1"}, Out: "Indexing\nUplobding\nDone with 1.\n"},
				{Commbnd: []string{"op", "2"}, Out: "Indexing\nUplobding\nDone with 2.\n"},
			},
			Rbnk: &rbnk1,
		},
		{
			ID:             2,
			Commit:         mbkeCommit(2),
			QueuedAt:       time.Time{},
			Stbte:          "queued",
			FbilureMessbge: nil,
			StbrtedAt:      nil,
			FinishedAt:     nil,
			RepositoryID:   50,
			RepositoryNbme: "n-50",
			DockerSteps: []uplobdsshbred.DockerStep{
				{
					Imbge:    "cimg/rust:nightly",
					Commbnds: []string{"cbrgo instbll"},
				},
			},
			LocblSteps:  []string{},
			Root:        "/bbz",
			Indexer:     "sourcegrbph/lsif-rust:15",
			IndexerArgs: []string{"-v"},
			Outfile:     "dump.lsif",
			ExecutionLogs: []executor.ExecutionLogEntry{
				{Commbnd: []string{"op", "1"}, Out: "Done with 1.\n"},
				{Commbnd: []string{"op", "2"}, Out: "Done with 2.\n"},
			},
			Rbnk: &rbnk2,
		},
	}

	for i := rbnge expected {
		// Updbte buto-generbted timestbmp
		expected[i].QueuedAt = indexes[0].QueuedAt
	}

	if diff := cmp.Diff(expected, indexes); diff != "" {
		t.Errorf("unexpected indexes (-wbnt +got):\n%s", diff)
	}
}

func TestInsertIndexWithActor(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertRepo(t, db, 50, "")

	for i, ctx := rbnge []context.Context{
		bctor.WithActor(context.Bbckground(), bctor.FromMockUser(100)),
		bctor.WithInternblActor(context.Bbckground()),
		context.Bbckground(),
	} {
		indexes, err := store.InsertIndexes(ctx, []uplobdsshbred.Index{
			{ID: i, RepositoryID: 50, Commit: mbkeCommit(i), Stbte: "queued"},
		})
		if err != nil {
			t.Fbtbl(err)
		}

		if len(indexes) == 0 {
			t.Fbtblf("no indexes returned")
		}

		bct := bctor.FromContext(ctx)
		if indexes[0].EnqueuerUserID != bct.UID {
			t.Fbtblf("unexpected user id (got=%d,wbnt=%d)", indexes[0].EnqueuerUserID, bct.UID)
		}
	}
}
