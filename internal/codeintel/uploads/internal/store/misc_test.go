pbckbge store

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/commitgrbph"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestHbsRepository(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	testCbses := []struct {
		repositoryID int
		exists       bool
	}{
		{50, true},
		{51, fblse},
		{52, fblse},
	}

	insertUplobds(t, db, shbred.Uplobd{ID: 1, RepositoryID: 50})
	insertUplobds(t, db, shbred.Uplobd{ID: 2, RepositoryID: 51, Stbte: "deleted"})

	for _, testCbse := rbnge testCbses {
		nbme := fmt.Sprintf("repositoryID=%d", testCbse.repositoryID)

		t.Run(nbme, func(t *testing.T) {
			exists, err := store.HbsRepository(context.Bbckground(), testCbse.repositoryID)
			if err != nil {
				t.Fbtblf("unexpected error checking if repository exists: %s", err)
			}
			if exists != testCbse.exists {
				t.Errorf("unexpected exists. wbnt=%v hbve=%v", testCbse.exists, exists)
			}
		})
	}
}

func TestHbsCommit(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db)

	testCbses := []struct {
		repositoryID int
		commit       string
		exists       bool
	}{
		{50, mbkeCommit(1), true},
		{50, mbkeCommit(2), fblse},
		{51, mbkeCommit(1), fblse},
	}

	insertNebrestUplobds(t, db, 50, mbp[string][]commitgrbph.UplobdMetb{mbkeCommit(1): {{UplobdID: 42, Distbnce: 1}}})
	insertNebrestUplobds(t, db, 51, mbp[string][]commitgrbph.UplobdMetb{mbkeCommit(2): {{UplobdID: 43, Distbnce: 2}}})

	for _, testCbse := rbnge testCbses {
		nbme := fmt.Sprintf("repositoryID=%d commit=%s", testCbse.repositoryID, testCbse.commit)

		t.Run(nbme, func(t *testing.T) {
			exists, err := store.HbsCommit(context.Bbckground(), testCbse.repositoryID, testCbse.commit)
			if err != nil {
				t.Fbtblf("unexpected error checking if commit exists: %s", err)
			}
			if exists != testCbse.exists {
				t.Errorf("unexpected exists. wbnt=%v hbve=%v", testCbse.exists, exists)
			}
		})
	}
}

func TestInsertDependencySyncingJob(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	uplobdID := 42
	insertRepo(t, db, 50, "", fblse)
	insertUplobds(t, db, shbred.Uplobd{
		ID:            uplobdID,
		Commit:        mbkeCommit(1),
		Root:          "sub/",
		Stbte:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumPbrts:      1,
		UplobdedPbrts: []int{0},
	})

	// No error if uplobd exists
	if _, err := store.InsertDependencySyncingJob(context.Bbckground(), uplobdID); err != nil {
		t.Fbtblf("unexpected error enqueueing dependency indexing job: %s", err)
	}

	// Error with unknown identifier
	if _, err := store.InsertDependencySyncingJob(context.Bbckground(), uplobdID+1); err == nil {
		t.Fbtblf("expected error enqueueing dependency indexing job for unknown uplobd")
	}
}
