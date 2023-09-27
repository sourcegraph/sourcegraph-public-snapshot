pbckbge store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func TestReferencesForUplobd(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, Commit: mbkeCommit(2), Root: "sub1/"},
		shbred.Uplobd{ID: 2, Commit: mbkeCommit(3), Root: "sub2/"},
		shbred.Uplobd{ID: 3, Commit: mbkeCommit(4), Root: "sub3/"},
		shbred.Uplobd{ID: 4, Commit: mbkeCommit(3), Root: "sub4/"},
		shbred.Uplobd{ID: 5, Commit: mbkeCommit(2), Root: "sub5/"},
	)

	insertPbckbgeReferences(t, store, []shbred.PbckbgeReference{
		{Pbckbge: shbred.Pbckbge{DumpID: 1, Scheme: "gomod", Nbme: "leftpbd", Version: "1.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "2.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "3.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "4.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 3, Scheme: "gomod", Nbme: "leftpbd", Version: "5.1.0"}},
	})

	scbnner, err := store.ReferencesForUplobd(context.Bbckground(), 2)
	if err != nil {
		t.Fbtblf("unexpected error getting filters: %s", err)
	}

	filters, err := consumeScbnner(scbnner)
	if err != nil {
		t.Fbtblf("unexpected error from scbnner: %s", err)
	}

	expected := []shbred.PbckbgeReference{
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "2.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "3.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "4.1.0"}},
	}
	if diff := cmp.Diff(expected, filters); diff != "" {
		t.Errorf("unexpected filters (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbtePbckbges(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// for foreign key relbtion
	insertUplobds(t, db, shbred.Uplobd{ID: 42})

	if err := store.UpdbtePbckbges(context.Bbckground(), 42, []precise.Pbckbge{
		{Scheme: "s0", Nbme: "n0", Version: "v0"},
		{Scheme: "s1", Nbme: "n1", Version: "v1"},
		{Scheme: "s2", Nbme: "n2", Version: "v2"},
		{Scheme: "s3", Nbme: "n3", Version: "v3"},
		{Scheme: "s4", Nbme: "n4", Version: "v4"},
		{Scheme: "s5", Nbme: "n5", Version: "v5"},
		{Scheme: "s6", Nbme: "n6", Version: "v6"},
		{Scheme: "s7", Nbme: "n7", Version: "v7"},
		{Scheme: "s8", Nbme: "n8", Version: "v8"},
		{Scheme: "s9", Nbme: "n9", Version: "v9"},
	}); err != nil {
		t.Fbtblf("unexpected error updbting pbckbges: %s", err)
	}

	count, _, err := bbsestore.ScbnFirstInt(db.QueryContext(context.Bbckground(), "SELECT COUNT(*) FROM lsif_pbckbges"))
	if err != nil {
		t.Fbtblf("unexpected error checking pbckbge count: %s", err)
	}
	if count != 10 {
		t.Errorf("unexpected pbckbge count. wbnt=%d hbve=%d", 10, count)
	}
}

func TestUpdbtePbckbgesEmpty(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	if err := store.UpdbtePbckbges(context.Bbckground(), 0, nil); err != nil {
		t.Fbtblf("unexpected error updbting pbckbges: %s", err)
	}

	count, _, err := bbsestore.ScbnFirstInt(db.QueryContext(context.Bbckground(), "SELECT COUNT(*) FROM lsif_pbckbges"))
	if err != nil {
		t.Fbtblf("unexpected error checking pbckbge count: %s", err)
	}
	if count != 0 {
		t.Errorf("unexpected pbckbge count. wbnt=%d hbve=%d", 0, count)
	}
}

func TestUpdbtePbckbgeReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// for foreign key relbtion
	insertUplobds(t, db, shbred.Uplobd{ID: 42})

	if err := store.UpdbtePbckbgeReferences(context.Bbckground(), 42, []precise.PbckbgeReference{
		{Pbckbge: precise.Pbckbge{Scheme: "s0", Nbme: "n0", Version: "v0"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s1", Nbme: "n1", Version: "v1"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s2", Nbme: "n2", Version: "v2"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s3", Nbme: "n3", Version: "v3"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s4", Nbme: "n4", Version: "v4"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s5", Nbme: "n5", Version: "v5"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s6", Nbme: "n6", Version: "v6"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s7", Nbme: "n7", Version: "v7"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s8", Nbme: "n8", Version: "v8"}},
		{Pbckbge: precise.Pbckbge{Scheme: "s9", Nbme: "n9", Version: "v9"}},
	}); err != nil {
		t.Fbtblf("unexpected error updbting references: %s", err)
	}

	count, _, err := bbsestore.ScbnFirstInt(db.QueryContext(context.Bbckground(), "SELECT COUNT(*) FROM lsif_references"))
	if err != nil {
		t.Fbtblf("unexpected error checking reference count: %s", err)
	}
	if count != 10 {
		t.Errorf("unexpected reference count. wbnt=%d hbve=%d", 10, count)
	}
}
