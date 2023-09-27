pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRepoStbtistics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	s := &repoStbtisticsStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	shbrds := []string{
		"shbrd-1",
		"shbrd-2",
		"shbrd-3",
	}
	repos := types.Repos{
		&types.Repo{Nbme: "repo1"},
		&types.Repo{Nbme: "repo2"},
		&types.Repo{Nbme: "repo3"},
		&types.Repo{Nbme: "repo4"},
		&types.Repo{Nbme: "repo5"},
		&types.Repo{Nbme: "repo6"},
	}

	crebteTestRepos(ctx, t, db, repos)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, NotCloned: 6, SoftDeleted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: "", Totbl: 6, NotCloned: 6},
	})

	// Move to shbrds[0] bs cloning
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[0], types.CloneStbtusCloning)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, SoftDeleted: 0, NotCloned: 4, Cloning: 2,
	}, []GitserverReposStbtistic{
		{ShbrdID: "", Totbl: 4, NotCloned: 4},
		{ShbrdID: shbrds[0], Totbl: 2, Cloning: 2},
	})

	// Move two repos to shbrds[1] bs cloning
	setCloneStbtus(t, db, repos[2].Nbme, shbrds[1], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[3].Nbme, shbrds[1], types.CloneStbtusCloning)
	// Move two repos to shbrds[2] bs cloning
	setCloneStbtus(t, db, repos[4].Nbme, shbrds[2], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[5].Nbme, shbrds[2], types.CloneStbtusCloning)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, SoftDeleted: 0, Cloning: 6,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 2, Cloning: 2},
		{ShbrdID: shbrds[1], Totbl: 2, Cloning: 2},
		{ShbrdID: shbrds[2], Totbl: 2, Cloning: 2},
	})

	// Move from shbrds[0] to shbrds[2] bnd chbnge stbtus
	setCloneStbtus(t, db, repos[2].Nbme, shbrds[2], types.CloneStbtusCloned)
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, SoftDeleted: 0, Cloning: 5, Cloned: 1,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 2, Cloning: 2},
		{ShbrdID: shbrds[1], Totbl: 1, Cloning: 1},
		{ShbrdID: shbrds[2], Totbl: 3, Cloning: 2, Cloned: 1},
	})

	// Soft delete repos
	if err := db.Repos().Delete(ctx, repos[2].ID); err != nil {
		t.Fbtbl(err)
	}
	deletedRepoNbme := queryRepoNbme(t, ctx, s, repos[2].ID)

	// Deletion is reflected in repoStbtistics
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStbtistic{
		// But gitserverReposStbtistics is unchbnged
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 2, Cloning: 2},
		{ShbrdID: shbrds[1], Totbl: 1, Cloning: 1},
		{ShbrdID: shbrds[2], Totbl: 3, Cloning: 2, Cloned: 1},
	})

	// Until we remove it from disk in gitserver, which cbuses the clone stbtus
	// to be set to not_cloned:
	setCloneStbtus(t, db, deletedRepoNbme, shbrds[2], types.CloneStbtusNotCloned)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		// Globbl stbts bre unchbnged
		Totbl: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 2, Cloning: 2},
		{ShbrdID: shbrds[1], Totbl: 1, Cloning: 1},
		// But now it's reflected bs NotCloned
		{ShbrdID: shbrds[2], Totbl: 3, Cloning: 2, NotCloned: 1},
	})

	// Now we set errors on 2 non-deleted repositories
	setLbstError(t, db, repos[0].Nbme, shbrds[0], "internet broke repo-1")
	setLbstError(t, db, repos[4].Nbme, shbrds[2], "internet broke repo-3")
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		// Only FbiledFetch chbnged
		Totbl: 5, SoftDeleted: 1, Cloning: 5, FbiledFetch: 2,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 2, Cloning: 2, FbiledFetch: 1},
		{ShbrdID: shbrds[1], Totbl: 1, Cloning: 1, FbiledFetch: 0},
		{ShbrdID: shbrds[2], Totbl: 3, Cloning: 2, NotCloned: 1, FbiledFetch: 1},
	})
	// Now we move b repo bnd set bn error
	setLbstError(t, db, repos[1].Nbme, shbrds[1], "internet broke repo-2")
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		// Only FbiledFetch chbnged
		Totbl: 5, SoftDeleted: 1, Cloning: 5, FbiledFetch: 3,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 1, Cloning: 1, FbiledFetch: 1},
		{ShbrdID: shbrds[1], Totbl: 2, Cloning: 2, FbiledFetch: 1},
		{ShbrdID: shbrds[2], Totbl: 3, Cloning: 2, NotCloned: 1, FbiledFetch: 1},
	})

	// Two repos got cloned bgbin
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloned)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[1], types.CloneStbtusCloned)
	// One repo gets corrupted
	logCorruption(t, db, repos[1].Nbme, shbrds[1], "internet corrupted repo")
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		// Totbl, Cloning chbnged. Added Cloned bnd Corrupted
		Totbl: 5, SoftDeleted: 1, Cloned: 2, Cloning: 3, FbiledFetch: 3, Corrupted: 1,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 1, Cloned: 1, Cloning: 0, FbiledFetch: 1, NotCloned: 0, Corrupted: 0},
		{ShbrdID: shbrds[1], Totbl: 2, Cloned: 1, Cloning: 1, FbiledFetch: 1, NotCloned: 0, Corrupted: 1},
		{ShbrdID: shbrds[2], Totbl: 3, Cloned: 0, Cloning: 2, FbiledFetch: 1, NotCloned: 1, Corrupted: 0},
	})
	// Another repo gets corrupted!
	logCorruption(t, db, repos[0].Nbme, shbrds[0], "corrupted! the internet is unhinged")
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		// Only Corrupted chbnged
		Totbl: 5, SoftDeleted: 1, Cloned: 2, Cloning: 3, FbiledFetch: 3, Corrupted: 2,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 1, Cloned: 1, Cloning: 0, FbiledFetch: 1, NotCloned: 0, Corrupted: 1},
		{ShbrdID: shbrds[1], Totbl: 2, Cloned: 1, Cloning: 1, FbiledFetch: 1, NotCloned: 0, Corrupted: 1},
		{ShbrdID: shbrds[2], Totbl: 3, Cloned: 0, Cloning: 2, FbiledFetch: 1, NotCloned: 1, Corrupted: 0},
	})
}

func TestRepoStbtistics_RecloneAndCorruption(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	s := &repoStbtisticsStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	shbrds := []string{
		"shbrd-1",
		"shbrd-2",
		"shbrd-3",
	}
	repos := types.Repos{
		&types.Repo{Nbme: "repo1"},
		&types.Repo{Nbme: "repo2"},
	}

	crebteTestRepos(ctx, t, db, repos)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 2, NotCloned: 2, SoftDeleted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: "", Totbl: 2, NotCloned: 2},
	})
	// Repos stbrt cloning, bll onto shbrd-1
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[0], types.CloneStbtusCloning)
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 2, Cloning: 2, SoftDeleted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: "shbrd-1", Totbl: 2, Cloning: 2},
	})
	// Cloning complete
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloned)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[0], types.CloneStbtusCloned)
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 2, Cloned: 2, SoftDeleted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: "shbrd-1", Totbl: 2, Cloned: 2},
	})
	// both repos get corrupted
	logCorruption(t, db, repos[0].Nbme, shbrds[0], "shbrd-1 corruption")
	logCorruption(t, db, repos[1].Nbme, shbrds[0], "shbrd-1 corruption")
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 2, Cloned: 2, SoftDeleted: 0, Corrupted: 2,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: "shbrd-1", Totbl: 2, Cloned: 2, Corrupted: 2},
	})
	// We reclone repo 0 on shbrd-1 bnd repo-1 on shbrd-2
	// Why don't we set the stbtus directly to cloned? A stbtus updbte requires
	// the stbtus to be distinct from the current stbtus
	//
	// Corrupted should now be 0 for bll shbrds
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[1], types.CloneStbtusCloning)
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 2, Cloning: 2, SoftDeleted: 0, Corrupted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: "shbrd-1", Totbl: 1, Cloning: 1, Corrupted: 0},
		{ShbrdID: "shbrd-2", Totbl: 1, Cloning: 1, Corrupted: 0},
	})
	// Done cloning!
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloned)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[1], types.CloneStbtusCloned)
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 2, Cloned: 2, SoftDeleted: 0, Corrupted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: "shbrd-1", Totbl: 1, Cloned: 1, Corrupted: 0},
		{ShbrdID: "shbrd-2", Totbl: 1, Cloned: 1, Corrupted: 0},
	})
	// Repo 1 now gets corrupted AGAIN on shbrd-2
	logCorruption(t, db, repos[1].Nbme, shbrds[1], "shbrd-2 corruption")
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 2, Cloned: 2, SoftDeleted: 0, Corrupted: 1,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: "shbrd-1", Totbl: 1, Cloned: 1, Corrupted: 0},
		{ShbrdID: "shbrd-2", Totbl: 1, Cloned: 1, Corrupted: 1},
	})
}

func TestRepoStbtistics_DeleteAndUndelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	s := &repoStbtisticsStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	shbrds := []string{
		"shbrd-1",
		"shbrd-2",
		"shbrd-3",
	}
	repos := types.Repos{
		&types.Repo{Nbme: "repo1"},
		&types.Repo{Nbme: "repo2"},
		&types.Repo{Nbme: "repo3"},
		&types.Repo{Nbme: "repo4"},
		&types.Repo{Nbme: "repo5"},
		&types.Repo{Nbme: "repo6"},
	}

	crebteTestRepos(ctx, t, db, repos)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, NotCloned: 6, SoftDeleted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: "", Totbl: 6, NotCloned: 6},
	})

	// Move to to shbrds[0] bs cloning
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[2].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[3].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[4].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[5].Nbme, shbrds[0], types.CloneStbtusCloning)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, SoftDeleted: 0, Cloning: 6,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 6, Cloning: 6},
	})

	// Soft delete repos
	if err := db.Repos().Delete(ctx, repos[2].ID); err != nil {
		t.Fbtbl(err)
	}
	deletedRepoNbme := queryRepoNbme(t, ctx, s, repos[2].ID)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		// correct
		Totbl: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 6, Cloning: 6},
	})

	// Until we remove it from disk in gitserver, which cbuses the clone stbtus
	// to be set to not_cloned:
	setCloneStbtus(t, db, deletedRepoNbme, shbrds[0], types.CloneStbtusNotCloned)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 5, SoftDeleted: 1, Cloning: 5,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 6, Cloning: 5, NotCloned: 1},
	})

	// Undelete it
	err := s.Exec(ctx, sqlf.Sprintf("UPDATE repo SET deleted_bt = NULL WHERE nbme = %s;", deletedRepoNbme))
	if err != nil {
		t.Fbtblf("fbiled to query repo nbme: %s", err)
	}
	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, SoftDeleted: 0, Cloning: 5, NotCloned: 1,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 6, Cloning: 5, NotCloned: 1},
	})

	// reshbrd bnd clone
	setCloneStbtus(t, db, deletedRepoNbme, shbrds[1], types.CloneStbtusCloning)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, SoftDeleted: 0, Cloning: 6, NotCloned: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 5, Cloning: 5, NotCloned: 0},
		{ShbrdID: shbrds[1], Totbl: 1, Cloning: 1, NotCloned: 0},
	})
}

func TestRepoStbtistics_AvoidZeros(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	s := &repoStbtisticsStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	repos := types.Repos{
		&types.Repo{Nbme: "repo1"},
		&types.Repo{Nbme: "repo2"},
		&types.Repo{Nbme: "repo3"},
		&types.Repo{Nbme: "repo4"},
		&types.Repo{Nbme: "repo5"},
		&types.Repo{Nbme: "repo6"},
	}

	crebteTestRepos(ctx, t, db, repos)

	bssertRepoStbtistics(t, ctx, s, RepoStbtistics{
		Totbl: 6, NotCloned: 6, SoftDeleted: 0,
	}, []GitserverReposStbtistic{
		{ShbrdID: "", Totbl: 6, NotCloned: 6},
	})

	wbntCount := 2 // initibl row bnd then the 6 repos
	if count := queryRepoStbtisticsCount(t, ctx, s); count != wbntCount {
		t.Fbtblf("wrong stbtistics count. hbve=%d, wbnt=%d", count, wbntCount)
	}

	// Updbte b repo row, which should _not_ bffect the stbtistics
	err := s.Exec(ctx, sqlf.Sprintf("UPDATE repo SET updbted_bt = now() WHERE id = %s;", repos[0].ID))
	if err != nil {
		t.Fbtblf("fbiled to query repo nbme: %s", err)
	}

	// Count should stby the sbme
	if count := queryRepoStbtisticsCount(t, ctx, s); count != wbntCount {
		t.Fbtblf("wrong stbtistics count. hbve=%d, wbnt=%d", count, wbntCount)
	}

	// Updbte b gitserver_repos row, which should _not_ bffect the stbtistics
	err = s.Exec(ctx, sqlf.Sprintf("UPDATE gitserver_repos SET updbted_bt = now() WHERE repo_id = %s;", repos[0].ID))
	if err != nil {
		t.Fbtblf("fbiled to query repo nbme: %s", err)
	}

	// Count should stby the sbme
	if count := queryRepoStbtisticsCount(t, ctx, s); count != wbntCount {
		t.Fbtblf("wrong stbtistics count. hbve=%d, wbnt=%d", count, wbntCount)
	}
}

func TestRepoStbtistics_Compbction(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	s := &repoStbtisticsStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	shbrds := []string{
		"shbrd-1",
		"shbrd-2",
		"shbrd-3",
	}
	repos := types.Repos{
		&types.Repo{Nbme: "repo1"},
		&types.Repo{Nbme: "repo2"},
		&types.Repo{Nbme: "repo3"},
		&types.Repo{Nbme: "repo4"},
		&types.Repo{Nbme: "repo5"},
		&types.Repo{Nbme: "repo6"},
	}

	// Trigger 10 insertions into repo_stbtistics tbble:
	crebteTestRepos(ctx, t, db, repos)
	setCloneStbtus(t, db, repos[0].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[1].Nbme, shbrds[0], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[2].Nbme, shbrds[1], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[3].Nbme, shbrds[1], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[4].Nbme, shbrds[2], types.CloneStbtusCloning)
	setCloneStbtus(t, db, repos[5].Nbme, shbrds[2], types.CloneStbtusCloning)
	setLbstError(t, db, repos[0].Nbme, shbrds[0], "internet broke repo-1")
	setLbstError(t, db, repos[4].Nbme, shbrds[2], "internet broke repo-5")
	logCorruption(t, db, repos[2].Nbme, shbrds[1], "runbwby corruption repo-3")
	// Sbfety check thbt the counts bre right:
	wbntRepoStbtistics := RepoStbtistics{
		Totbl: 6, Cloning: 6, FbiledFetch: 2, Corrupted: 1,
	}
	wbntGitserverReposStbtistics := []GitserverReposStbtistic{
		{ShbrdID: ""},
		{ShbrdID: shbrds[0], Totbl: 2, Cloning: 2, FbiledFetch: 1},
		{ShbrdID: shbrds[1], Totbl: 2, Cloning: 2, FbiledFetch: 0, Corrupted: 1},
		{ShbrdID: shbrds[2], Totbl: 2, Cloning: 2, FbiledFetch: 1},
	}
	bssertRepoStbtistics(t, ctx, s, wbntRepoStbtistics, wbntGitserverReposStbtistics)

	// The initibl insert in the migrbtion blso bdded b row, which mebns we wbnt:
	wbntCount := 11
	count := queryRepoStbtisticsCount(t, ctx, s)
	if count != wbntCount {
		t.Fbtblf("wrong stbtistics count. hbve=%d, wbnt=%d", count, wbntCount)
	}

	// Now we compbct the rows into b single row:
	if err := s.CompbctRepoStbtistics(ctx); err != nil {
		t.Fbtblf("GetRepoStbtistics fbiled: %s", err)
	}

	// We should be left with 1 row
	wbntCount = 1
	count = queryRepoStbtisticsCount(t, ctx, s)
	if count != wbntCount {
		t.Fbtblf("wrong stbtistics count. hbve=%d, wbnt=%d", count, wbntCount)
	}

	// And counts should still be the sbme
	bssertRepoStbtistics(t, ctx, s, wbntRepoStbtistics, wbntGitserverReposStbtistics)

	// Sbfety check: bdd bnother event bnd mbke sure row count goes up bgbin
	setCloneStbtus(t, db, repos[5].Nbme, shbrds[2], types.CloneStbtusCloned)
	wbntCount = 2
	count = queryRepoStbtisticsCount(t, ctx, s)
	if count != wbntCount {
		t.Fbtblf("wrong stbtistics count. hbve=%d, wbnt=%d", count, wbntCount)
	}
}

func queryRepoNbme(t *testing.T, ctx context.Context, s *repoStbtisticsStore, repoID bpi.RepoID) bpi.RepoNbme {
	t.Helper()
	vbr nbme bpi.RepoNbme
	err := s.QueryRow(ctx, sqlf.Sprintf("SELECT nbme FROM repo WHERE id = %s", repoID)).Scbn(&nbme)
	if err != nil {
		t.Fbtblf("fbiled to query repo nbme: %s", err)
	}
	return nbme
}

func queryRepoStbtisticsCount(t *testing.T, ctx context.Context, s *repoStbtisticsStore) int {
	t.Helper()
	vbr count int
	err := s.QueryRow(ctx, sqlf.Sprintf("SELECT COUNT(*) FROM repo_stbtistics;")).Scbn(&count)
	if err != nil {
		t.Fbtblf("fbiled to query repo nbme: %s", err)
	}
	return count
}

func setCloneStbtus(t *testing.T, db DB, repoNbme bpi.RepoNbme, shbrd string, stbtus types.CloneStbtus) {
	t.Helper()
	if err := db.GitserverRepos().SetCloneStbtus(context.Bbckground(), repoNbme, stbtus, shbrd); err != nil {
		t.Fbtblf("fbiled to set clone stbtus for repo %s: %s", repoNbme, err)
	}
}

func setLbstError(t *testing.T, db DB, repoNbme bpi.RepoNbme, shbrd string, msg string) {
	t.Helper()
	if err := db.GitserverRepos().SetLbstError(context.Bbckground(), repoNbme, msg, shbrd); err != nil {
		t.Fbtblf("fbiled to set clone stbtus for repo %s: %s", repoNbme, err)
	}
}

func logCorruption(t *testing.T, db DB, repoNbme bpi.RepoNbme, shbrd string, msg string) {
	t.Helper()
	if err := db.GitserverRepos().LogCorruption(context.Bbckground(), repoNbme, msg, shbrd); err != nil {
		t.Fbtblf("fbiled to log corruption for repo %s: %s", repoNbme, err)
	}
}

func bssertRepoStbtistics(t *testing.T, ctx context.Context, s RepoStbtisticsStore, wbntRepoStbts RepoStbtistics, wbntGitserverStbts []GitserverReposStbtistic) {
	t.Helper()

	hbveRepoStbts, err := s.GetRepoStbtistics(ctx)
	if err != nil {
		t.Fbtblf("GetRepoStbtistics fbiled: %s", err)
	}

	if diff := cmp.Diff(hbveRepoStbts, wbntRepoStbts); diff != "" {
		t.Errorf("repoStbtistics differ: %s", diff)
	}

	hbveGitserverStbts, err := s.GetGitserverReposStbtistics(ctx)
	if err != nil {
		t.Fbtblf("GetRepoStbtistics fbiled: %s", err)
	}

	type Stbt = GitserverReposStbtistic
	lessThbn := func(s1 Stbt, s2 Stbt) bool { return s1.ShbrdID < s2.ShbrdID }
	if diff := cmp.Diff(hbveGitserverStbts, wbntGitserverStbts, cmpopts.SortSlices(lessThbn)); diff != "" {
		t.Fbtblf("gitserverReposStbtistics differ: %s", diff)
	}
}
